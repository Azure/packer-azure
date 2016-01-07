// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"github.com/Azure/azure-sdk-for-go/management"
	vm "github.com/Azure/azure-sdk-for-go/management/virtualmachine"
)

type StepPollStatus struct {
	TmpServiceName string
	TmpVmName      string
	OSType         string
}

func (s *StepPollStatus) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	vmc := vm.NewClient(client)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error polling temporary Azure VM is ready: %s"

	ui.Say("Polling till temporary Azure VM is ready...")

	if len(s.OSType) == 0 {
		err := fmt.Errorf(errorMsg, "'OSType' param is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var count uint = 60
	var duration time.Duration = 40
	sleepTime := time.Second * duration
	total := count * uint(duration)

	var deployment vm.DeploymentResponse

	for count > 0 {
		var err error // deployment needs to be accessed outside of this loop, can't use :=
		deployment, err = vmc.GetDeployment(s.TmpServiceName, s.TmpVmName)
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(deployment.RoleInstanceList) > 0 {
			powerState := deployment.RoleInstanceList[0].PowerState
			instanceStatus := deployment.RoleInstanceList[0].InstanceStatus

			if powerState == vm.PowerStateStarted && instanceStatus == vm.InstanceStatusReadyRole {
				break
			}

			if instanceStatus == vm.InstanceStatusFailedStartingRole ||
				instanceStatus == vm.InstanceStatusFailedStartingVM ||
				instanceStatus == vm.InstanceStatusUnresponsiveRole {
				err := fmt.Errorf(errorMsg, "deployment.RoleInstanceList[0].instanceStatus is "+instanceStatus)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			if powerState == vm.PowerStateStopping ||
				powerState == vm.PowerStateStopped ||
				powerState == vm.PowerStateUnknown {
				err := fmt.Errorf(errorMsg, "deployment.RoleInstanceList[0].PowerState is "+powerState)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}

		// powerState_Starting or deployment.RoleInstanceList[0] == 0
		log.Println(fmt.Sprintf("Waiting for another %v seconds...", uint(duration)))
		time.Sleep(sleepTime)
		count--
	}

	if count == 0 {
		err := fmt.Errorf(errorMsg, fmt.Sprintf("time is up (%d seconds)", total))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.VmRunning, 1)

	log.Println("s.OSType = " + s.OSType)

	if s.OSType == constants.Target_Linux {
		endpoints := deployment.RoleInstanceList[0].InstanceEndpoints
		if len(endpoints) == 0 {
			err := fmt.Errorf(errorMsg, "deployment.RoleInstanceList[0].InstanceEndpoints list is empty")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		vip := endpoints[0].Vip
		state.Put(constants.SSHHost, vip)

		ui.Message("VM Endpoint: " + vip)
	}

	roleList := deployment.RoleList
	if len(roleList) == 0 {
		err := fmt.Errorf(errorMsg, "deployment.RoleList is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	diskName := roleList[0].OSVirtualHardDisk.DiskName
	ui.Message("VM DiskName: " + diskName)
	state.Put(constants.HardDiskName, diskName)

	mediaLink := roleList[0].OSVirtualHardDisk.MediaLink
	ui.Message("VM MediaLink: " + mediaLink)
	state.Put(constants.MediaLink, mediaLink)

	return multistep.ActionContinue
}

func (s *StepPollStatus) Cleanup(state multistep.StateBag) {
	// nothing to do
}
