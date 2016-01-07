// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package powershell

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type PS4Driver struct {
	ExecPath string
}

func NewPS4Driver() (*PS4Driver, error) {
	appliesTo := "Applies to Windows 8.1, Windows PowerShell 4.0, Windows Server 2012 R2 only"

	// Check this is Windows
	if runtime.GOOS != "windows" {
		err := fmt.Errorf("%s", appliesTo)
		return nil, err
	}

	ps4Driver := &PS4Driver{ExecPath: "powershell"}

	if err := ps4Driver.Verify(); err != nil {
		return nil, err
	}

	log.Printf("ExecPath: %s", ps4Driver.ExecPath)

	return ps4Driver, nil
}

func (d *PS4Driver) Verify() error {

	if err := d.verifyPSVersion(); err != nil {
		return err
	}

	if err := d.verifyElevatedMode(); err != nil {
		return err
	}

	if err := d.setExecutionPolicy(); err != nil {
		return err
	}

	return nil
}

func (d *PS4Driver) verifyPSVersion() error {

	log.Printf("Enter method: %s", "verifyPSVersion")
	// check PS is available and is of proper version
	versionCmd := "$host.version.Major"
	cmd := exec.Command(d.ExecPath, versionCmd)

	cmdOut, err := cmd.Output()
	if err != nil {
		return err
	}

	versionOutput := strings.TrimSpace(string(cmdOut))
	log.Printf("%s output: %s", versionCmd, versionOutput)

	ver, err := strconv.ParseInt(versionOutput, 10, 32)

	if err != nil {
		return err
	}

	if ver < 4 {
		err := fmt.Errorf("%s", "Windows PowerShell version 4.0 or higher is expected")
		return err
	}

	return nil
}

func (d *PS4Driver) verifyElevatedMode() error {

	log.Printf("Enter method: %s", "verifyElevatedMode")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {function foo(){try{")
	blockBuffer.WriteString("$myWindowsID=[System.Security.Principal.WindowsIdentity]::GetCurrent();")
	blockBuffer.WriteString("$myWindowsPrincipal=new-object System.Security.Principal.WindowsPrincipal($myWindowsID);")
	blockBuffer.WriteString("$adminRole=[System.Security.Principal.WindowsBuiltInRole]::Administrator;")
	blockBuffer.WriteString("if($myWindowsPrincipal.IsInRole($adminRole)){return $true}else{return $false}")
	blockBuffer.WriteString("}catch{return $false}} foo}")

	log.Printf(" blockBuffer: %s", blockBuffer.String())
	cmd := exec.Command(d.ExecPath, blockBuffer.String())

	cmdOut, err := cmd.Output()
	if err != nil {
		return err
	}

	res := strings.TrimSpace(string(cmdOut))
	log.Printf("cmdOut: " + string(res))

	if res == "False" {
		err := fmt.Errorf("%s", "Please restart your shell in elevated mode")
		return err
	}

	return nil
}

func (d *PS4Driver) setExecutionPolicy() error {

	log.Printf("Enter method: %s", "setExecutionPolicy")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Set-ExecutionPolicy RemoteSigned -Force}")

	err := d.Exec(blockBuffer.String())

	return err
}

func (d *PS4Driver) VerifyPSAzureModule() error {
	log.Printf("Enter method: %s", "VerifyPSAzureModule")

	versionCmd := "Invoke-Command -scriptblock { function foo(){try{ $commands = Get-Command -Module Azure;if($commands.Length -eq 0){return $false} }catch{return $false}; return $true} foo}"
	cmd := exec.Command(d.ExecPath, versionCmd)

	cmdOut, err := cmd.Output()
	if err != nil {
		return err
	}

	res := strings.TrimSpace(string(cmdOut))

	if res == "False" {
		err := fmt.Errorf("%s", "Azure PowerShell not found. Try this link to install Azure PowerShell: http://go.microsoft.com/?linkid=9811175&clcid=0x409 and restart the shell.")
		return err
	}

	return nil
}

func (d *PS4Driver) Exec(block string) error {

	log.Printf("Executing: %#v", block)

	var stdout, stderr bytes.Buffer

	script := exec.Command(d.ExecPath, block)
	script.Stdout = &stdout
	script.Stderr = &stderr

	err := script.Run()

	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("Exec error: %s", stderrString)
	}

	if len(stderrString) > 0 {
		err = fmt.Errorf("%s", stderrString)
	}

	stdoutString := strings.TrimSpace(stdout.String())

	log.Printf("Exec stdout: %s", stdoutString)
	log.Printf("Exec stderr: %s", stderrString)

	return err
}

func (d *PS4Driver) ExecRet(block string) (string, error) {

	log.Printf("Executing ExecRet: %#v", block)

	var stdout, stderr bytes.Buffer

	script := exec.Command(d.ExecPath, block)
	script.Stdout = &stdout
	script.Stderr = &stderr

	err := script.Run()

	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("ExecRet error: %s", stderrString)
	}

	if len(stderrString) > 0 {
		err = fmt.Errorf("%s", stderrString)
	}

	stdoutString := strings.TrimSpace(stdout.String())

	log.Printf("ExecRet stdout: %s", stdoutString)
	log.Printf("ExecRet stderr: %s", stderrString)

	return stdoutString, err
}

func (d *PS4Driver) ExecRemote(cmd *packer.RemoteCmd) error {

	log.Printf("Executing ExecRemote: %#v", cmd.Command)

	script := exec.Command(d.ExecPath, cmd.Command)
	script.Stdout = cmd.Stdout
	script.Stderr = cmd.Stderr

	err := script.Run()

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("ExecRemote error: %s", err)
	}

	return err
}
