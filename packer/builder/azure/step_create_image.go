// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package azure

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	msbldcommon "github.com/MSOpenTech/packer-azure/packer/builder/common"
)

type StepCreateImage struct {
	storageAccount string
	tmpVmName string
	userImageLabel string
	userImageName string
	osType string
}

func (s *StepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Creating Azure Image: %s"

	ui.Say("Creating Azure Image...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$storageAccount = '" + s.storageAccount + "';")
	blockBuffer.WriteString("$userImageName = '" + s.userImageName + "';")
	blockBuffer.WriteString("$userImageLabel = '" + s.userImageLabel + "';")
	blockBuffer.WriteString("$tmpVmName = '" + s.tmpVmName + "';")
	blockBuffer.WriteString("$osType = '" + s.osType + "';")
	blockBuffer.WriteString("$containerUrl = \"https://$storageAccount.blob.core.windows.net/vhds\";")
	blockBuffer.WriteString("$mediaLoc = \"$containerUrl/$tmpVmName.vhd\";")
	blockBuffer.WriteString("Add-AzureVMImage -ImageName $userImageName -MediaLocation $mediaLoc -OS $osType -Label $userImageLabel;")
	blockBuffer.WriteString("}")

	err := driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("imageCreated", 1)

	return multistep.ActionContinue
}

func (s *StepCreateImage) Cleanup(state multistep.StateBag) {
	// do nothing
}
