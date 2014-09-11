// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package lin

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/request"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/constants"
)

type StepCreateVm struct {
	OsType string
	StorageAccount string
	StorageAccountContainer string
	OsImageLabel string
	TmpVmName string
	TmpServiceName string
	InstanceSize string
	Username string
}

func (s *StepCreateVm) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Creating Temporary Azure VM: %s"
	var err error

	certThumbprint := state.Get(constants.UserCertThumbprint).(string)
	if len(certThumbprint) == 0 {
		err := fmt.Errorf(errorMsg, "Certificate Thumbprint is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Creating Temporary Azure VM...")

	osImageName := state.Get(constants.OsImageName).(string)
	if len(osImageName) == 0 {
		err := fmt.Errorf(errorMsg, fmt.Errorf("osImageName is empty"))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt

	}

	mediaLoc := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s.vhd", s.StorageAccount, s.StorageAccountContainer, s.TmpVmName)

	requestData := reqManager.CreateVirtualMachineDeploymentLin(s.TmpServiceName, s.TmpVmName, s.InstanceSize, certThumbprint, s.Username, osImageName, mediaLoc )
	err = reqManager.ExecuteSync(requestData)

	if err != nil {
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
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	ui.Say("Cleanning up...")

	var err error
	var res int

	if res = state.Get(constants.VmRunning).(int); res == 1 {
		ui.Message("Stopping Temporary Azure VM...")
		errorMsg := "Error Stopping Temporary Azure VM: %s"

		requestData := reqManager.ShutdownRoles(s.TmpServiceName, s.TmpVmName)
		err = reqManager.ExecuteSync(requestData)

		if err != nil {
			err = fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}

		state.Put(constants.VmRunning, 0)
	}

	vmExists := state.Get(constants.VmExists).(int)
//	imageCreated := state.Get(constants.ImageCreated).(int)

//	if vmExists == 1 && imageCreated == 0 {
	if vmExists == 1 {
		ui.Message("Removing Temporary Azure VM...")
		errorMsg := "Error Removig Temporary Azure VM: %s"

		requestData := reqManager.DeleteDeploymentAndMedia(s.TmpServiceName, s.TmpVmName)
		err = reqManager.ExecuteSync(requestData)

		if err != nil {
			err = fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}

		state.Put(constants.VmExists, 0)
		state.Put(constants.DiskExists, 0)
	}

	if res = state.Get(constants.DiskExists).(int); res == 1 {
		ui.Message("Removing Temporary Azure Disk...")
		errorMsg := "Error Removing Temporary Azure Disk: %s"

//		var requestData *request.Data
		diskName, ok := state.Get(constants.HardDiskName).(string)
		if ok {
			if len(diskName) == 0 {
				err := fmt.Errorf(errorMsg, err)
				ui.Error(err.Error())
				return
			}

			requestData := reqManager.DeleteDiskAndMedia(diskName)
			err = reqManager.ExecuteSync(requestData)

			if err != nil {
				err := fmt.Errorf(errorMsg, err)
				ui.Error(err.Error())
				return
			}

			state.Put(constants.DiskExists, 0)
		}
	}
}
