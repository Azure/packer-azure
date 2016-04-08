// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	"fmt"
	"strings"
	"time"

	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/Azure/packer-azure/packer/builder/azure/smapi/retry"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"github.com/Azure/azure-sdk-for-go/management"
	vmdisk "github.com/Azure/azure-sdk-for-go/management/virtualmachinedisk"
	vmi "github.com/Azure/azure-sdk-for-go/management/virtualmachineimage"
)

type StepCreateImage struct {
	TmpServiceName    string
	TmpVmName         string
	UserImageLabel    string
	UserImageName     string
	RecommendedVMSize string
}

func (s *StepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	ui := state.Get(constants.Ui).(packer.Ui)

	errorMsg := "Error Creating Azure Image: %s"

	ui.Say("Creating Azure Image. If Successful, This Will Remove the Temporary VM...")

	description := "packer made image"
	imageFamily := "PackerMade"

	if err := retry.ExecuteAsyncOperation(client, func() (management.OperationID, error) {
		return vmi.NewClient(client).Capture(s.TmpServiceName, s.TmpVmName, s.TmpVmName,
			s.UserImageName, s.UserImageLabel, vmi.OSStateGeneralized, vmi.CaptureParameters{
				Description:       description,
				ImageFamily:       imageFamily,
				RecommendedVMSize: s.RecommendedVMSize,
			})
	}); err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// CatpureVMImage removes the VM
	state.Put(constants.ImageCreated, 1)
	state.Put(constants.VmExists, 0)

	return multistep.ActionContinue
}

func (s *StepCreateImage) Cleanup(state multistep.StateBag) {
	client := state.Get(constants.RequestManager).(management.Client)
	ui := state.Get(constants.Ui).(packer.Ui)

	var err error
	var res int

	if res = state.Get(constants.VmExists).(int); res == 1 { //VM was not removed at image creation step
		return
	}

	// Since VM was successfully removed - remove it's media as well

	if res = state.Get(constants.DiskExists).(int); res == 1 {
		ui.Message("Removing Temporary Azure Disk...")
		errorMsg := "Error Removing Temporary Azure Disk: %s"

		diskName, ok := state.Get(constants.HardDiskName).(string)
		if ok {
			if len(diskName) == 0 {
				err := fmt.Errorf(errorMsg, err)
				ui.Error(err.Error())
				return
			}

			if err := retry.ExecuteOperation(func() error {
				return vmdisk.NewClient(client).DeleteDisk(diskName, true)
			}, retry.ConstantBackoffRule("busy", func(err management.AzureError) bool {
				return strings.Contains(err.Message, "is currently performing an operation on deployment") ||
					strings.Contains(err.Message, "is currently in use by virtual machine")
			}, 30*time.Second, 10)); err != nil {
				err := fmt.Errorf(errorMsg, err)
				ui.Error(err.Error())
				return
			}

			state.Put(constants.DiskExists, 0)
		}
	}
}
