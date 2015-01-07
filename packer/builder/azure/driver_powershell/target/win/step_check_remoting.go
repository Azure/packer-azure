// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package win

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"strings"
	"time"
)

type StepCheckRemoting struct {
}

func (s *StepCheckRemoting) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)

	var err error
	errorMsg := "Error step CheckRemoting: %s"

	magicWord := "Ready!"

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("{ Write-Host '" + magicWord + "' }")

	var stdoutBuff bytes.Buffer
	var stderrBuff bytes.Buffer
	var cmd packer.RemoteCmd
	cmd.Stdout = &stdoutBuff
	cmd.Stderr = &stderrBuff
	cmd.Command = "-ScriptBlock " + blockBuffer.String()

	count := 5
	var duration time.Duration = 1
	sleepTime := time.Minute * duration

	ui.Say("Checking Connection...")

	for count > 0 {
		stdoutBuff.Reset()
		err = comm.Start(&cmd)
		if err != nil {
			err := fmt.Errorf(errorMsg, "Remote connection failed")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		stderrString := stderrBuff.String()
		log.Printf("StepCheckRemoting stderr: %s", stderrString)

		if len(stderrString) > 0 {
			err = fmt.Errorf(errorMsg, stderrString)
			log.Printf(errorMsg, stderrString)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		stdoutString := strings.TrimSpace(stdoutBuff.String())
		log.Printf("StepCheckRemoting stdout: '%s'", stdoutString)
		if stdoutString == magicWord {
			ui.Say(stdoutString)
			break
		}

		log.Println(fmt.Sprintf("Waiting %v minutes for the remote connection to get ready...", uint(duration)))
		time.Sleep(sleepTime)
		count--
	}

	if count == 0 {
		err := fmt.Errorf(errorMsg, "Remote connection failed")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCheckRemoting) Cleanup(state multistep.StateBag) {
	// do nothing
}
