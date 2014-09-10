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

type StepRemoveDisk struct {
}

func (s *StepRemoveDisk) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error Removing  Azure Temporary Disk: %s"
	ui.Say("Removing Azure Temporary Disk...")

	diskName := state.Get(constants.HardDiskName).(string)
	if len(diskName) == 0 {
		err := fmt.Errorf(errorMsg, fmt.Errorf("HardDiskName is empty"))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	requestData := reqManager.DeleteDisk(diskName)
	err := reqManager.ExecuteSync(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("diskExists", 0)

	return multistep.ActionContinue
}

func (s *StepRemoveDisk) Cleanup(state multistep.StateBag) {
	// do nothing
}
