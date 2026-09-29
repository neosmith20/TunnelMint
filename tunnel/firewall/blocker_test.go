/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 WireHush contributors. All Rights Reserved.
 */

package firewall

import (
	"errors"
	"reflect"
	"testing"
)

func TestReplaceSessionMakeBeforeBreak(t *testing.T) {
	events := []string{}
	got, err := replaceSession(7, func() (uintptr, error) {
		events = append(events, "install")
		return 9, nil
	}, func(session uintptr) {
		events = append(events, "close:"+string(rune(session)))
	})
	if err != nil || got != 9 {
		t.Fatalf("replacement = %d, %v", got, err)
	}
	if want := []string{"install", "close:\x07"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestReplaceSessionFailureKeepsOld(t *testing.T) {
	closed := false
	wantErr := errors.New("injected install failure")
	got, err := replaceSession(7, func() (uintptr, error) { return 0, wantErr }, func(uintptr) { closed = true })
	if !errors.Is(err, wantErr) || got != 7 {
		t.Fatalf("replacement = %d, %v", got, err)
	}
	if closed {
		t.Fatal("old firewall session was closed after replacement failure")
	}
}
