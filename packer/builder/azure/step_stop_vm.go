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

type StepStopVm struct {
	tmpVmName string
	tmpServiceName string
}

func (s *StepStopVm) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Stopping Temporary Azure VM: %s"

	ui.Say("Stopping Temporary Azure VM...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$tmpVmName = '" + s.tmpVmName + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.tmpServiceName + "';")
	blockBuffer.WriteString("Stop-AzureVM -ServiceName $tmpServiceName -Name $tmpVmName -Force;")
	blockBuffer.WriteString("}")

	err := driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vmRunning", 0)

	return multistep.ActionContinue
}

func (s *StepStopVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
