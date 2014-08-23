// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package driver_powershell

import (
	"errors"
	"fmt"
	"log"
	"bytes"

	"github.com/mitchellh/multistep"
	ps "github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_powershell/driver"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_powershell/target"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_powershell/target/win"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_powershell/target/lin"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"math/rand"
	"time"
	"code.google.com/p/go-uuid/uuid"
	"path/filepath"
	"regexp"
	"os"
)

// Builder implements packer.Builder and builds the actual Azure
// images.
type Builder struct {
	config azure_config
	runner multistep.Runner
}

type azure_config struct {
	SubscriptionName        string     	`mapstructure:"subscription_name"`
	PublishSettingsPath     string     	`mapstructure:"publish_settings_path"`
	StorageAccount          string     	`mapstructure:"storage_account"`
	OsType         			string   	`mapstructure:"os_type"`
	OsImageLabel         	string   	`mapstructure:"os_image_label"`
	Location 				string 		`mapstructure:"location"`
	InstanceSize 			string		`mapstructure:"instance_size"`
	UserImageLabel 			string		`mapstructure:"user_image_label"`
	common.PackerConfig           		`mapstructure:",squash"`
	tpl *packer.ConfigTemplate

	username          		string		`mapstructure:"username"`
	tmpVmName              	string
	tmpServiceName          string
	userImageName          	string

}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {

	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("%s: %v", "PackerUserVars", b.config.PackerUserVars))

	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors and warnings
	errs := common.CheckUnusedConfig(md)
	warnings := make([]string, 0)

	if b.config.SubscriptionName == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("subscription_name: The option can't be missed."))
	}

	if b.config.StorageAccount == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("storage_account: The option can't be missed."))
	}

	if b.config.OsImageLabel == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("os_image_label: The option can't be missed."))
	}

	if b.config.Location == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("location: The option can't be missed."))
	}
	log.Println(fmt.Sprintf("%s: %v","location", b.config.Location))


	if b.config.UserImageLabel == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("user_image_label: The option can't be missed."))
	}

	templates := map[string]*string{
		"subscription_name":  &b.config.SubscriptionName,
		"publish_settings_path":  &b.config.PublishSettingsPath,
		"storage_account":  &b.config.StorageAccount,
		"os_type":  &b.config.OsType,
		"os_image_label":  &b.config.OsImageLabel,
		"location":  &b.config.Location,
		"instance_size":  &b.config.InstanceSize,
		"user_image_label":  &b.config.UserImageLabel,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	log.Println(fmt.Sprintf("%s: %v","subscription_name", b.config.SubscriptionName))

	if len(b.config.PublishSettingsPath) > 0 {
		if _, err := os.Stat(b.config.PublishSettingsPath); err != nil {
			errs = packer.MultiErrorAppend(errs, errors.New("publish_settings_path: Check the path is correct."))
		}

		log.Println(fmt.Sprintf("%s: %v","publish_settings_path", b.config.PublishSettingsPath))
	}

	log.Println(fmt.Sprintf("%s: %v","storage_account", b.config.StorageAccount))

	osTypeIsValid := false
	osTypeArr := []string{
		target.Linux,
		target.Windows,
	}

	for _, osType := range osTypeArr {
		if b.config.OsType == osType {
			osTypeIsValid = true
			break
		}
	}

	if !osTypeIsValid {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("os_type: The value is invalid. Must be one of: %v", osTypeArr))
	}

	log.Println(fmt.Sprintf("%s: %v","os_type", b.config.OsType))


	log.Println(fmt.Sprintf("%s: %v","os_image_label", b.config.OsImageLabel))

	sizeIsValid := false
	instanceSizeArr := []string{
		target.ExtraSmall,
		target.Small,
		target.Medium,
		target.Large,
		target.ExtraLarge,
		target.A5,
		target.A6,
		target.A7,
		target.A8,
		target.A9,
	}


	for _, instanceSize := range instanceSizeArr {
		if b.config.InstanceSize == instanceSize {
			sizeIsValid = true
			break
		}
	}

	if !sizeIsValid {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("instance_size: The value is invalid. Must be one of: %v", instanceSizeArr))
	}

	log.Println(fmt.Sprintf("%s: %v","instance_size", b.config.InstanceSize))


	log.Println(fmt.Sprintf("%s: %v","user_image_label", b.config.UserImageLabel))
	b.config.userImageName = fmt.Sprintf("%s_%s", b.config.UserImageLabel, uuid.New())
	log.Println(fmt.Sprintf("%s: %v","user_image_name", b.config.userImageName))

	// random symbols for vm name (should be unique)
	// for Win  - the computer name cannot be more than 15 characters long
	rand.Seed(time.Now().Unix())
	availSymb := "0123456789abcdefghijklmnopqrstuvwxyz"
	availSymbLen := len(availSymb)
	const tmpServiceNamePrefix = "PkrSrv"
	const tmpVmNamePrefix = "PkrVM"
	const allowedVmNameLength = 15
	genLen := allowedVmNameLength - len(tmpVmNamePrefix)
	var rnd string
	for i := 0; i < genLen; i++ {
		rnd += string(availSymb[rand.Intn(availSymbLen)])
	}

	if b.config.tmpVmName == "" {
		b.config.tmpVmName = fmt.Sprintf("%s%s", tmpVmNamePrefix, rnd)
	}
	log.Println(fmt.Sprintf("%s: %v","tmpVmName", b.config.tmpVmName))

	if b.config.tmpServiceName == "" {
		b.config.tmpServiceName = fmt.Sprintf("%s%s", tmpServiceNamePrefix, rnd)
	}
	log.Println(fmt.Sprintf("%s: %v","tmpServiceName", b.config.tmpServiceName))

	if b.config.username == "" {
		b.config.username = fmt.Sprintf("%s", "packer")
	}
	log.Println(fmt.Sprintf("%s: %v","username", b.config.username))

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

// Run executes a Packer build and returns a packer.Artifact representing
// a PS Azure appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with PS Azure
	driver, err := ps.NewPS4Driver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating PowerShell driver: %s", err)
	}

	err = driver.VerifyPSAzureModule()
	if err != nil {
		return nil, fmt.Errorf("Azure PowerShell module verifivation failed: %s", err)
	}

	err = b.validateAzureOptions(ui, driver)
	if err != nil {
		return nil, fmt.Errorf("Some Azure options failed: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("tmpServiceName", b.config.tmpServiceName)
	state.Put("certTempDir", "")

	// complete flags
	state.Put("srvExists", 0)
	state.Put("certUploaded", 0)
	state.Put("vmExists", 0)
	state.Put("diskExists", 0)
	state.Put("vmRunning", 0)
	state.Put("imageCreated", 0)

	var steps []multistep.Step

	if b.config.OsType == "Linux" {
		certFileName:= "cert.pem"
		keyFileName := "key.pem"

		steps = []multistep.Step{
			&target.StepSelectSubscription {
				SubscriptionName: b.config.SubscriptionName,
				StorageAccount: b.config.StorageAccount,
			},
			&lin.StepCreateCert {
				CertFileName: certFileName,
				KeyFileName: keyFileName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&target.StepCreateService {
				Location: b.config.Location,
				TmpServiceName: b.config.tmpServiceName,
				StorageAccount: b.config.StorageAccount,
				TmpVmName: b.config.tmpVmName,
			},
			&target.StepUploadCertificate {
				CertFileName: filepath.Join(state.Get("certTempDir").(string), certFileName),
				TmpServiceName: b.config.tmpServiceName,
				Username: b.config.username,
			},
			&lin.StepCreateVm {
				OsType: b.config.OsType,
				StorageAccount: b.config.StorageAccount,
				OsImageLabel: b.config.OsImageLabel,
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
				InstanceSize: b.config.InstanceSize,
				Username: b.config.username,
			},
			&target.StepGetEndpoint {
				OsType: b.config.OsType,
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&common.StepConnectSSH {
				SSHAddress:     lin.SSHAddress,
				SSHConfig:      lin.SSHConfig(b.config.username),
				SSHWaitTimeout: 20*time.Minute,
			},
			&common.StepProvision {},
			&lin.StepGeneralizeOs{
				Command: "sudo /usr/sbin/waagent -force -deprovision && export HISTSIZE=0",
			},
			&target.StepStopVm {
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&target.StepRemoveVm {
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&target.StepRemoveService {
				TmpServiceName: b.config.tmpServiceName,
			},
			&target.StepRemoveDisk {
				StorageAccount: b.config.StorageAccount,
				TmpVmName: b.config.tmpVmName,
			},

			&target.StepCreateImage {
				StorageAccount: b.config.StorageAccount,
				TmpVmName: b.config.tmpVmName,
				UserImageLabel: b.config.UserImageLabel,
				UserImageName: b.config.userImageName,
				OsType: b.config.OsType,
			},
		}
	} else if b.config.OsType == "Windows" {
//		b.config.tmpVmName = "PkrVM-95129190"
//		b.config.tmpServiceName = "PkrSrv-95129190"
		password := "Zxcv1234"
		steps = []multistep.Step {
			&target.StepSelectSubscription {
				SubscriptionName: b.config.SubscriptionName,
				StorageAccount: b.config.StorageAccount,
				},

			&win.StepCreateVm {
				OsType: b.config.OsType,
				StorageAccount: b.config.StorageAccount,
				OsImageLabel: b.config.OsImageLabel,
				Location: b.config.Location,
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
				InstanceSize: b.config.InstanceSize,
				Username: b.config.username,
				Password: password,
				},

			&win.StepInstallCert {
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
				},
			&target.StepGetEndpoint {
				OsType: b.config.OsType,
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
				},
			&win.StepSetRemoting {
				Username: b.config.username,
				Password: password,
				},
			new(win.StepCheckRemoting),
			&common.StepProvision{},
			&win.StepSysprep{
				OsImageLabel: b.config.OsImageLabel,
			},

			&target.StepStopVm {
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
				},
			&target.StepRemoveVm {
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
				},
			&target.StepRemoveService {
				TmpServiceName: b.config.tmpServiceName,
				},
			&target.StepRemoveDisk {
				StorageAccount: b.config.StorageAccount,
				TmpVmName: b.config.tmpVmName,
				},
			&target.StepCreateImage {
				StorageAccount: b.config.StorageAccount,
				TmpVmName: b.config.tmpVmName,
				UserImageLabel: b.config.UserImageLabel,
				UserImageName: b.config.userImageName,
				OsType: b.config.OsType,
				},
		}

	} else {
		return nil, fmt.Errorf("Unkonwn OS type: %s", b.config.OsType)
	}

	// Run the steps.
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}
	b.runner.Run(state)

	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	return &artifact {
		imageLabel: b.config.UserImageLabel,
		imageName: b.config.userImageName,
		mediaLocation: fmt.Sprintf("https://%s.blob.core.windows.net/vhds/%s.vhd", b.config.StorageAccount,  b.config.tmpVmName),
		}, nil
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func (b *Builder)validateAzureOptions(ui packer.Ui, driver ps.Driver) error {
	ui.Say("Validating Azure options...")

	var blockBuffer bytes.Buffer
	var err error
	var res string

	// check Azure subscription

	if len(b.config.PublishSettingsPath) > 0 { // use PublishSettings file
		log.Printf("Importing PublishSettings file '%s'", b.config.PublishSettingsPath)
		blockBuffer.Reset()
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$psPath = '" + b.config.PublishSettingsPath + "';")
		blockBuffer.WriteString("Import-AzurePublishSettingsFile $psPath;")
		blockBuffer.WriteString("}")

		res, err = driver.ExecRet( blockBuffer.String() )

		if err != nil {
			return err
		}

		blockBuffer.Reset()
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$subscriptionName = '" + b.config.SubscriptionName + "';")
		blockBuffer.WriteString("$s = Get-AzureSubscription | ? {$_.SubscriptionName -eq $subscriptionName};")
		blockBuffer.WriteString("$s.Count -eq 1;")
		blockBuffer.WriteString("}")

		res, err = driver.ExecRet( blockBuffer.String() )

		if err != nil {
			return err
		}

		if(res == "False"){
			err = fmt.Errorf("Can't find subscription '%s'", b.config.SubscriptionName)
			return err
		}

	} else { // use AAD
		blockBuffer.Reset()
		blockBuffer.WriteString("Invoke-Command -scriptblock {")
		blockBuffer.WriteString("$subscriptionName = '" + b.config.SubscriptionName + "';")
		blockBuffer.WriteString("$s = Get-AzureSubscription | ? {$_.SubscriptionName -eq $subscriptionName};")
		blockBuffer.WriteString("$s.Count -eq 1;")
		blockBuffer.WriteString("}")

		res, err = driver.ExecRet( blockBuffer.String() )

		if err != nil {
			return err
		}

		if(res == "False"){

			blockBuffer.Reset()
			blockBuffer.WriteString("Invoke-Command -scriptblock {")
			blockBuffer.WriteString("Add-AzureAccount")
			blockBuffer.WriteString("}")

			err = driver.Exec( blockBuffer.String() )

			if err != nil {
				return err
			}

			blockBuffer.Reset()
			blockBuffer.WriteString("Invoke-Command -scriptblock {")
			blockBuffer.WriteString("$subscriptionName = '" + b.config.SubscriptionName + "';")
			blockBuffer.WriteString("$s = Get-AzureSubscription | ? {$_.SubscriptionName -eq $subscriptionName};")
			blockBuffer.WriteString("$s.Count -eq 1;")
			blockBuffer.WriteString("}")

			res, err = driver.ExecRet( blockBuffer.String() )

			if err != nil {
				return err
			}

			if(res == "False"){
				err = fmt.Errorf("Can't find subscription '%s'", b.config.SubscriptionName)
				return err
			}
		}
	}


	// check Storage account
	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$storageAccount = '" + b.config.StorageAccount + "';")
	blockBuffer.WriteString("Test-AzureName -Storage -Name $storageAccount;")
	blockBuffer.WriteString("}")

	res, err = driver.ExecRet(blockBuffer.String())

	if err != nil {
		// Sometimes Test-AzureName cmdlet returns this error (bellow)
		pattern := "Your Windows Azure credential in the Windows PowerShell session has expired"
		value := err.Error()

		match, _ := regexp.MatchString(pattern, value)
		if match {
			// Renew subscription if so
			blockBuffer.Reset()
			blockBuffer.WriteString("Invoke-Command -scriptblock {")
			blockBuffer.WriteString("Add-AzureAccount")
			blockBuffer.WriteString("}")

			err = driver.Exec( blockBuffer.String() )

			if err != nil {
				return err
			}

			blockBuffer.Reset()
			blockBuffer.WriteString("Invoke-Command -scriptblock {")
			blockBuffer.WriteString("$storageAccount = '" + b.config.StorageAccount + "';")
			blockBuffer.WriteString("Test-AzureName -Storage -Name $storageAccount;")
			blockBuffer.WriteString("}")

			res, err = driver.ExecRet(blockBuffer.String())
			if err != nil {
				return err
			}
		} else {
			return err
		}

	}

	if(res == "False"){
		err = fmt.Errorf("Can't find storage account '%s'", b.config.StorageAccount)
		return err
	}

	// check os image
	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$osImageLabel = '" + b.config.OsImageLabel + "';")
	blockBuffer.WriteString("$location = '" + b.config.Location + "';")
	if b.config.OsType == target.Linux {
		blockBuffer.WriteString("$image = Get-AzureVMImage | where { ($_.Label -like $osImageLabel) -or ($_.ImageFamily -like $osImageLabel) } | where { $_.Location.Split(';') -contains $location} | Sort-Object -Descending -Property PublishedDate | Select -First 1;")
	} else if  b.config.OsType ==  target.Windows {
		blockBuffer.WriteString("$image = Get-AzureVMImage | where {  $_.Label -Match $osImageLabel } | where { $_.Location.Split(';') -contains $location} | Sort-Object -Descending -Property PublishedDate | Select -First 1;")
	} else {
		err := fmt.Errorf("Can't find OS image '%s' with OS type '%s'", b.config.OsImageLabel, b.config.OsType )
		return err
	}
	blockBuffer.WriteString("$image -ne $null;")
	blockBuffer.WriteString("}")

	res, err = driver.ExecRet( blockBuffer.String() )

	if err != nil {
		return err
	}

	if(res == "False"){
		err = fmt.Errorf("Can't find OS image '%s' located at '%s'", b.config.OsImageLabel, b.config.Location)
		return err
	}

	// TODO: check user imageName/label

	return nil
}

func appendWarnings(slice []string, data ...string) []string {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]string, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

