// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package targets

import (
	"fmt"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/constants"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/request"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepCreateService struct {
	Location       string
	TmpServiceName string
	VNet           string
}

func (s *StepCreateService) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error Creating Temporary Azure Service: %s"

	ui.Say("Creating Temporary Azure Service...")

	requestData := reqManager.CreateCloudService(s.TmpServiceName, s.Location, s.VNet)
	err := reqManager.ExecuteSync(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.SrvExists, 1)

	return multistep.ActionContinue
}

func (s *StepCreateService) Cleanup(state multistep.StateBag) {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	var err error
	var res int

	if res = state.Get(constants.SrvExists).(int); res == 1 {
		ui.Say("Removing Temporary Azure Service and It's Deployments If Any...")
		errorMsg := "Error Removing Temporary Azure Service: %s"

		var requestData *request.Data
		requestData = reqManager.DeleteCloudServiceAndMedia(s.TmpServiceName)

		err = reqManager.ExecuteSync(requestData)

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			ui.Error(err.Error())
			return
		}
	}
}
