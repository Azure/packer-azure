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

type StepCreateService struct {
	location string
	tmpServiceName string
	storageAccount string
	tmpVmName string
}

func (s *StepCreateService) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Creating Temporary Azure Service: %s"

	ui.Say("Creating Temporary Azure Service...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$location = '" + s.location + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.tmpServiceName + "';")
	blockBuffer.WriteString("New-AzureService -ServiceName $tmpServiceName -Location $location;")
	blockBuffer.WriteString("}")

	err := driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("srvExists", 1)

	return multistep.ActionContinue
}

func (s *StepCreateService) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	var err error
	var res int

	if res = state.Get("srvExists").(int); res == 1 {
		ui.Say("Removing Azure Temporary Service...")
		errorMsg := "Error Removing Temporary Azure Service: %s"

		var blockBuffer bytes.Buffer
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$tmpServiceName = '" + s.tmpServiceName + "';")
		blockBuffer.WriteString("Remove-AzureService -ServiceName $tmpServiceName -Force;")
		blockBuffer.WriteString("}")

		err = driver.Exec( blockBuffer.String() )

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}
	}

	if res = state.Get("diskExists").(int); res == 1 {
		ui.Say("Removing Azure Temporary Disk...")
		errorMsg := "Error Removing Temporary Azure Disk: %s"

		var blockBuffer bytes.Buffer
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$storageAccount = '" + s.storageAccount + "';")
		blockBuffer.WriteString("$tmpVmName = '" + s.tmpVmName + "';")

		blockBuffer.WriteString("$containerUrl = \"https://$storageAccount.blob.core.windows.net/vhds\";")
		blockBuffer.WriteString("$mediaLoc = \"$containerUrl/$tmpVmName.vhd\";")

		blockBuffer.WriteString("$disk = Get-AzureDisk | Where-Object {$_.MediaLink –eq $mediaLoc };")
		blockBuffer.WriteString("Remove-AzureDisk -DiskName $disk.DiskName;")
		blockBuffer.WriteString("}")

		err = driver.Exec( blockBuffer.String() )

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}
	}
}
