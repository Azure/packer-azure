// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package arm

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
)

func TestStepCaptureImageShouldFailIfValidateFails(t *testing.T) {

	var testSubject = &StepCaptureImage{
		capture: func(string, string, *compute.VirtualMachineCaptureParameters) error {
			return fmt.Errorf("!! Unit Test FAIL !!")
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepCaptureImage()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%s'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}

	if _, ok := stateBag.GetOk(constants.ArmIsValid); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.ArmIsValid)
	}
}

func TestStepCaptureImageShouldPassIfValidatePasses(t *testing.T) {
	var testSubject = &StepCaptureImage{
		capture: func(string, string, *compute.VirtualMachineCaptureParameters) error { return nil },
		say:     func(message string) {},
		error:   func(e error) {},
	}

	stateBag := createTestStateBagStepCaptureImage()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%s'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}

	if _, ok := stateBag.GetOk(constants.ArmIsValid); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.ArmIsValid)
	}
}

func TestStepCaptureImageShouldTakeValidateArgumentsFromStateBag(t *testing.T) {
	var actualResourceGroupName string
	var actualComputeName string
	var actualVirtualMachineCaptureParameters *compute.VirtualMachineCaptureParameters

	var testSubject = &StepCaptureImage{
		capture: func(resourceGroupName string, computeName string, parameters *compute.VirtualMachineCaptureParameters) error {
			actualResourceGroupName = resourceGroupName
			actualComputeName = computeName
			actualVirtualMachineCaptureParameters = parameters

			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepCaptureImage()
	var result = testSubject.Run(stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%s'.", result)
	}

	var expectedComputeName = stateBag.Get(constants.ArmComputeName).(string)
	var expectedResourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)
	var expectedVirtualMachineCaptureParameters = stateBag.Get(constants.ArmVirtualMachineCaptureParameters).(*compute.VirtualMachineCaptureParameters)

	if actualComputeName != expectedComputeName {
		t.Fatalf("Expected StepCaptureImage to source 'constants.ArmComputeName' from the state bag, but it did not.")
	}

	if actualResourceGroupName != expectedResourceGroupName {
		t.Fatalf("Expected StepCaptureImage to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}

	if actualVirtualMachineCaptureParameters != expectedVirtualMachineCaptureParameters {
		t.Fatalf("Expected StepCaptureImage to source 'constants.ArmVirtualMachineCaptureParameters' from the state bag, but it did not.")
	}
}

func createTestStateBagStepCaptureImage() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmComputeName, "Unit Test: ComputeName")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")
	stateBag.Put(constants.ArmVirtualMachineCaptureParameters, &compute.VirtualMachineCaptureParameters{})

	return stateBag
}
