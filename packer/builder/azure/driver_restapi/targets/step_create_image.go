// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package targets

import (
	"fmt"
//	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/constants"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/request"
)

type StepCreateImage struct {
	TmpServiceName string
	TmpVmName string
	UserImageLabel string
	UserImageName string
	RecommendedVMSize string
}

func (s *StepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error Creating Azure Image: %s"

	ui.Say("Creating Azure Image...")

	description := "paker made image"
	imageFamily := "PackerMade"

	requestData := reqManager.CaptureVMImage(s.TmpServiceName, s.TmpVmName, s.UserImageName, s.UserImageLabel,description, imageFamily, s.RecommendedVMSize )
	err := reqManager.ExecuteSync(requestData)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.ImageCreated, 1)
	state.Put(constants.VmExists, 0)

	return multistep.ActionContinue
}

func (s *StepCreateImage) Cleanup(state multistep.StateBag) {
	// do nothing
}
