// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package win

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	ps "github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_powershell/driver"
)

type StepCreateVm struct {
	OsType string
	StorageAccount string
	OsImageLabel string
	Location string
	TmpVmName string
	TmpServiceName string
	InstanceSize string
	Username string
	Password string
	ContainerUrl string
}

func (s *StepCreateVm) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Creating Temporary Azure VM: %s"

	ui.Say("Creating Temporary Azure VM...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$storageAccount = '" + s.StorageAccount + "';")
	blockBuffer.WriteString("$osImageLabel = '" + s.OsImageLabel + "';")
	blockBuffer.WriteString("$location = '" + s.Location + "';")
	blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
	blockBuffer.WriteString("$instanceSize = '" + s.InstanceSize + "';")
	blockBuffer.WriteString("$username = '" + s.Username + "';")
	blockBuffer.WriteString("$password = '" + s.Password + "';")

	blockBuffer.WriteString("$containerUrl = '" + s.ContainerUrl + "';")
	blockBuffer.WriteString("$mediaLoc = \"$containerUrl/$tmpVmName.vhd\";")

	blockBuffer.WriteString("$image = Get-AzureVMImage | where { $_.Label -Match $osImageLabel } | where { $_.Location.Split(';') -contains $location} | Sort-Object -Descending -Property PublishedDate | Select -First 1;")
	blockBuffer.WriteString("$myVM = New-AzureVMConfig -Name $tmpVmName -InstanceSize $instanceSize -ImageName $image.ImageName -DiskLabel 'PackerMade' -MediaLocation $mediaLoc | Add-AzureProvisioningConfig -Windows -Password $password -AdminUsername $username;")

	blockBuffer.WriteString("New-AzureVM -ServiceName $tmpServiceName -VMs $myVM -Location $location -WaitForBoot;")
	blockBuffer.WriteString("}")

	err := driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vmRunning", 1)
	state.Put("vmExists", 1)
	state.Put("srvExists", 1)
	state.Put("diskExists", 1)

	return multistep.ActionContinue
}

func (s *StepCreateVm) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Cleanning up...")

	var err error
	var res int

	if res = state.Get("vmRunning").(int); res == 1 {
		ui.Say("Stopping Temporary Azure VM...")
		errorMsg := "Error Stopping Temporary Azure VM: %s"

		var blockBuffer bytes.Buffer
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")
		blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
		blockBuffer.WriteString("Stop-AzureVM -ServiceName $tmpServiceName -Name $tmpVmName -Force;")
		blockBuffer.WriteString("}")

		err = driver.Exec( blockBuffer.String() )

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}
	}

	if res = state.Get("vmExists").(int); res == 1 {
		ui.Say("Removing Temporary Azure VM...")
		errorMsg := "Error Removig Temporary Azure VM: %s"

		var blockBuffer bytes.Buffer
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")
		blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
		blockBuffer.WriteString("Remove-AzureVM -ServiceName $tmpServiceName -Name $tmpVmName;")
		blockBuffer.WriteString("}")

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}
	}

	if res = state.Get("srvExists").(int); res == 1 {
		ui.Say("Removing Azure Temporary Service...")
		errorMsg := "Error Removing Temporary Azure Service: %s"

		var blockBuffer bytes.Buffer
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
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
		blockBuffer.WriteString("$storageAccount = '" + s.StorageAccount + "';")
		blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")

		blockBuffer.WriteString("$containerUrl = '" + s.ContainerUrl + "';")
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

	if res = state.Get("imageCreated").(int); res == 0 {
		// TODO: remove vhd
	}
}
