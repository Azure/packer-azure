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

type StepRemoveVm struct {
	TmpVmName      string
	TmpServiceName string
}

func (s *StepRemoveVm) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error Removig Temporary Azure VM: %s"

	ui.Say("Removing Temporary Azure VM...")

	requestData := reqManager.DeleteDeployment(s.TmpServiceName, s.TmpVmName)
	err := reqManager.ExecuteSync(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.VmExists, 0)

	return multistep.ActionContinue
}

func (s *StepRemoveVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
