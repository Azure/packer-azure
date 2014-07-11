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

type StepRemoveDisk struct {
	storageAccount string
	tmpVmName string
}

func (s *StepRemoveDisk) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Removing Temporary Azure Disk: %s"

	ui.Say("Removing Azure Disk...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$storageAccount = '" + s.storageAccount + "';")
	blockBuffer.WriteString("$tmpVmName = '" + s.tmpVmName + "';")

	blockBuffer.WriteString("$containerUrl = \"https://$storageAccount.blob.core.windows.net/vhds\";")
	blockBuffer.WriteString("$mediaLoc = \"$containerUrl/$tmpVmName.vhd\";")

	blockBuffer.WriteString("$disk = Get-AzureDisk | Where-Object {$_.MediaLink –eq $mediaLoc };")
	blockBuffer.WriteString("Remove-AzureDisk -DiskName $disk.DiskName;")
	blockBuffer.WriteString("}")

	err := driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("diskExists", 0)

	return multistep.ActionContinue
}

func (s *StepRemoveDisk) Cleanup(state multistep.StateBag) {
	// do nothing
}
