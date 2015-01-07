// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package targets

import (
	"encoding/base64"
	"fmt"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/constants"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/request"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
)

type StepUploadCertificate struct {
	TmpServiceName string
}

func (s *StepUploadCertificate) Run(state multistep.StateBag) multistep.StepAction {
	reqManager := state.Get(constants.RequestManager).(*request.Manager)
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
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	certDataBase64 := base64.StdEncoding.EncodeToString(certData)
	certFormat := "pfx"

	requestData := reqManager.AddCertificate(s.TmpServiceName, certDataBase64, certFormat, "")
	err = reqManager.ExecuteSync(requestData)

	if err != nil {
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
