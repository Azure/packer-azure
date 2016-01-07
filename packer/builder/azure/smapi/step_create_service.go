// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	"encoding/base64"
	"fmt"

	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/Azure/packer-azure/packer/builder/azure/smapi/retry"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/hostedservice"
)

type StepCreateService struct {
	Location       string
	TmpServiceName string
}

func (s *StepCreateService) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	hsc := hostedservice.NewClient(client)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error creating temporary Azure service: %s"

	ui.Say("Creating temporary Azure service...")

	if err := hsc.CreateHostedService(hostedservice.CreateHostedServiceParameters{
		ServiceName: s.TmpServiceName,
		Location:    s.Location,
		Label:       base64.StdEncoding.EncodeToString([]byte(s.TmpServiceName)),
	}); err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.SrvExists, 1)

	return multistep.ActionContinue
}

func (s *StepCreateService) Cleanup(state multistep.StateBag) {
	client := state.Get(constants.RequestManager).(management.Client)
	hsc := hostedservice.NewClient(client)
	ui := state.Get(constants.Ui).(packer.Ui)

	if res := state.Get(constants.SrvExists).(int); res == 1 {
		ui.Say("Removing temporary Azure service and its deployments, if any...")
		errorMsg := "Error removing temporary Azure service: %s"

		if err := retry.ExecuteAsyncOperation(client, func() (management.OperationID, error) {
			return hsc.DeleteHostedService(s.TmpServiceName, true)
		}); err != nil {
			ui.Error(fmt.Sprintf(errorMsg, err))
			return
		}
	}
}
