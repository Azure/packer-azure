// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package driver_restapi

import (
	"errors"
	"fmt"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/constants"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/request"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/targets"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/targets/lin"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/targets/win"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/utils"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

// Builder implements packer.Builder and builds the actual Azure
// images.
type Builder struct {
	config *Config
	runner multistep.Runner
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := newConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	return warnings, nil
}

// Run executes a Packer build and returns a packer.Artifact representing
// a PS Azure appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {

	var err error
	ui.Say("Preparing builder...")

	ui.Message("Creating a new request manager...")
	reqManager, err := request.NewManager(b.config.PublishSettingsPath, b.config.SubscriptionName)
	if err != nil {
		return nil, fmt.Errorf("Error creating request manager: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put(constants.Config, &b.config)
	state.Put(constants.RequestManager, reqManager)
	state.Put("hook", hook)
	state.Put(constants.Ui, ui)

	// complete flags
	state.Put(constants.CertCreated, 0)
	state.Put(constants.SrvExists, 0)
	state.Put(constants.CertInstalled, 0)
	state.Put(constants.CertUploaded, 0)
	state.Put(constants.VmExists, 0)
	state.Put(constants.DiskExists, 0)
	state.Put(constants.VmRunning, 0)
	state.Put(constants.ImageCreated, 0)

	ui.Say("Validating Azure Options...")
	err = b.validateAzureOptions(ui, state, reqManager)
	if err != nil {
		return nil, fmt.Errorf("Some Azure options failed: %s", err)
	}

	var steps []multistep.Step

	if b.config.OSType == targets.Linux {
		certFileName := "cert.pem"
		keyFileName := "key.pem"

		steps = []multistep.Step{
			&lin.StepCreateCert{
				CertFileName:   certFileName,
				KeyFileName:    keyFileName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&targets.StepCreateService{
				Location:       b.config.Location,
				TmpServiceName: b.config.tmpServiceName,
			},
			&targets.StepUploadCertificate{
				TmpServiceName: b.config.tmpServiceName,
			},
			&lin.StepCreateVm{
				StorageAccount:   b.config.StorageAccount,
				StorageContainer: b.config.StorageContainer,
				TmpVmName:        b.config.tmpVmName,
				TmpServiceName:   b.config.tmpServiceName,
				InstanceSize:     b.config.InstanceSize,
				Username:         b.config.userName,
			},

			&targets.StepPollStatus{
				TmpServiceName: b.config.tmpServiceName,
				TmpVmName:      b.config.tmpVmName,
				OSType:         b.config.OSType,
			},

			&common.StepConnectSSH{
				SSHAddress:     lin.SSHAddress,
				SSHConfig:      lin.SSHConfig(b.config.userName),
				SSHWaitTimeout: 20 * time.Minute,
			},
			&common.StepProvision{},

			&lin.StepGeneralizeOs{
				Command: "sudo /usr/sbin/waagent -force -deprovision+user && export HISTSIZE=0 && sync",
			},
			&targets.StepStopVm{
				TmpVmName:      b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&targets.StepCreateImage{
				TmpServiceName:    b.config.tmpServiceName,
				TmpVmName:         b.config.tmpVmName,
				UserImageName:     b.config.userImageName,
				UserImageLabel:    b.config.UserImageLabel,
				RecommendedVMSize: b.config.InstanceSize,
			},
		}
	} else if b.config.OSType == targets.Windows {
		//		b.config.tmpVmName = "shchTemp"
		//		b.config.tmpServiceName = "shchTemp"
		steps = []multistep.Step{

			&targets.StepCreateService{
				Location:       b.config.Location,
				TmpServiceName: b.config.tmpServiceName,
			},
			&win.StepCreateVm{
				StorageAccount:   b.config.StorageAccount,
				StorageContainer: b.config.StorageContainer,
				TmpVmName:        b.config.tmpVmName,
				TmpServiceName:   b.config.tmpServiceName,
				InstanceSize:     b.config.InstanceSize,
				Username:         b.config.userName,
				Password:         utils.RandomPassword(),
			},
			&targets.StepPollStatus{
				TmpServiceName: b.config.tmpServiceName,
				TmpVmName:      b.config.tmpVmName,
				OSType:         b.config.OSType,
			},
			&win.StepSetProvisionInfrastructure{
				VmName:             b.config.tmpVmName,
				ServiceName:        b.config.tmpServiceName,
				StorageAccountName: b.config.StorageAccount,
				TempContainerName:  b.config.tmpContainerName,
			},
			&common.StepProvision{},
			&targets.StepStopVm{
				TmpVmName:      b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&targets.StepCreateImage{
				TmpServiceName:    b.config.tmpServiceName,
				TmpVmName:         b.config.tmpVmName,
				UserImageName:     b.config.userImageName,
				UserImageLabel:    b.config.UserImageLabel,
				RecommendedVMSize: b.config.InstanceSize,
			},
		}

	} else {
		return nil, fmt.Errorf("Unkonwn OS type: %s", b.config.OSType)
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

	requestData := reqManager.GetVmImages()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		log.Printf("reqManager.GetVmImages returned error: %s", err.Error())
		return nil, fmt.Errorf("Can't create artifact")
	}

	vmImageList, err := response.ParseVmImageList(resp.Body)

	if err != nil {
		log.Printf("response.ParseVmImageList returned error: %s", err.Error())
		return nil, fmt.Errorf("Can't create artifact")
	}

	userImage := vmImageList.First(b.config.userImageName)
	if userImage == nil {
		log.Printf("vmImageList.First returned nil")
		return nil, fmt.Errorf("Can't create artifact")
	}

	return &artifact{
		imageLabel:    userImage.Label,
		imageName:     userImage.Name,
		mediaLocation: userImage.OSDiskConfiguration.MediaLink,
	}, nil
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func (b *Builder) validateAzureOptions(ui packer.Ui, state *multistep.BasicStateBag, reqManager *request.Manager) error {

	var err error

	// Check Storage account (& container)
	ui.Message("Checking Storage Account...")

	requestData := reqManager.CheckStorageAccountNameAvailability(b.config.StorageAccount)
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		return err
	}

	availabilityResponse, err := response.ParseAvailabilityResponse(resp.Body)

	log.Printf("availabilityResponse:\n %v", availabilityResponse)

	if availabilityResponse.Result == "true" {
		return fmt.Errorf("Can't Find Storage Account '%s'", b.config.StorageAccount)
	}

	// Check image exists
	exists, err := b.checkOsImageExists(ui, state, reqManager)
	if err != nil {
		return err
	}

	if exists == false {
		exists, err = b.checkOsUserImageExists(ui, state, reqManager)
		if err != nil {
			return err
		}

		if exists == false {
			err = fmt.Errorf("Can't Find OS Image '%s' Located at '%s'", b.config.OSImageLabel, b.config.Location)
			return err
		}
	}

	return nil
}

func (b *Builder) checkOsImageExists(ui packer.Ui, state *multistep.BasicStateBag, reqManager *request.Manager) (bool, error) {
	ui.Message("Checking OS image with the label '" + b.config.OSImageLabel + "' exists...")
	requestData := reqManager.GetOsImages()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		return false, err
	}

	imageList, err := response.ParseOsImageList(resp.Body)

	if err != nil {
		return false, err
	}

	filteredImageList := imageList.Filter(b.config.OSImageLabel, b.config.Location)

	if len(filteredImageList) != 0 {

		ui.Message(fmt.Sprintf("Found %v image(s).", len(filteredImageList)))
		ui.Message("Take the most recent:")

		imageList.SortByDateDesc(filteredImageList)

		osImageName := filteredImageList[0].Name
		ui.Message("OS Image Label: " + filteredImageList[0].Label)
		ui.Message("OS Image Family: " + filteredImageList[0].ImageFamily)
		ui.Message("OS Image Name: " + osImageName)
		ui.Message("OS Image PublishedDate: " + filteredImageList[0].PublishedDate)
		state.Put(constants.OsImageName, osImageName)
		state.Put(constants.IsOSImage, true)
		return true, nil
	}

	ui.Message("Image not found.")
	return false, nil
}

func (b *Builder) checkOsUserImageExists(ui packer.Ui, state *multistep.BasicStateBag, reqManager *request.Manager) (bool, error) {
	// check user images
	ui.Message("Checking VM image with the label '" + b.config.OSImageLabel + "' exists...")

	requestData := reqManager.GetVmImages()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		return false, err
	}

	imageList, err := response.ParseVmImageList(resp.Body)

	if err != nil {
		return false, err
	}

	filteredImageList := imageList.Filter(b.config.OSImageLabel, b.config.Location)

	if len(filteredImageList) != 0 {

		ui.Message(fmt.Sprintf("Found %v image(s).", len(filteredImageList)))
		ui.Message("Take the most recent:")

		imageList.SortByDateDesc(filteredImageList)

		osImageName := filteredImageList[0].Name
		ui.Message("VM Image Label: " + filteredImageList[0].Label)
		ui.Message("VM Image Family: " + filteredImageList[0].ImageFamily)
		ui.Message("VM Image Name: " + osImageName)
		ui.Message("VM Image PublishedDate: " + filteredImageList[0].PublishedDate)
		state.Put(constants.OsImageName, osImageName)
		state.Put(constants.IsOSImage, false)
		return true, nil
	}

	ui.Message("Image not found.")
	return false, nil
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
