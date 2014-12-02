// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package win

import (
	"fmt"
	ps "github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_powershell/driver"
	"github.com/MSOpenTech/packer-azure/packer/communicator/powershell"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepSetRemoting struct {
	comm     packer.Communicator
	Username string
	Password string
}

func (s *StepSetRemoting) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(ps.Driver)
	ui := state.Get("ui").(packer.Ui)
	azureVmAddr := state.Get("azureVmAddr").(string)

	errorMsg := "Error StepRemoteSession: %s"

	comm, err := powershell.New(
		&powershell.Config{
			Driver:        driver,
			Username:      s.Username,
			Password:      s.Password,
			RemoteHostUrl: azureVmAddr,
			Ui:            ui,
		})

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	packerCommunicator := packer.Communicator(comm)

	s.comm = packerCommunicator
	state.Put("communicator", packerCommunicator)

	return multistep.ActionContinue
}

func (s *StepSetRemoting) Cleanup(state multistep.StateBag) {
	// do nothing
}
