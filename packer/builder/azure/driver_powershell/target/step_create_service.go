// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package target

import (
	"bytes"
	"fmt"
	ps "github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_powershell/driver"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepCreateService struct {
	Location       string
	TmpServiceName string
	StorageAccount string
	TmpVmName      string
	ContainerUrl   string
}

func (s *StepCreateService) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Creating Temporary Azure Service: %s"

	ui.Say("Creating Temporary Azure Service...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$location = '" + s.Location + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
	blockBuffer.WriteString("New-AzureService -ServiceName $tmpServiceName -Location $location;")
	blockBuffer.WriteString("}")

	err := driver.Exec(blockBuffer.String())

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
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)

	var err error
	var res int

	if res = state.Get("srvExists").(int); res == 1 {
		ui.Say("Removing Temporary Azure Service...")
		errorMsg := "Error Removing Temporary Azure Service: %s"

		var blockBuffer bytes.Buffer
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
		blockBuffer.WriteString("Remove-AzureService -ServiceName $tmpServiceName -Force;")
		blockBuffer.WriteString("}")

		err = driver.Exec(blockBuffer.String())

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}
	}

	if res = state.Get("diskExists").(int); res == 1 {
		ui.Say("Removing Temporary Azure Disk...")
		errorMsg := "Error Removing Temporary Azure Disk: %s"

		var blockBuffer bytes.Buffer
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$storageAccount = '" + s.StorageAccount + "';")
		blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")

		blockBuffer.WriteString("$containerUrl = '" + s.ContainerUrl + "';")
		blockBuffer.WriteString("$mediaLoc = \"$containerUrl/$tmpVmName.vhd\";")

		blockBuffer.WriteString("$disk = Get-AzureDisk | Where-Object {$_.MediaLink –eq $mediaLoc };")
		blockBuffer.WriteString("Remove-AzureDisk -DiskName $disk.DiskName;")
		blockBuffer.WriteString("}")

		err = driver.Exec(blockBuffer.String())

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}
	}
}
