// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	"errors"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/management"
	vmimage "github.com/Azure/azure-sdk-for-go/management/virtualmachineimage"

	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/Azure/packer-azure/packer/builder/azure/common/lin"
	"github.com/Azure/packer-azure/packer/builder/azure/smapi/win"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/packer"
)

// Builder implements packer.Builder and builds the actual Azure
// images.
type Builder struct {
	config *Config
	runner multistep.Runner
	client management.Client
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
// a Azure VM image.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	ui.Say("Preparing builder...")

	ui.Message("Creating Azure Service Management client...")
	subscriptionID, err := findSubscriptionID(b.config.PublishSettingsPath, b.config.SubscriptionName)
	if err != nil {
		return nil, fmt.Errorf("Error creating new Azure client: %v", err)
	}
	b.client, err = management.ClientFromPublishSettingsFile(b.config.PublishSettingsPath, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("Error creating new Azure client: %v", err)
	}

	// add logger if appropriate
	b.client = GetLoggedClient(b.client)

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put(constants.Config, b.config)
	state.Put(constants.RequestManager, b.client)
	state.Put("hook", hook)
	state.Put(constants.Ui, ui)

	// complete flags
	state.Put(constants.SrvExists, 0)
	state.Put(constants.CertInstalled, 0)
	state.Put(constants.CertUploaded, 0)
	state.Put(constants.VmExists, 0)
	state.Put(constants.DiskExists, 0)
	state.Put(constants.VmRunning, 0)
	state.Put(constants.ImageCreated, 0)

	var steps []multistep.Step

	if b.config.OSType == constants.Target_Linux {
		steps = []multistep.Step{
			&lin.StepCreateCert{
				TmpServiceName: b.config.tmpServiceName,
			},
			new(StepValidate),
			&StepCreateService{
				Location:       b.config.Location,
				TmpServiceName: b.config.tmpServiceName,
			},
			&StepUploadCertificate{
				TmpServiceName: b.config.tmpServiceName,
			},
			new(StepCreateVm),
			&StepPollStatus{
				TmpServiceName: b.config.tmpServiceName,
				TmpVmName:      b.config.tmpVmName,
				OSType:         b.config.OSType,
			},

			&communicator.StepConnectSSH{
				Config:    &b.config.Comm,
				Host:      lin.SSHHost,
				SSHConfig: lin.SSHConfig(b.config.UserName),
			},
			&common.StepProvision{},

			&lin.StepGeneralizeOS{
				Command: "sudo /usr/sbin/waagent -force -deprovision+user && export HISTSIZE=0 && sync",
			},
			&StepStopVm{
				TmpVmName:      b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&StepCreateImage{
				TmpServiceName:    b.config.tmpServiceName,
				TmpVmName:         b.config.tmpVmName,
				UserImageName:     b.config.userImageName,
				UserImageLabel:    b.config.UserImageLabel,
				RecommendedVMSize: b.config.InstanceSize,
			},
		}
	} else if b.config.OSType == constants.Target_Windows {
		steps = []multistep.Step{
			new(StepValidate),
			&StepCreateService{
				Location:       b.config.Location,
				TmpServiceName: b.config.tmpServiceName,
			},
			new(StepCreateVm),
			&StepPollStatus{
				TmpServiceName: b.config.tmpServiceName,
				TmpVmName:      b.config.tmpVmName,
				OSType:         b.config.OSType,
			},
			&win.StepSetProvisionInfrastructure{
				VmName:                    b.config.tmpVmName,
				ServiceName:               b.config.tmpServiceName,
				StorageAccountName:        b.config.StorageAccount,
				TempContainerName:         b.config.tmpContainerName,
				ProvisionTimeoutInMinutes: b.config.ProvisionTimeoutInMinutes,
			},
			&common.StepProvision{},
			&StepStopVm{
				TmpVmName:      b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&StepCreateImage{
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

	vmImageList, err := vmimage.NewClient(b.client).ListVirtualMachineImages(
		vmimage.ListParameters{
			Location: b.config.Location,
		})
	if err != nil {
		log.Printf("VM image client returned error: %s", err)
		return nil, fmt.Errorf("Can't create artifact")
	}

	if userImage, found := FindVmImage(vmImageList.VMImages, b.config.userImageName, b.config.UserImageLabel); found {
		return &artifact{
			imageLabel:    userImage.Label,
			imageName:     userImage.Name,
			mediaLocation: userImage.OSDiskConfiguration.MediaLink,

			publishSettingsPath: b.config.PublishSettingsPath,
			subscriptionID:      subscriptionID,
		}, nil
	} else {
		log.Printf("could not find image %s", b.config.userImageName)
		return nil, fmt.Errorf("Can't create artifact")
	}
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
