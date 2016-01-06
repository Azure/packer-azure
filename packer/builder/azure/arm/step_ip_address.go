// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package arm

import (
	"fmt"

	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepIPAddress struct {
	client *AzureClient
	get    func(resourceGroupName string, ipAddressName string) (string, error)
	say    func(message string)
	error  func(e error)
}

func NewStepIPAddress(client *AzureClient, ui packer.Ui) *StepIPAddress {
	var step = &StepIPAddress{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.get = step.getIPAddress
	return step
}

func (s *StepIPAddress) getIPAddress(resourceGroupName string, ipAddressName string) (string, error) {
	res, err := s.client.PublicIPAddressesClient.Get(resourceGroupName, ipAddressName)
	if err != nil {
		return "", nil
	}

	return *res.Properties.IPAddress, nil
}

func (s *StepIPAddress) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Getting the public IP address ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var ipAddressName = state.Get(constants.ArmPublicIPAddressName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName   : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> PublicIPAddressName : '%s'", ipAddressName))

	address, err := s.get(resourceGroupName, ipAddressName)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	s.say(fmt.Sprintf(" -> SSHHost             : '%s'", address))
	state.Put(constants.SSHHost, address)

	return multistep.ActionContinue
}

func (*StepIPAddress) Cleanup(multistep.StateBag) {
}
