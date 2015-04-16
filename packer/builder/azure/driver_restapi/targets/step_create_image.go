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
	"log"
	"strings"
	"time"
)

type StepCreateImage struct {
	TmpServiceName    string
	TmpVmName         string
	UserImageLabel    string
	UserImageName     string
	RecommendedVMSize string
}

func (s *StepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error Creating Azure Image: %s"

	ui.Say("Creating Azure Image. If Successful, This Will Remove the Temporary VM...")

	description := "paker made image"
	imageFamily := "PackerMade"

	requestData := reqManager.CaptureVMImage(s.TmpServiceName, s.TmpVmName, s.UserImageName, s.UserImageLabel, description, imageFamily, s.RecommendedVMSize)
	err := reqManager.ExecuteSync(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// CatpureVMImage removes the VM
	state.Put(constants.ImageCreated, 1)
	state.Put(constants.VmExists, 0)

	return multistep.ActionContinue
}

func (s *StepCreateImage) Cleanup(state multistep.StateBag) {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	var err error
	var res int

	if res = state.Get(constants.VmExists).(int); res == 1 { //VM was not removed at image creation step
		return
	}

	// Since VM was successfully removed - remove it's media as well

	if res = state.Get(constants.DiskExists).(int); res == 1 {
		ui.Message("Removing Temporary Azure Disk...")
		errorMsg := "Error Removing Temporary Azure Disk: %s"

		diskName, ok := state.Get(constants.HardDiskName).(string)
		if ok {
			if len(diskName) == 0 {
				err := fmt.Errorf(errorMsg, err)
				ui.Error(err.Error())
				return
			}

			requestData := reqManager.DeleteDiskAndMedia(diskName)

			const stepsLimit int = 10
			stepNumber := 0
			for {
				err = reqManager.ExecuteSync(requestData)

				if err == nil {
					break
				}

				patterns := []string{
					"is currently performing an operation on deployment",
					"is currently in use by virtual machine",
				}

				needToRetry := false

				for _, pattern := range patterns {
					if strings.Contains(err.Error(), pattern) {
						needToRetry = true
						break
					}
				}

				if needToRetry {
					stepNumber++
					if stepNumber == stepsLimit {
						err := fmt.Errorf(errorMsg, err)
						ui.Error(err.Error())
						return
					}

					const p = 30
					log.Println(fmt.Sprintf("Disk is in use. Waiting for %d sec (%d of %d)", uint(p), stepNumber, stepsLimit))
					time.Sleep(time.Second * p)
					continue
				}

				err := fmt.Errorf(errorMsg, err)
				ui.Error(err.Error())
				return
			}

			state.Put(constants.DiskExists, 0)
		}
	}
}
