// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package powershell

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os"
	"bytes"
	"log"
	"path/filepath"
)

const DistrDstPathDefault = "C:/PackerDistr"

type config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The local path of the script.
	ScriptPath string `mapstructure:"script_path"`
	DistrSrcPath string `mapstructure:"distr_src_path"`
	DistrDstPath string `mapstructure:"distr_dst_dir_path"`
	Inline []string		`mapstructure:"inline"`
	tpl *packer.ConfigTemplate
}

type Provisioner struct {
	config config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	if(p.config.DistrDstPath == "" ){
		p.config.DistrDstPath = DistrDstPathDefault
	}

	sliceTemplates := map[string][]string{
		"inline":           p.config.Inline,
	}

	for n, slice := range sliceTemplates {
		for i, elem := range slice {
			var err error
			slice[i], err = p.config.tpl.Process(elem, nil)
			if err != nil {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Error processing %s[%d]: %s", n, i, err))
			}
		}
	}

	log.Println(fmt.Sprintf("%s: %v","inline", p.config.ScriptPath))


	templates := map[string]*string{
		"script_path":      &p.config.ScriptPath,
		"distr_src_path": 	&p.config.DistrSrcPath,
		"distr_dst_path": 	&p.config.DistrDstPath,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	log.Println(fmt.Sprintf("%s: %v","script_path", p.config.DistrDstPath))

	if len(p.config.ScriptPath) == 0 && p.config.Inline == nil {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Either a script file or inline script must be specified."))
	}

	if len(p.config.ScriptPath) != 0 {
		if _, err := os.Stat(p.config.ScriptPath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("script_path: '%v' check the path is correct.", p.config.ScriptPath))
		}
	}
	log.Println(fmt.Sprintf("%s: %v","script_path", p.config.ScriptPath))

	if len(p.config.DistrSrcPath) != 0 {
		if _, err := os.Stat(p.config.DistrSrcPath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("distr_src_path: '%v' check the path is correct.", p.config.DistrSrcPath))
		}
	}
	log.Println(fmt.Sprintf("%s: %v","distr_src_path", p.config.DistrSrcPath))

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {

	var err error
	errorMsg := "Provision error: %s"

	ui.Say("Provision...")

	if len(p.config.DistrSrcPath) != 0 {
		err = comm.UploadDir(p.config.DistrDstPath, p.config.DistrSrcPath, nil)
		if err != nil {
			return err
		}
	}

	if p.config.Inline != nil {

		var blockBuffer bytes.Buffer
		blockBuffer.WriteString("{")
		for _, command := range p.config.Inline {
			blockBuffer.WriteString(command + ";")
		}
		blockBuffer.WriteString("}")

		var stdoutBuff bytes.Buffer
		var stderrBuff bytes.Buffer
		var cmd packer.RemoteCmd
		cmd.Stdout = &stdoutBuff;
		cmd.Stderr = &stderrBuff;

		cmd.Command = "-ScriptBlock " + blockBuffer.String()

		err = comm.Start(&cmd)
		if err != nil {
			err = fmt.Errorf(errorMsg, err)
		}

		stderrString := stderrBuff.String()
		if(len(stderrString)>0) {
			err = fmt.Errorf(errorMsg, stderrString)
			log.Printf("Provision Inline stderr: %s", stderrString)
		}

		stdoutString := stdoutBuff.String()
		if(len(stdoutString)>0) {
			log.Printf("Provision Inline stdout: %s", stdoutString)
			ui.Message(stdoutString)
		}
	}

	if len(p.config.ScriptPath) != 0 {
		var stdoutBuff bytes.Buffer
		var stderrBuff bytes.Buffer
		var cmd packer.RemoteCmd
		cmd.Stdout = &stdoutBuff;
		cmd.Stderr = &stderrBuff;
		cmd.Command = "-filepath " + filepath.FromSlash(p.config.ScriptPath)

		err = comm.Start(&cmd)
		if err != nil {
			err = fmt.Errorf(errorMsg, err)
		}

		stderrString := stderrBuff.String()
		if(len(stderrString)>0) {
			err = fmt.Errorf(errorMsg, stderrString)
			log.Printf("Provision from file stderr: %s", stderrString)
		}

		stdoutString := stdoutBuff.String()
		if(len(stdoutString)>0) {
			log.Printf("Provision from file stdout: %s", stdoutString)
			ui.Message(stdoutString)
		}
	}

	return err
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}
