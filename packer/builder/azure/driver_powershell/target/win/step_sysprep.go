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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

type StepSysprep struct {
	OsImageLabel string
}

func (s *StepSysprep) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)

	var err error
	errorMsg := "Error step Sysprep: %s"

	pattern := "Windows Server 2008"
	value := s.OsImageLabel

	match, _ := regexp.MatchString(pattern, value)
	is2008 := match

	ui.Say("Executing sysprep...")

	// could be problems with Windows Server 2008 R2 - sysprep sometimes drops remote connection
	// this is a workaround
	if is2008 {

		// create a temp file
		tempDir := os.TempDir()
		packerTempDir, err := ioutil.TempDir(tempDir, "pkrsp")
		if err != nil {
			err := fmt.Errorf("Error creating temporary directory: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		f, err := os.Create(filepath.Join(packerTempDir, "sysprep.ps1"))
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// add content
		f.WriteString("Write-Host 'Executing Sysprep from File...';\n")
		_, err = f.WriteString("Start-Process \"$env:windir\\System32\\Sysprep\\sysprep.exe\" -Wait -Argument '/quiet /generalize /oobe /quit';\n")
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		f.Close()

		log.Printf("Temp file: %s", f.Name())

		// execute the file remotely

		var stdoutBuff bytes.Buffer
		var stderrBuff bytes.Buffer
		var cmd packer.RemoteCmd
		cmd.Stdout = &stdoutBuff
		cmd.Stderr = &stderrBuff

		cmd.Command = "-filepath " + filepath.FromSlash(f.Name())

		err = comm.Start(&cmd)

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		stderrString := stderrBuff.String()
		log.Printf("StepSysprep stderr: %s", stderrString)

		if len(stderrString) > 0 {
			err = fmt.Errorf(errorMsg, stderrString)
			log.Printf(errorMsg, stderrString)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		stdoutString := stdoutBuff.String()
		if len(stdoutString) > 0 {
			log.Printf("Provision from file stdout: %s", stdoutString)
			ui.Message(stdoutString)
		}

		// remove the file
		os.RemoveAll(packerTempDir)

		return multistep.ActionContinue
	}

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("{")
	blockBuffer.WriteString("Start-Process \"$env:windir\\System32\\Sysprep\\sysprep.exe\" -Wait -Argument '/quiet /generalize /oobe /quit'")
	//	blockBuffer.WriteString("Start-Process 'cmd' -NoNewWindow -Wait -Argument '/k c:\\Windows\\System32\\Sysprep\\sysprep.exe /quiet /generalize /oobe /quit'")
	blockBuffer.WriteString("}")

	var stdoutBuff bytes.Buffer
	var stderrBuff bytes.Buffer
	var cmd packer.RemoteCmd
	cmd.Stdout = &stdoutBuff
	cmd.Stderr = &stderrBuff

	cmd.Command = "-ScriptBlock " + blockBuffer.String()

	err = comm.Start(&cmd)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	stderrString := stderrBuff.String()
	log.Printf("StepSysprep stderr: %s", stderrString)

	if len(stderrString) > 0 {
		err = fmt.Errorf(errorMsg, stderrString)
		log.Printf(errorMsg, stderrString)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	//	sleepTime := time.Minute * 5
	//	time.Sleep(sleepTime)

	return multistep.ActionContinue
}

func (s *StepSysprep) Cleanup(state multistep.StateBag) {
	// do nothing
}
