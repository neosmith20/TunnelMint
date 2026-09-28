/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package bootstrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
)

// Entry is one ordered bootstrap resolver choice. Built-in entries are marked
// as such so the settings UI can prevent removing them while still allowing
// them to be disabled.
type Entry struct {
	Address netip.Addr `json:"-"`
	Enabled bool       `json:"enabled"`
	Custom  bool       `json:"custom"`
}

type entryJSON struct {
	Address string `json:"address"`
	Enabled bool   `json:"enabled"`
	Custom  bool   `json:"custom"`
}

// Settings is the persisted, ordered bootstrap resolver list.
type Settings struct {
	Entries []Entry
}

// DefaultSettings returns all built-in resolvers enabled in their documented
// failover order.
func DefaultSettings() Settings {
	addresses := DefaultResolvers()
	entries := make([]Entry, len(addresses))
	for i, address := range addresses {
		entries[i] = Entry{Address: address, Enabled: true}
	}
	return Settings{Entries: entries}
}

// ParseAddress accepts only IP literals. Hostnames and scoped IPv6 literals
// are rejected because they would recurse through the resolver being bootstrapped.
func ParseAddress(value string) (netip.Addr, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return netip.Addr{}, errors.New("bootstrap resolver address is empty")
	}
	address, err := netip.ParseAddr(value)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("bootstrap resolver must be an IP address: %q", value)
	}
	if address.IsUnspecified() {
		return netip.Addr{}, fmt.Errorf("bootstrap resolver cannot be unspecified: %s", address)
	}
	if address.IsMulticast() {
		return netip.Addr{}, fmt.Errorf("bootstrap resolver cannot be multicast: %s", address)
	}
	return address, nil
}

// EnabledResolvers hands the ordered enabled list to the bootstrap engine.
func (s Settings) EnabledResolvers() ([]netip.Addr, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	resolvers := make([]netip.Addr, 0, len(s.Entries))
	for _, entry := range s.Entries {
		if entry.Enabled {
			resolvers = append(resolvers, entry.Address)
		}
	}
	return resolvers, nil
}

// Validate checks persisted values and ensures encrypted DNS always has a
// usable bootstrap path.
func (s Settings) Validate() error {
	if len(s.Entries) == 0 {
		return errors.New("at least one bootstrap resolver is required")
	}
	seen := make(map[netip.Addr]struct{}, len(s.Entries))
	enabled := 0
	for _, entry := range s.Entries {
		address, err := ParseAddress(entry.Address.String())
		if err != nil {
			return err
		}
		if _, ok := seen[address]; ok {
			return fmt.Errorf("duplicate bootstrap resolver: %s", address)
		}
		seen[address] = struct{}{}
		if entry.Enabled {
			enabled++
		}
	}
	if enabled == 0 {
		return errors.New("at least one bootstrap resolver must be enabled")
	}
	return nil
}

// AddCustom validates and appends an enabled custom resolver.
func (s *Settings) AddCustom(value string) error {
	address, err := ParseAddress(value)
	if err != nil {
		return err
	}
	for _, entry := range s.Entries {
		if entry.Address == address {
			return fmt.Errorf("bootstrap resolver already exists: %s", address)
		}
	}
	s.Entries = append(s.Entries, Entry{Address: address, Enabled: true, Custom: true})
	return nil
}

// RemoveCustom removes a custom entry. Built-in entries can be disabled but
// remain available for RestoreDefaults.
func (s *Settings) RemoveCustom(index int) error {
	if index < 0 || index >= len(s.Entries) || !s.Entries[index].Custom {
		return errors.New("only custom bootstrap resolvers can be removed")
	}
	s.Entries = append(s.Entries[:index], s.Entries[index+1:]...)
	return nil
}

func (s *Settings) Toggle(index int) error {
	if index < 0 || index >= len(s.Entries) {
		return errors.New("invalid bootstrap resolver selection")
	}
	s.Entries[index].Enabled = !s.Entries[index].Enabled
	if _, err := s.EnabledResolvers(); err != nil {
		s.Entries[index].Enabled = !s.Entries[index].Enabled
		return err
	}
	return nil
}

func (s *Settings) Move(index, delta int) error {
	target := index + delta
	if index < 0 || index >= len(s.Entries) || target < 0 || target >= len(s.Entries) {
		return errors.New("bootstrap resolver cannot be moved further")
	}
	s.Entries[index], s.Entries[target] = s.Entries[target], s.Entries[index]
	return nil
}

func (s *Settings) RestoreDefaults() {
	*s = DefaultSettings()
}

// Load reads settings from a TunnelMint-owned path. Missing settings use the
// built-in defaults without creating a file until the user changes them.
func Load(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return DefaultSettings(), nil
	}
	if err != nil {
		return Settings{}, err
	}
	var encoded []entryJSON
	if err := json.Unmarshal(data, &encoded); err != nil {
		return Settings{}, fmt.Errorf("decode bootstrap settings: %w", err)
	}
	settings := Settings{Entries: make([]Entry, len(encoded))}
	for i, entry := range encoded {
		address, err := ParseAddress(entry.Address)
		if err != nil {
			return Settings{}, err
		}
		settings.Entries[i] = Entry{Address: address, Enabled: entry.Enabled, Custom: entry.Custom}
	}
	if err := settings.Validate(); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

// Save writes settings atomically with a private file mode. The parent path is
// created by the TunnelMint data-root setup before this function is called.
func (s Settings) Save(path string) error {
	if err := s.Validate(); err != nil {
		return err
	}
	encoded := make([]entryJSON, len(s.Entries))
	for i, entry := range s.Entries {
		encoded[i] = entryJSON{Address: entry.Address.String(), Enabled: entry.Enabled, Custom: entry.Custom}
	}
	data, err := json.MarshalIndent(encoded, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".bootstrap-dns-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
