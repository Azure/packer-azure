// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package target

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	ps "github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_powershell/driver"
)

type StepCreateImage struct {
	StorageAccount string
	TmpVmName string
	UserImageLabel string
	UserImageName string
	OsType string
}

func (s *StepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Creating Azure Image: %s"

	ui.Say("Creating Azure Image...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$storageAccount = '" + s.StorageAccount + "';")
	blockBuffer.WriteString("$userImageName = '" + s.UserImageName + "';")
	blockBuffer.WriteString("$userImageLabel = '" + s.UserImageLabel + "';")
	blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")
	blockBuffer.WriteString("$osType = '" + s.OsType + "';")
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
