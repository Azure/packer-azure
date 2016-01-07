// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package powershell

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"path/filepath"
)

const DistrDstPathDefault = "C:/PackerDistr"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The local path of the script.
	ScriptPath   string   `mapstructure:"script_path"`
	DistrSrcPath string   `mapstructure:"distr_src_path"`
	DistrDstPath string   `mapstructure:"distr_dst_dir_path"`
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

	if p.config.DistrDstPath == "" {
		p.config.DistrDstPath = DistrDstPathDefault
	}

	// Accumulate any errors
	var errs *packer.MultiError

	log.Println(fmt.Sprintf("%s: %v", "inline", p.config.Inline))
	log.Println(fmt.Sprintf("%s: %v", "script_path", p.config.ScriptPath))
	log.Println(fmt.Sprintf("%s: %v", "distr_src_path", p.config.DistrSrcPath))
	log.Println(fmt.Sprintf("%s: %v", "distr_dst_dir_path", p.config.DistrDstPath))

	if p.config.ScriptPath == "" && p.config.Inline == nil {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("one of script_path or inline must be specified"))
	}

	if p.config.ScriptPath != "" {
		if _, err := os.Stat(p.config.ScriptPath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("script_path is not a valid path: %v", err))
		}
	}

	if p.config.DistrSrcPath != "" {
		if _, err := os.Stat(p.config.DistrSrcPath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("distr_src_path is not a valid path: %v", err))
		}
	}

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
		cmd.Stdout = &stdoutBuff
		cmd.Stderr = &stderrBuff

		cmd.Command = "-ScriptBlock " + blockBuffer.String()

		err = comm.Start(&cmd)
		if err != nil {
			err = fmt.Errorf(errorMsg, err)
		}

		stderrString := stderrBuff.String()
		if len(stderrString) > 0 {
			err = fmt.Errorf(errorMsg, stderrString)
			log.Printf("Provision Inline stderr: %s", stderrString)
		}

		stdoutString := stdoutBuff.String()
		if len(stdoutString) > 0 {
			log.Printf("Provision Inline stdout: %s", stdoutString)
			ui.Message(stdoutString)
		}
	}

	if len(p.config.ScriptPath) != 0 {
		var stdoutBuff bytes.Buffer
		var stderrBuff bytes.Buffer
		var cmd packer.RemoteCmd
		cmd.Stdout = &stdoutBuff
		cmd.Stderr = &stderrBuff
		cmd.Command = "-filepath " + filepath.FromSlash(p.config.ScriptPath)

		err = comm.Start(&cmd)
		if err != nil {
			err = fmt.Errorf(errorMsg, err)
		}

		stderrString := stderrBuff.String()
		if len(stderrString) > 0 {
			err = fmt.Errorf(errorMsg, stderrString)
			log.Printf("Provision from file stderr: %s", stderrString)
		}

		stdoutString := stdoutBuff.String()
		if len(stdoutString) > 0 {
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
