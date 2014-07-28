// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package win

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	msbldcommon "github.com/MSOpenTech/packer-azure/packer/builder/common"
)

type StepInstallCert struct {
	TmpVmName string
	TmpServiceName string
}

func (s *StepInstallCert) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Installing Temporary Certificate: %s"

	ui.Say("Installing Temporary Certificate...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
	blockBuffer.WriteString("$certThumbprint = (Get-AzureVM -ServiceName $tmpServiceName -Name $tmpVmName | select -ExpandProperty vm).DefaultWinRMCertificateThumbprint;")
	blockBuffer.WriteString("$certX509 = Get-AzureCertificate -ServiceName $tmpServiceName -Thumbprint $certThumbprint -ThumbprintAlgorithm sha1;")
	blockBuffer.WriteString("$certTempFile = [IO.Path]::GetTempFileName();")
	blockBuffer.WriteString("$certX509.Data | Out-File $certTempFile;")
	blockBuffer.WriteString("$certToImport = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2 $certTempFile;")
	blockBuffer.WriteString("$store = New-Object System.Security.Cryptography.X509Certificates.X509Store 'Root', 'LocalMachine';")
	blockBuffer.WriteString("$store.Open([System.Security.Cryptography.X509Certificates.OpenFlags]::ReadWrite);")
	blockBuffer.WriteString("$store.Add($certToImport);")
	blockBuffer.WriteString("$store.Close();")
	blockBuffer.WriteString("Remove-Item $certTempFile;")
	blockBuffer.WriteString("}")

	err := driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepInstallCert) Cleanup(state multistep.StateBag) {
	// TODO: uninstall cert
}
