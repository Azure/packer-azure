// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package lin

import (
	"fmt"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/constants"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/retry"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"github.com/Azure/azure-sdk-for-go/management"
	vm "github.com/Azure/azure-sdk-for-go/management/virtualmachine"
	"github.com/Azure/azure-sdk-for-go/management/vmutils"
)

type StepCreateVm struct {
	StorageAccount   string
	StorageContainer string
	TmpVmName        string
	TmpServiceName   string
	InstanceSize     string
	Username         string
}

func (s *StepCreateVm) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Creating Temporary Azure VM: %s"

	certThumbprint := state.Get(constants.Thumbprint).(string)
	if len(certThumbprint) == 0 {
		err := fmt.Errorf(errorMsg, "Certificate Thumbprint is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Creating Temporary Azure VM...")

	osImageName := state.Get(constants.OSImageName).(string)
	if len(osImageName) == 0 {
		err := fmt.Errorf(errorMsg, fmt.Errorf("osImageName is empty"))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	isOSImage := state.Get(constants.IsOSImage).(bool)

	mediaLoc := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s.vhd", s.StorageAccount, s.StorageContainer, s.TmpVmName)

	role := createRole(isOSImage, s.TmpVmName, s.InstanceSize, certThumbprint, s.Username, osImageName, mediaLoc)
	if err := retry.ExecuteAsyncOperation(client, func() (management.OperationID, error) {
		return vm.NewClient(client).CreateDeployment(role, s.TmpServiceName, vm.CreateDeploymentOptions{})
	}); err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.VmExists, 1)
	state.Put(constants.DiskExists, 1)

	return multistep.ActionContinue
}

func (s *StepCreateVm) Cleanup(state multistep.StateBag) {
	// do nothing
}

func createRole(isOSImage bool, vmName, vmSize, certThumbprint, userName, osImageName, mediaLoc string) (role vm.Role) {
	role = vmutils.NewVMConfiguration(vmName, vmSize)
	vmutils.ConfigureForLinux(&role, vmName, userName, "", certThumbprint)
	vmutils.ConfigureWithPublicSSH(&role)
	if isOSImage {
		vmutils.ConfigureDeploymentFromPlatformImage(&role, osImageName, mediaLoc, "")
	} else {
		vmutils.ConfigureDeploymentFromVMImage(&role, osImageName, mediaLoc, true)
	}
	return role
}
