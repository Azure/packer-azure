// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package targets

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/request"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/constants"
	"log"
	"time"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"regexp"
)

const(
	powerState_Started string = "Started"
	powerState_Stopping string = "Stopping"
	powerState_Stopped string = "Stopped"
	powerState_Unknown string = "Unknown"
)

type StepPollStatus struct {
//	Location string
	TmpServiceName string
//	StorageAccount string
	TmpVmName string
}

func (s *StepPollStatus) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error Polling Temporary Azure VM is ready: %s"

	ui.Say("Polling Temporary Azure VM is ready...")

	firstSleepMin := time.Duration(2)
	firstSleepTime := time.Minute * firstSleepMin
	log.Printf("Sleeping for %v min to make the VM to start", uint(firstSleepMin))
	time.Sleep(firstSleepTime)

	count := 60
	var duration time.Duration = 15
	sleepTime := time.Second * duration

	//	var err error
	var deployment *model.Deployment

	requestData := reqManager.GetDeployment(s.TmpServiceName, s.TmpVmName)

	for count != 0 {
		resp, err := reqManager.Execute(requestData)
		if err != nil {
			pattern := "Request needs to have a x-ms-version header"
			errString := err.Error()
			// Sometimes server returns strange error - ignore it
			ignore, _ := regexp.MatchString(pattern, errString)
			if ignore {
				log.Println("StepPollStatus ignore error: " + errString)
				count--
				continue
			}
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		deployment, err = response.ParseDeployment(resp.Body)
		log.Printf("deployment:\n%v", deployment)

		if len(deployment.RoleInstanceList) > 0 {
			powerState := deployment.RoleInstanceList[0].PowerState

			if powerState == powerState_Started {
				break;
			}

			if powerState == powerState_Stopping ||
				powerState == powerState_Stopped ||
				powerState == powerState_Unknown {
				err := fmt.Errorf(errorMsg, "deployment.RoleInstanceList[0].PowerState is " + powerState)
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

	if(count == 0){
		err := fmt.Errorf(errorMsg, "timeout")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.VmRunning, 1)

	endpoints := deployment.RoleInstanceList[0].InstanceEndpoints
	if len(endpoints) == 0{
		err := fmt.Errorf(errorMsg, "deployment.deployment.RoleInstanceList[0].InstanceEndpoints list is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vip := endpoints[0].Vip
	port := endpoints[0].PublicPort
	endpoint := fmt.Sprintf("%s:%s", vip, port)

	ui.Message("VM Endpoint: " + endpoint)
	state.Put(constants.AzureVmAddr, endpoint)

	roleList := deployment.RoleList
	if len(roleList) == 0{
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
