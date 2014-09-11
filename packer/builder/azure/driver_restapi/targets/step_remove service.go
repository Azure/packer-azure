// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package targets

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/constants"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/request"
)

type StepRemoveService struct {
	TmpServiceName string
}

func (s *StepRemoveService) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error Removing Temporary Azure Service: %s"

	ui.Say("Removing Temporary Azure Service...")

	requestData := reqManager.DeleteCloudService(s.TmpServiceName)
	err := reqManager.ExecuteSync(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.SrvExists, 0)

	return multistep.ActionContinue
}

func (s *StepRemoveService) Cleanup(state multistep.StateBag) {
	// do nothing
}
