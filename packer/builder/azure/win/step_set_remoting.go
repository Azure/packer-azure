// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package win

import (
	"fmt"
//	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/MSOpenTech/packer-azure/packer/communicator/powershell"
	msbldcommon "github.com/MSOpenTech/packer-azure/packer/builder/common"
)

type StepSetRemoting struct {
	comm packer.Communicator
	Username string
	Password string
}

func (s *StepSetRemoting) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	azureVmAddr := state.Get("azureVmAddr").(string)

	errorMsg := "Error StepRemoteSession: %s"

//	var blockBuffer bytes.Buffer
//	blockBuffer.WriteString("Invoke-Command -scriptblock {")
//	blockBuffer.WriteString("$uri = '" + azureVmAddr + "';")
//	blockBuffer.WriteString("$username = '" + s.Username + "';")
//	blockBuffer.WriteString("$password = '" + s.Password + "';")
//	blockBuffer.WriteString("$secPassword = ConvertTo-SecureString $password -AsPlainText -Force;")
//	blockBuffer.WriteString("$credential = New-Object -typename System.Management.Automation.PSCredential -argumentlist $username, $secPassword;")
//	blockBuffer.WriteString("$sess = New-PSSession -ConnectionUri $uri -Credential $credential;")
//	blockBuffer.WriteString("}")
//
//	err := driver.Exec( blockBuffer.String() )
//
//	if err != nil {
//		err := fmt.Errorf(errorMsg, err)
//		state.Put("error", err)
//		ui.Error(err.Error())
//		return multistep.ActionHalt
//	}

	comm, err := powershell.New(
		&powershell.Config{
			Driver: driver,
			Username: s.Username,
			Password: s.Password,
			RemoteHostUrl: azureVmAddr,
			Ui: ui,
		})

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.comm = comm
	state.Put("communicator", comm)

	return multistep.ActionContinue
}

func (s *StepSetRemoting) Cleanup(state multistep.StateBag) {
	// do nothing
}
