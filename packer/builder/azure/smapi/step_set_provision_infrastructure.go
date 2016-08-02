// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	"fmt"

	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/Azure/packer-azure/packer/communicator/azureVmCustomScriptExtension"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/storage"
)

type StepSetProvisionInfrastructure struct {
	VmName                    string
	ServiceName               string
	StorageAccountName        string
	TempContainerName         string
	ProvisionTimeoutInMinutes uint

	flagTempContainerCreated bool
}

func (s *StepSetProvisionInfrastructure) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	client := state.Get(constants.RequestManager).(management.Client)
	config := state.Get(constants.Config).(*Config)

	errorMsg := "Error StepSetProvisionInfrastructure: %s"
	ui.Say("Preparing infrastructure for provision...")

	// get key for storage account
	ui.Message("Looking up storage account...")

	//create temporary container
	s.flagTempContainerCreated = false

	ui.Message("Creating Azure temporary container...")
	err := config.storageClient.GetBlobService().CreateContainer(s.TempContainerName, storage.ContainerAccessTypePrivate)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.flagTempContainerCreated = true

	comm := azureVmCustomScriptExtension.New(
		azureVmCustomScriptExtension.Config{
			ServiceName:               s.ServiceName,
			VmName:                    s.VmName,
			StorageAccountName:        s.StorageAccountName,
			StorageAccountKey:         config.storageAccountKey,
			BlobClient:                config.storageClient.GetBlobService(),
			ContainerName:             s.TempContainerName,
			Ui:                        ui,
			ManagementClient:          client,
			ProvisionTimeoutInMinutes: s.ProvisionTimeoutInMinutes,
		})

	packerCommunicator := packer.Communicator(comm)

	state.Put("communicator", packerCommunicator)

	return multistep.ActionContinue
}

func (s *StepSetProvisionInfrastructure) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get(constants.Config).(*Config)
	ui.Say("Cleaning Up Infrastructure for provision...")

	if s.flagTempContainerCreated {
		ui.Message("Removing Azure temporary container...")

		err := config.storageClient.GetBlobService().DeleteContainer(s.TempContainerName)
		if err != nil {
			ui.Message("Error removing temporary container: " + err.Error())
		}
	}
}
