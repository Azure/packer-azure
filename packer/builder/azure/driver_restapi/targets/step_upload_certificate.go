// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package targets

import (
	"fmt"

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

	certData := []byte(state.Get(constants.Certificate).(string))

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
