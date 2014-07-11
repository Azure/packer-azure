// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strings"
	"log"
)

type StepSysprep struct {
}

func (s *StepSysprep) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)

	var err error
	errorMsg := "Error step Sysprep: %s"

	ui.Say("Executing sysprep...")

	var cmd packer.RemoteCmd
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("{")
	blockBuffer.WriteString("Start-Process 'c:\\Windows\\System32\\Sysprep\\sysprep.exe' -NoNewWindow -Wait -Argument '/oobe /generalize /quiet /quit'")
	blockBuffer.WriteString("}")

	cmd.Command = "-ScriptBlock " + blockBuffer.String()
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = comm.Start(&cmd)

	stderrString := strings.TrimSpace(stderr.String())
	stdoutString := strings.TrimSpace(stdout.String())

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	if len(stderrString) > 0 {
		err = fmt.Errorf("Provision error: %s", stderrString)
	}

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepSysprep) Cleanup(state multistep.StateBag) {
	// do nothing
}
