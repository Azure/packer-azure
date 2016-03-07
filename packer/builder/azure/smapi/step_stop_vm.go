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
		return vm.NewClient(client).ShutdownRole(s.TmpServiceName, s.TmpVmName, s.TmpVmName, vm.PostShutdownActionStopped)
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
