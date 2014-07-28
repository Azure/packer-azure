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
	osType string
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
	if s.osType == Linux {
		blockBuffer.WriteString("$ep = Get-AzureVM –ServiceName $tmpServiceName –Name $tmpVmName | Get-AzureEndpoint;")
		blockBuffer.WriteString("[string]::Format(\"{0}:{1}\", $ep.Vip, $ep.Port)")
	} else if  s.osType == Windows {
		blockBuffer.WriteString("$uri = Get-AzureWinRMUri -ServiceName $tmpServiceName -Name $tmpVmName;")
		blockBuffer.WriteString("$uri.AbsoluteUri;")
	} else {
		err := fmt.Errorf(errorMsg, "Unknown osType")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	blockBuffer.WriteString("}")

	var res string
	var err error

	res, err = driver.ExecRet( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if(len(res) == 0 ){
		err := fmt.Errorf(errorMsg, "stdout is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt

	}

	state.Put("azureVmAddr", res)

	return multistep.ActionContinue
}

func (s *StepGetEndpoint) Cleanup(state multistep.StateBag) {
	// do nothing
}
