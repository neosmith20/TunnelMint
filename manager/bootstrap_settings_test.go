/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 WireHush contributors. All Rights Reserved.
 */

package manager

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"

	"golang.zx2c4.com/wireguard/windows/bootstrap"
)

func TestBootstrapSettingsIPCSerialization(t *testing.T) {
	want := bootstrap.DefaultSettings()
	if err := want.AddCustom("192.0.2.53"); err != nil {
		t.Fatal(err)
	}
	var wire bytes.Buffer
	if err := gob.NewEncoder(&wire).Encode(want); err != nil {
		t.Fatal(err)
	}
	var got bootstrap.Settings
	if err := gob.NewDecoder(&wire).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("decoded settings %#v, want %#v", got, want)
	}
	if err := got.Validate(); err != nil {
		t.Fatal(err)
	}
}
