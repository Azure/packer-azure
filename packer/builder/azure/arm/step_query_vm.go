// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package arm

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"

	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepQueryVM struct {
	client *AzureClient
	query  func(resourceGroupName string, computeName string) (compute.VirtualMachine, error)
	say    func(message string)
	error  func(e error)
}

func NewStepQueryVM(client *AzureClient, ui packer.Ui) *StepQueryVM {
	var step = &StepQueryVM{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.query = step.queryCompute
	return step
}

func (s *StepQueryVM) queryCompute(resourceGroupName string, computeName string) (compute.VirtualMachine, error) {
	return s.client.VirtualMachinesClient.Get(resourceGroupName, computeName, "")
}

func (s *StepQueryVM) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Querying the machine's properties ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var computeName = state.Get(constants.ArmComputeName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ComputeName       : '%s'", computeName))

	vm, err := s.query(resourceGroupName, computeName)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	s.say(fmt.Sprintf(" -> OS Disk           : '%s'", *vm.Properties.StorageProfile.OsDisk.Vhd.URI))
	state.Put(constants.ArmOSDiskVhd, *vm.Properties.StorageProfile.OsDisk.Vhd.URI)

	return multistep.ActionContinue
}

func (*StepQueryVM) Cleanup(multistep.StateBag) {
}
