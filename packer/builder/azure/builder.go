// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package azure

import (
	"errors"
	"fmt"
	"log"
	"bytes"

	"github.com/mitchellh/multistep"
	msbldcommon "github.com/MSOpenTech/packer-azure/packer/builder/common"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/win"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/lin"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"math/rand"
	"time"
	"code.google.com/p/go-uuid/uuid"
	"path/filepath"
	"regexp"
)

// Instance size
const (
	ExtraSmall string 	= "ExtraSmall"
	Small string 		= "Small"
	Medium string 		= "Medium"
	Large string 		= "Large"
	ExtraLarge string 	= "ExtraLarge"
	A5 string 			= "A5"
	A6 string 			= "A6"
	A7 string 			= "A7"
	A8 string 			= "A8"
	A9 string 			= "A9"
)

// Instance type
const (
	Linux string = "Linux"
	Windows string = "Windows"
)

// Builder implements packer.Builder and builds the actual Azure
// images.
type Builder struct {
	config azure_config
	runner multistep.Runner
}

type azure_config struct {
	SubscriptionName        string     	`mapstructure:"subscription_name"`
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
	log.Println(fmt.Sprintf("%s: %v","subscription_name", b.config.SubscriptionName))

	if b.config.StorageAccount == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("storage_account: The option can't be missed."))
	}
	log.Println(fmt.Sprintf("%s: %v","storage_account", b.config.StorageAccount))

	osTypeIsValid := false
	osTypeArr := []string{
		Linux,
		Windows,
	}

	log.Println(fmt.Sprintf("%s: %v","instance_size", b.config.OsType))

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

	if b.config.OsImageLabel == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("os_image_label: The option can't be missed."))
	}
	log.Println(fmt.Sprintf("%s: %v","os_image_label", b.config.OsImageLabel))

	if b.config.Location == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("location: The option can't be missed."))
	}
	log.Println(fmt.Sprintf("%s: %v","location", b.config.Location))

	sizeIsValid := false
	instanceSizeArr := []string{
		ExtraSmall,
		Small,
		Medium,
		Large,
		ExtraLarge,
		A5,
		A6,
		A7,
		A8,
		A9,
	}

	log.Println(fmt.Sprintf("%s: %v","instance_size", b.config.InstanceSize))

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

	if b.config.UserImageLabel == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("user_image_label: The option can't be missed."))
	}
	log.Println(fmt.Sprintf("%s: %v","user_image_label", b.config.UserImageLabel))

	b.config.userImageName = fmt.Sprintf("%s_%s", b.config.UserImageLabel, uuid.New())
	log.Println(fmt.Sprintf("%s: %v","user_image_name", b.config.userImageName))

	// Errors
	templates := map[string]*string{
		"user_image_name":  &b.config.UserImageLabel,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	log.Println(fmt.Sprintf("%s: %v","user_image_name", b.config.UserImageLabel))

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
	driver, err := msbldcommon.NewPS4Driver()
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
			&StepSelectSubscription {
				SubscriptionName: b.config.SubscriptionName,
				StorageAccount: b.config.StorageAccount,
			},
			&lin.StepCreateCert {
				CertFileName: certFileName,
				KeyFileName: keyFileName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&StepCreateService {
				location: b.config.Location,
				tmpServiceName: b.config.tmpServiceName,
				storageAccount: b.config.StorageAccount,
				tmpVmName: b.config.tmpVmName,
			},
			&StepUploadCertificate {
				certFileName: filepath.Join(state.Get("certTempDir").(string), certFileName),
				tmpServiceName: b.config.tmpServiceName,
				username: b.config.username,
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
			&StepGetEndpoint {
				osType: b.config.OsType,
				tmpVmName: b.config.tmpVmName,
				tmpServiceName: b.config.tmpServiceName,
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
			&StepStopVm {
				tmpVmName: b.config.tmpVmName,
				tmpServiceName: b.config.tmpServiceName,
			},
			&StepRemoveVm {
				tmpVmName: b.config.tmpVmName,
				tmpServiceName: b.config.tmpServiceName,
			},
			&StepRemoveService {
				tmpServiceName: b.config.tmpServiceName,
			},
			&StepRemoveDisk {
				storageAccount: b.config.StorageAccount,
				tmpVmName: b.config.tmpVmName,
			},

			&StepCreateImage {
				storageAccount: b.config.StorageAccount,
				tmpVmName: b.config.tmpVmName,
				userImageLabel: b.config.UserImageLabel,
				userImageName: b.config.userImageName,
				osType: b.config.OsType,
			},
		}
	} else if b.config.OsType == "Windows" {
//		b.config.tmpVmName = "PkrVM-95129190"
//		b.config.tmpServiceName = "PkrSrv-95129190"
		password := "Zxcv1234"
		steps = []multistep.Step {
			&StepSelectSubscription {
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
			&StepGetEndpoint {
				osType: b.config.OsType,
				tmpVmName: b.config.tmpVmName,
				tmpServiceName: b.config.tmpServiceName,
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

			&StepStopVm {
				tmpVmName: b.config.tmpVmName,
				tmpServiceName: b.config.tmpServiceName,
				},
			&StepRemoveVm {
				tmpVmName: b.config.tmpVmName,
				tmpServiceName: b.config.tmpServiceName,
				},
			&StepRemoveService {
				tmpServiceName: b.config.tmpServiceName,
				},
			&StepRemoveDisk {
				storageAccount: b.config.StorageAccount,
				tmpVmName: b.config.tmpVmName,
				},
			&StepCreateImage {
				storageAccount: b.config.StorageAccount,
				tmpVmName: b.config.tmpVmName,
				userImageLabel: b.config.UserImageLabel,
				userImageName: b.config.userImageName,
				osType: b.config.OsType,
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

func (b *Builder)validateAzureOptions(ui packer.Ui, driver msbldcommon.Driver) error {
	// check Azure subscription

	var blockBuffer bytes.Buffer
	var err error
	var res string

	ui.Say("Validating Azure options...")

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

	// check image
	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$osImageLabel = '" + b.config.OsImageLabel + "';")
	blockBuffer.WriteString("$location = '" + b.config.Location + "';")
	if b.config.OsType == Linux {
		blockBuffer.WriteString("$image = Get-AzureVMImage | where { ($_.Label -like $osImageLabel) -or ($_.ImageFamily -like $osImageLabel) } | where { $_.Location.Split(';') -contains $location} | Sort-Object -Descending -Property PublishedDate | Select -First 1;")
	} else if  b.config.OsType == Windows {
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

