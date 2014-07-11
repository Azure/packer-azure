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

type StepRemoveService struct {
	tmpServiceName string
}

func (s *StepRemoveService) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Removing Temporary Azure Service: %s"

	ui.Say("Removing Azure Temporary Service...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$tmpServiceName = '" + s.tmpServiceName + "';")
	blockBuffer.WriteString("Remove-AzureService -ServiceName $tmpServiceName -Force;")
	blockBuffer.WriteString("}")

	err := driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("srvExists", 0)

	return multistep.ActionContinue
}

func (s *StepRemoveService) Cleanup(state multistep.StateBag) {
	// do nothing
}
