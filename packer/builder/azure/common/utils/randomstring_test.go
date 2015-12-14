// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package utils

import (
	"testing"
)

func TestRandomPassword_generates_15char_passwords(t *testing.T) {
	for i := 0; i < 100; i++ {
		pw := RandomPassword()
		t.Logf("pw: %v", pw)
		if len(pw) != 15 {
			t.Fatalf("len(pw)!=15, but %v: %v (%v)", len(pw), pw, i)
		}
	}
}
