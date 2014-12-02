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

type StepRemoveVm struct {
	TmpVmName      string
	TmpServiceName string
}

func (s *StepRemoveVm) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Removig Temporary Azure VM: %s"

	ui.Say("Removing Temporary Azure VM...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$tmpVmName = '" + s.TmpVmName + "';")
	blockBuffer.WriteString("$tmpServiceName = '" + s.TmpServiceName + "';")
	blockBuffer.WriteString("Remove-AzureVM -ServiceName $tmpServiceName -Name $tmpVmName;")
	blockBuffer.WriteString("}")

	err := driver.Exec(blockBuffer.String())

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vmExists", 0)

	return multistep.ActionContinue
}

func (s *StepRemoveVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
