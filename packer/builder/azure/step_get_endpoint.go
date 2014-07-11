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
)

type StepGetEndpoint struct {
	tmpVmName string
	tmpServiceName string
}

func (s *StepGetEndpoint) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Getting Endpoint: %s"

	ui.Say("Getting Endpoint...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$tmpVmName = '" + s.tmpVmName + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.tmpServiceName + "';")
	blockBuffer.WriteString("$endpoint = Get-AzureVM –ServiceName $tmpServiceName –Name $tmpVmName | Get-AzureEndpoint;")
	blockBuffer.WriteString("$endpoint.Port;")
	blockBuffer.WriteString("}")

	var port string
	var err error

	port, err = driver.ExecRet( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("port", port)

	return multistep.ActionContinue
}

func (s *StepGetEndpoint) Cleanup(state multistep.StateBag) {
	// do nothing
}
