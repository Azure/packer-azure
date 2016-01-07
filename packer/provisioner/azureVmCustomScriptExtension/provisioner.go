// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azureVmCustomScriptExtension

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/pborman/uuid"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The local path of the script.
	ScriptPath   string   `mapstructure:"script_path"`
	DistrSrcPath string   `mapstructure:"distr_src_path"`
	Inline       []string `mapstructure:"inline"`
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate: true,
	}, raws...)
	if err != nil {
		return err
	}

	// Defaults
	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	// Accumulate any errors
	var errs *packer.MultiError

	log.Println(fmt.Sprintf("%s: %v", "inline", p.config.Inline))
	log.Println(fmt.Sprintf("%s: %v", "script_path", p.config.ScriptPath))
	log.Println(fmt.Sprintf("%s: %v", "distr_src_path", p.config.DistrSrcPath))

	if p.config.ScriptPath == "" && p.config.Inline == nil {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("one of script_path or inline must be specified"))
	}

	if p.config.ScriptPath != "" {
		if _, err := os.Stat(p.config.ScriptPath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("script_path is not a valid path: %s", err))
		}
	}

	if p.config.DistrSrcPath != "" {
		if _, err := os.Stat(p.config.DistrSrcPath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("distr_src_path is not a valid path: %s", err))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {

	var err error
	errorMsg := "Error preparing shell script, %s error: %s"

	ui.Say("Provisioning...")

	if len(p.config.DistrSrcPath) != 0 {
		err = comm.UploadDir("skiped param", p.config.DistrSrcPath, nil)
		if err != nil {
			return err
		}
	}

	tempDir := os.TempDir()
	packerTempDir, err := ioutil.TempDir(tempDir, "packer_script")
	if err != nil {
		err := fmt.Errorf("Error creating temporary directory: %s", err.Error())
		return err
	}

	// create a temporary script file and upload it to Azure storage
	ui.Message("Preparing execution script...")
	provisionFileName := fmt.Sprintf("provision-%s.ps1", uuid.New())

	scriptPath := filepath.Join(packerTempDir, provisionFileName)
	tf, err := os.Create(scriptPath)

	if err != nil {
		return fmt.Errorf(errorMsg, "os.Create", err.Error())
	}

	defer os.RemoveAll(packerTempDir)

	writer := bufio.NewWriter(tf)

	if p.config.Inline != nil {
		log.Println("Writing inline commands to execution script...")

		// Write our contents to it
		for _, command := range p.config.Inline {
			log.Println(command)
			if _, err := writer.WriteString(command + ";\n"); err != nil {
				return fmt.Errorf(errorMsg, "writer.WriteString", err.Error())
			}
		}
	}

	// add content to the temp script file
	if len(p.config.ScriptPath) != 0 {
		log.Println("Writing file commands to execution script...")

		f, err := os.Open(p.config.ScriptPath)
		if err != nil {
			return fmt.Errorf(errorMsg, "os.Open", err.Error())
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			log.Println(scanner.Text())
			if _, err := writer.WriteString(scanner.Text() + "\n"); err != nil {
				return fmt.Errorf(errorMsg, "writer.WriteString", err.Error())
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf(errorMsg, "scanner.Scan", err.Error())
		}
	}

	log.Println("Writing SysPrep to execution script...")
	sysprepPs := []string{
		"Write-Host 'Executing Sysprep from File...'",
		"Start-Process $env:windir\\System32\\Sysprep\\sysprep.exe -NoNewWindow -Wait -Argument '/quiet /generalize /oobe /quit'",
		"Write-Host 'Sysprep is done!'",
	}

	for _, command := range sysprepPs {
		log.Println(command)
		if _, err := writer.WriteString(command + ";\n"); err != nil {
			return fmt.Errorf(errorMsg, "writer.WriteString", err.Error())
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Error preparing shell script: %s", err.Error())
	}

	tf.Close()

	// upload to Azure storage
	ui.Message("Uploading execution script...")
	err = comm.UploadDir("skiped param", scriptPath, nil)
	if err != nil {
		return err
	}

	// execute script with Custom script extension
	runScript := provisionFileName

	var stdoutBuff bytes.Buffer
	var stderrBuff bytes.Buffer
	var cmd packer.RemoteCmd
	cmd.Stdout = &stdoutBuff
	cmd.Stderr = &stderrBuff

	cmd.Command = runScript

	ui.Message("Starting provisioning. It may take some time...")
	err = comm.Start(&cmd)
	if err != nil {
		err = fmt.Errorf(errorMsg, "comm.Start", err.Error())
		return err
	}

	ui.Message("Provision is Completed")

	stderrString := stderrBuff.String()
	if len(stderrString) > 0 {
		err = fmt.Errorf(errorMsg, "stderrString", stderrString)
		log.Printf("Provision stderr: %s", stderrString)
	}

	ui.Say("Script output")

	stdoutString := stdoutBuff.String()
	if len(stdoutString) > 0 {
		log.Printf("Provision stdout: %s", stdoutString)
		scriptMessages := strings.Split(stdoutString, "\\n")
		for _, m := range scriptMessages {
			ui.Message(m)
		}
	}

	return err
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}
