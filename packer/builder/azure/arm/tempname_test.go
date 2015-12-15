// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package arm

import (
	"strings"
	"testing"
)

func TestTempNameShouldCreatePrefixedRandomNames (t *testing.T) {
	tempName := NewTempName()

	if strings.Index(tempName.ComputeName, "pkrvm") != 0 {
		t.Errorf("Exepcted ComputeName to begin with 'pkrvm', but got '%s'!", tempName.ComputeName)
	}

	if strings.Index(tempName.DeploymentName, "pkrdp") != 0 {
		t.Errorf("Exepcted ComputeName to begin with 'pkrdp', but got '%s'!", tempName.ComputeName)
	}

	if strings.Index(tempName.OSDiskName, "pkros") != 0 {
		t.Errorf("Exepcted OSDiskName to begin with 'pkros', but got '%s'!", tempName.OSDiskName)
	}

	if strings.Index(tempName.ResourceGroupName, "packer-Resource-Group-") != 0 {
		t.Errorf("Exepcted ResourceGroupName to begin with 'packer-Resource-Group-', but got '%s'!", tempName.ResourceGroupName)
	}
}