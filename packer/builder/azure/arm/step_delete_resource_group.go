// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"fmt"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/packer-azure/packer/builder/azure/common"
	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepDeleteResourceGroup struct {
	client *AzureClient
	delete func(resourceGroupName string) error
	say    func(message string)
	error  func(e error)
}

func NewStepDeleteResourceGroup(client *AzureClient, ui packer.Ui) *StepDeleteResourceGroup {
	var step = &StepDeleteResourceGroup{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.delete = step.deleteResourceGroup
	return step
}

func (s *StepDeleteResourceGroup) deleteResourceGroup(resourceGroupName string) error {
	_, err := s.client.GroupsClient.Delete(resourceGroupName)
	if err != nil {
		return err
	}

	return s.waitForResourceGroupToBeDeleted(resourceGroupName)
}

// The API does not correctly poll for completion of the Delete command for the groups client, so here's a workaround.
//
// There are two polling variations that I am aware of.
//  1. API returns a location header and HTTP status code of 202.  Client must poll the location until complete.
//  2. API returns an azure async operation header.
func (s *StepDeleteResourceGroup) waitForResourceGroupToBeDeleted(resourceGroupName string) error {
	deadline := time.Now().Add(15 * time.Minute)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("Exceeded poll duration while waiting for resource group %s to be deleted", resourceGroupName)
		}

		_, err := s.client.GroupsClient.Get(resourceGroupName)
		if err != nil {
			detailedError, ok := err.(autorest.DetailedError)
			if ok && detailedError.StatusCode == 404 {
				break
			}

		}

		time.Sleep(15 * time.Second)
	}

	return nil
}

func (s *StepDeleteResourceGroup) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Deleting resource group ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))

	result := common.StartInterruptibleTask(
		func() bool { return common.IsStateCancelled(state) },
		func() error { return s.delete(resourceGroupName) })

	return processInterruptibleResult(result, s.error, state)
}

func (*StepDeleteResourceGroup) Cleanup(multistep.StateBag) {
}
