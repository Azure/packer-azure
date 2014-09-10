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

type StepRemoveDisk struct {
	StorageAccount string
	TmpVmName string
	ContainerUrl string
}

func (s *StepRemoveDisk) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Removing  Azure Temporary Disk: %s"

	ui.Say("Removing Azure Temporary Disk...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$storageAccount = '" + s.StorageAccount + "';")
	blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")

	blockBuffer.WriteString("$containerUrl = '" + s.ContainerUrl + "';")
	blockBuffer.WriteString("$mediaLoc = \"$containerUrl/$tmpVmName.vhd\";")

	blockBuffer.WriteString("$disk = Get-AzureDisk | Where-Object {$_.MediaLink –eq $mediaLoc };")
	blockBuffer.WriteString("Remove-AzureDisk -DiskName $disk.DiskName;")
	blockBuffer.WriteString("}")

	err := driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())

		// this is not critical - report error and continue
		return multistep.ActionContinue
	}

	state.Put("diskExists", 0)

	return multistep.ActionContinue
}

func (s *StepRemoveDisk) Cleanup(state multistep.StateBag) {
	// do nothing
}
