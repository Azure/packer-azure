// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package win

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
	Password         string
}

func (s *StepCreateVm) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error creating temporary Azure VM: %s"

	ui.Say("Creating temporary Azure VM...")

	osImageName := state.Get(constants.OSImageName).(string)
	if len(osImageName) == 0 {
		err := fmt.Errorf(errorMsg, fmt.Errorf("osImageName is empty"))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	isOSImage := state.Get(constants.IsOSImage).(bool)

	mediaLoc := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s.vhd", s.StorageAccount, s.StorageContainer, s.TmpVmName)

	role := createRole(isOSImage, s.TmpVmName, s.InstanceSize, s.Username, s.Password, osImageName, mediaLoc)
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

func createRole(isOSImage bool, vmName, vmSize, userName, password, osImageName, mediaLoc string) (role vm.Role) {
	role = vmutils.NewVMConfiguration(vmName, vmSize)
	vmutils.ConfigureForWindows(&role, vmName, userName, password, true, "")
	vmutils.ConfigureWithPublicRDP(&role)
	vmutils.ConfigureWithPublicPowerShell(&role)

	if isOSImage {
		vmutils.ConfigureDeploymentFromPlatformImage(&role, osImageName, mediaLoc, "")
	} else {
		vmutils.ConfigureDeploymentFromVMImage(&role, osImageName, mediaLoc, true)
	}
	return role
}
