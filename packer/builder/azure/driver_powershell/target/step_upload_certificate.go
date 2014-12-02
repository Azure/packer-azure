// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package target

import (
	"bytes"
	"fmt"
	ps "github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_powershell/driver"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
)

type StepUploadCertificate struct {
	CertFileName   string
	TmpServiceName string
	Username       string
}

func (s *StepUploadCertificate) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)

	var err error
	certTempDir := state.Get("certTempDir").(string)
	if certTempDir == "" {
		err = fmt.Errorf("certTempDir is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	errorMsg := "Error Uploading Temporary Certificate: %s"
	certPath := filepath.Join(certTempDir, s.CertFileName)

	ui.Say("Uploading Temporary Certificate...")

	var blockBuffer bytes.Buffer

	// check certificate - get Thumbprint
	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2;")
	blockBuffer.WriteString("$certPath = '" + certPath + "';")
	blockBuffer.WriteString("$cert.Import($certPath);")
	blockBuffer.WriteString("$cert.Thumbprint;")
	blockBuffer.WriteString("}")

	res, err := driver.ExecRet(blockBuffer.String())

	if err != nil {
		err = fmt.Errorf("Can't get certificate thumbprint '%s'", certPath)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if res == "" {
		err = fmt.Errorf("Can't get certificate thumbprint '%s'", certPath)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("certThumbprint", res)

	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$certPath = '" + certPath + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
	blockBuffer.WriteString("Add-AzureCertificate -serviceName $tmpServiceName -certToDeploy $certPath;")
	blockBuffer.WriteString("}")

	err = driver.Exec(blockBuffer.String())

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
