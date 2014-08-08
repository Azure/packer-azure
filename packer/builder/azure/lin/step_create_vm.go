// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package lin

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	msbldcommon "github.com/MSOpenTech/packer-azure/packer/builder/common"
)

type StepCreateVm struct {
	OsType string
	StorageAccount string
	OsImageLabel string
	TmpVmName string
	TmpServiceName string
	InstanceSize string
	Username string
}

func (s *StepCreateVm) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Creating Temporary Azure VM: %s"

	certThumbprint := state.Get("certThumbprint").(string)
	if len(certThumbprint) == 0 {
		err := fmt.Errorf(errorMsg, "Certificate Thumbprint is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}


	ui.Say("Creating Temporary Azure VM...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$storageAccount = '" + s.StorageAccount + "';")
	blockBuffer.WriteString("$osImageLabel = '" + s.OsImageLabel + "';")
//	blockBuffer.WriteString("$location = '" + s.location + "';")
	blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
	blockBuffer.WriteString("$instanceSize = '" + s.InstanceSize + "';")
	blockBuffer.WriteString("$username = '" + s.Username + "';")

	blockBuffer.WriteString("$containerUrl = \"https://$storageAccount.blob.core.windows.net/vhds\";")
	blockBuffer.WriteString("$mediaLoc = \"$containerUrl/$tmpVmName.vhd\";")

//	blockBuffer.WriteString("$image = Get-AzureVMImage | where { $_.Label -eq $osImageLabel } | where { $_.Location.Split(';') -contains $location} | Sort-Object -Descending -Property PublishedDate | Select -First 1;")
	blockBuffer.WriteString("$image = Get-AzureVMImage | where { ($_.Label -like $osImageLabel) -or ($_.ImageFamily -like $osImageLabel) } | Sort-Object -Descending -Property PublishedDate | Select -First 1;")

	blockBuffer.WriteString("$certThumbprint = '" + certThumbprint + "';")
	blockBuffer.WriteString("$sshkey = New-AzureSSHKey -PublicKey -Fingerprint $certThumbprint -Path \"/home/$username/.ssh/authorized_keys\";")
	blockBuffer.WriteString("$myVM = New-AzureVMConfig -Name $tmpVmName -InstanceSize $instanceSize -ImageName $image.ImageName -DiskLabel 'PackerMade' -MediaLocation $mediaLoc | Add-AzureProvisioningConfig -Linux -NoSSHPassword -LinuxUser $username -SSHPublicKeys $sshKey;")
//	blockBuffer.WriteString("$myVM = New-AzureVMConfig -Name $tmpVmName -InstanceSize $instanceSize -ImageName $image.ImageName -DiskLabel 'PackerMade' -MediaLocation $mediaLoc | Add-AzureProvisioningConfig -Linux -Password $password -LinuxUser $username;")

//	blockBuffer.WriteString("New-AzureVM -ServiceName $tmpServiceName -VMs $myVM -Location $location -WaitForBoot;")
	blockBuffer.WriteString("New-AzureVM -ServiceName $tmpServiceName -VMs $myVM -WaitForBoot;")
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
//	state.Put("srvExists", 1)
	state.Put("diskExists", 1)

	return multistep.ActionContinue
}

func (s *StepCreateVm) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(msbldcommon.Driver)
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

//	if res = state.Get("srvExists").(int); res == 1 {
//		ui.Say("Removing Azure Temporary Service...")
//		errorMsg := "Error Removing Temporary Azure Service: %s"
//
//		var blockBuffer bytes.Buffer
//		blockBuffer.WriteString("Invoke-Command -scriptblock {")
//		blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
//		blockBuffer.WriteString("Remove-AzureService -ServiceName $tmpServiceName -Force;")
//		blockBuffer.WriteString("}")
//
//		err = driver.Exec( blockBuffer.String() )
//
//		if err != nil {
//			err := fmt.Errorf(errorMsg, err)
//			ui.Error(err.Error())
//			return
//		}
//	}

//	if res = state.Get("diskExists").(int); res == 1 {
//		ui.Say("Removing Azure Disk...")
//		errorMsg := "Error Removing Temporary Azure Disk: %s"
//
//		var blockBuffer bytes.Buffer
//		blockBuffer.WriteString("Invoke-Command -scriptblock {")
//		blockBuffer.WriteString("$storageAccount = '" + s.StorageAccount + "';")
//		blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")
//
//		blockBuffer.WriteString("$containerUrl = \"https://$storageAccount.blob.core.windows.net/vhds\";")
//		blockBuffer.WriteString("$mediaLoc = \"$containerUrl/$tmpVmName.vhd\";")
//
//		blockBuffer.WriteString("$disk = Get-AzureDisk | Where-Object {$_.MediaLink –eq $mediaLoc };")
//		blockBuffer.WriteString("Remove-AzureDisk -DiskName $disk.DiskName;")
//		blockBuffer.WriteString("}")
//
//		err = driver.Exec( blockBuffer.String() )
//
//		if err != nil {
//			err := fmt.Errorf(errorMsg, err)
//			ui.Error(err.Error())
//			return
//		}
//	}

	if res = state.Get("imageCreated").(int); res == 0 {
		// TODO: remove vhd
	}
}
