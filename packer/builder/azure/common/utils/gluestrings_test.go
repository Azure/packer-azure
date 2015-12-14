// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package utils

import (
	"testing"
)

func TestGlueStrings(t *testing.T) {
	cases := []struct{ a, b, expected string }{
		{
			"Some log that starts in a",
			"starts in a, but continues in b",
			"Some log that starts in a, but continues in b",
		},
		{
			"",
			"starts in b",
			"starts in b",
		},
	}
	for _, testcase := range cases {
		t.Logf("testcase: %+v\n", testcase)

		result := GlueStrings(testcase.a, testcase.b)
		t.Logf("result: '%s'", result)

		if result != testcase.expected {
			t.Error("expected %q, got %q", testcase.expected, result)
		}
	}
}
