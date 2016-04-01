// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	"fmt"

	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/Azure/packer-azure/packer/builder/azure/smapi/retry"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"github.com/Azure/azure-sdk-for-go/management"
	vm "github.com/Azure/azure-sdk-for-go/management/virtualmachine"
)

type StepCreateVm struct{}

func (*StepCreateVm) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get(constants.Config).(*Config)

	errorMsg := "Error Creating temporary Azure VM: %s"

	ui.Say("Creating temporary Azure VM...")

	role := state.Get("role").(*vm.Role)

	options := vm.CreateDeploymentOptions{}
	if config.VNet != "" && config.Subnet != "" {
		options.VirtualNetworkName = config.VNet
	}

	if err := retry.ExecuteAsyncOperation(client, func() (management.OperationID, error) {
		return vm.NewClient(client).CreateDeployment(*role, config.tmpServiceName, options)
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

func (*StepCreateVm) Cleanup(multistep.StateBag) {}
