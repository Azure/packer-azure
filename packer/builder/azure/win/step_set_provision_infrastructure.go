// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package win

import (
	"fmt"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/constants"
	"github.com/MSOpenTech/packer-azure/packer/communicator/azureVmCustomScriptExtension"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/storageservice"
	"github.com/Azure/azure-sdk-for-go/storage"
)

type StepSetProvisionInfrastructure struct {
	VmName                   string
	ServiceName              string
	StorageAccountName       string
	TempContainerName        string
	storageClient            storage.Client
	flagTempContainerCreated bool
}

func (s *StepSetProvisionInfrastructure) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	client := state.Get(constants.RequestManager).(management.Client)

	errorMsg := "Error StepRemoteSession: %s"
	ui.Say("Preparing infrastructure for provision...")

	// get key for storage account
	ui.Message("Getting key for storage account...")

	keys, err := storageservice.NewClient(client).GetStorageServiceKeys(s.StorageAccountName)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	//create storage driver
	s.storageClient, err = storage.NewBasicClient(s.StorageAccountName, keys.PrimaryKey)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	//create temporary container
	s.flagTempContainerCreated = false

	ui.Message("Creating Azure temporary container...")
	err = s.storageClient.GetBlobService().CreateContainer(s.TempContainerName, storage.ContainerAccessTypePrivate)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.flagTempContainerCreated = true

	isOSImage := state.Get(constants.IsOSImage).(bool)

	comm, err := azureVmCustomScriptExtension.New(
		azureVmCustomScriptExtension.Config{
			ServiceName:        s.ServiceName,
			VmName:             s.VmName,
			StorageAccountName: s.StorageAccountName,
			StorageAccountKey:  keys.PrimaryKey,
			ContainerName:      s.TempContainerName,
			Ui:                 ui,
			IsOSImage:          isOSImage,
			ManagementClient:   client,
		})

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	packerCommunicator := packer.Communicator(comm)

	state.Put("communicator", packerCommunicator)

	return multistep.ActionContinue
}

func (s *StepSetProvisionInfrastructure) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Cleaning Up Infrastructure for provision...")

	if s.flagTempContainerCreated {
		ui.Message("Removing Azure temporary container...")

		err := s.storageClient.GetBlobService().DeleteContainer(s.TempContainerName)
		if err != nil {
			ui.Message("Error removing temporary container: " + err.Error())
		}
	}
}
