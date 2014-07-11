// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package azure

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	msbldcommon "github.com/MSOpenTech/packer-azure/packer/builder/common"
	"path/filepath"
)

type StepUploadCertificate struct {
	certFileName string
	tmpServiceName string
	username string
}

func (s *StepUploadCertificate) Run(state multistep.StateBag) multistep.StepAction {
	certTempDir := state.Get("certTempDir").(string)
	if certTempDir == "" {
		return multistep.ActionContinue
	}
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Uploading Certificate: %s"
	certPath := filepath.Join(certTempDir, s.certFileName)

	var blockBuffer bytes.Buffer

	// check certificate - get Thumbprint
	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2;")
	blockBuffer.WriteString("$certPath = '" + certPath + "';")
	blockBuffer.WriteString("$cert.Import($certPath);")
	blockBuffer.WriteString("$cert.Thumbprint;")
	blockBuffer.WriteString("}")

	res, err := driver.ExecRet( blockBuffer.String() )

	if err != nil {
		err = fmt.Errorf("Can't get certificate thumbprint '%s'", certPath)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if(res == "") {
		err = fmt.Errorf("Can't get certificate thumbprint '%s'", certPath)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("certThumbprint", res)

	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$certPath = '" + certPath + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.tmpServiceName + "';")
	blockBuffer.WriteString("Add-AzureCertificate -serviceName $tmpServiceName -certToDeploy $certPath;")
	blockBuffer.WriteString("}")

	err = driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("certUploaded", 1)

	return multistep.ActionContinue
}

func (s *StepUploadCertificate) Cleanup(state multistep.StateBag) {
	// do nothing
}
