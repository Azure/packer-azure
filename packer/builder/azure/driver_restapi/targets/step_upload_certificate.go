// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package targets

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/constants"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/retry"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/hostedservice"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepUploadCertificate struct {
	TmpServiceName string
}

func (s *StepUploadCertificate) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	ui := state.Get("ui").(packer.Ui)
	errorMsg := "Error Uploading Temporary Certificate: %s"
	var err error

	ui.Say("Uploading Temporary Certificate...")

	userCertPath := state.Get(constants.UserCertPath).(string)
	if len(userCertPath) == 0 {
		err = fmt.Errorf("StepUploadCertificate CertPath is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println("userCertPath: " + userCertPath)

	var certData []byte
	certData, err = ioutil.ReadFile(userCertPath)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err = retry.ExecuteAsyncOperation(client, func() (management.OperationID, error) {
		return hostedservice.NewClient(client).AddCertificate(s.TmpServiceName, certData, hostedservice.CertificateFormatPfx, "")
	}); err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.CertUploaded, 1)

	return multistep.ActionContinue
}

func (s *StepUploadCertificate) Cleanup(state multistep.StateBag) {
	// do nothing
}
