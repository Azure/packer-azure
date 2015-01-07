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
)

type StepGetEndpoint struct {
	OsType         string
	TmpVmName      string
	TmpServiceName string
}

func (s *StepGetEndpoint) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Getting Endpoint: %s"

	ui.Say("Getting Endpoint...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
	if s.OsType == Linux {
		blockBuffer.WriteString("$ep = Get-AzureVM –ServiceName $tmpServiceName –Name $tmpVmName | Get-AzureEndpoint;")
		blockBuffer.WriteString("[string]::Format(\"{0}:{1}\", $ep.Vip, $ep.Port)")
	} else if s.OsType == Windows {
		blockBuffer.WriteString("$uri = Get-AzureWinRMUri -ServiceName $tmpServiceName -Name $tmpVmName;")
		blockBuffer.WriteString("$uri.AbsoluteUri;")
	} else {
		err := fmt.Errorf(errorMsg, "Unknown OsType")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	blockBuffer.WriteString("}")

	var res string
	var err error

	res, err = driver.ExecRet(blockBuffer.String())

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(res) == 0 {
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
