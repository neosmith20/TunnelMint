/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 WireHush contributors. All Rights Reserved.
 */

package bootstrap

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSettingsOrderingAndHandoff(t *testing.T) {
	settings := DefaultSettings()
	if err := settings.Move(0, 2); err != nil {
		t.Fatal(err)
	}
	if err := settings.Toggle(1); err != nil {
		t.Fatal(err)
	}
	resolvers, err := settings.EnabledResolvers()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"8.8.8.8", "1.1.1.1", "8.8.4.4", "4.2.2.1", "4.2.2.2",
		"2606:4700:4700::1111", "2606:4700:4700::1001", "2001:4860:4860::8888",
		"2001:4860:4860::8844", "2620:fe::11", "2620:fe::fe:11",
	}
	var got []string
	for _, resolver := range resolvers {
		got = append(got, resolver.String())
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolver order = %v, want %v", got, want)
	}
}

func TestSettingsValidationAndCustomEntries(t *testing.T) {
	settings := DefaultSettings()
	if err := settings.AddCustom("2001:db8::53"); err != nil {
		t.Fatal(err)
	}
	if err := settings.AddCustom("resolver.example"); err == nil {
		t.Fatal("hostname unexpectedly accepted")
	}
	if err := settings.AddCustom("239.1.1.1"); err == nil {
		t.Fatal("multicast resolver unexpectedly accepted")
	}
	if err := settings.RemoveCustom(len(settings.Entries) - 1); err != nil {
		t.Fatal(err)
	}
	if err := settings.RemoveCustom(0); err == nil {
		t.Fatal("built-in resolver unexpectedly removed")
	}
	for i := range settings.Entries {
		settings.Entries[i].Enabled = false
	}
	if _, err := settings.EnabledResolvers(); err == nil {
		t.Fatal("empty enabled resolver set unexpectedly accepted")
	}
}

func TestSettingsPersistenceAndRestoreDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bootstrap-dns.json")
	settings := DefaultSettings()
	if err := settings.AddCustom("192.0.2.53"); err != nil {
		t.Fatal(err)
	}
	if err := settings.Move(len(settings.Entries)-1, -len(settings.Entries)+1); err != nil {
		t.Fatal(err)
	}
	if err := settings.Save(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(settings, loaded) {
		t.Fatalf("loaded settings %#v, want %#v", loaded, settings)
	}
	settings.RestoreDefaults()
	if !reflect.DeepEqual(settings, DefaultSettings()) {
		t.Fatal("restore defaults did not reset settings")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
