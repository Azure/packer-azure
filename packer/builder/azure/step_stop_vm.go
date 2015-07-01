// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package azure

import (
	"fmt"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/constants"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/retry"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"github.com/Azure/azure-sdk-for-go/management"
	vm "github.com/Azure/azure-sdk-for-go/management/virtualmachine"
)

type StepStopVm struct {
	TmpVmName      string
	TmpServiceName string
}

func (s *StepStopVm) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error Stopping Temporary Azure VM: %s"

	ui.Say("Stopping Temporary Azure VM...")

	if err := retry.ExecuteAsyncOperation(client, func() (management.OperationID, error) {
		return vm.NewClient(client).ShutdownRole(s.TmpServiceName, s.TmpVmName, s.TmpVmName)
	}); err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.VmRunning, 0)

	return multistep.ActionContinue
}

func (s *StepStopVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
