// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package utils

import (
	"testing"
)

func TestBuildContainerName_generates_unique_names(t *testing.T) {
	one := BuildContainerName()
	two := BuildContainerName()

	if one == two {
		t.Fatalf("expected generated names to be different: %s <-> %s", one, two)
	}
}
