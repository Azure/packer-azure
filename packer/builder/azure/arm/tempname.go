// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package arm

import (
	"fmt"

	"github.com/Azure/packer-azure/packer/builder/azure/common/utils"
)

const (
	ResourceGroupNameTemplate = "packer-Resource-Group-%s"
)

type TempName struct {
	alphabet string
	suffix   string

	ComputeName       string
	DeploymentName    string
	ResourceGroupName string
	OSDiskName        string
}

func NewTempName() *TempName {
	tempName := &TempName{
		alphabet: "0123456789bcdfghjklmnpqrstvwxyz",
	}

	tempName.suffix = utils.RandomString(tempName.alphabet, 10)
	tempName.ComputeName = fmt.Sprintf("pkrvm%s", tempName.suffix)
	tempName.DeploymentName = fmt.Sprintf("pkrdp%s", tempName.suffix)
	tempName.OSDiskName = fmt.Sprintf("pkros%s", tempName.suffix)
	tempName.ResourceGroupName = fmt.Sprintf(ResourceGroupNameTemplate, tempName.suffix)

	return tempName
}
