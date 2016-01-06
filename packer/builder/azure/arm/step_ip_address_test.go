// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package arm

import (
	"fmt"
	"testing"

	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
)

func TestStepIPAddressShouldFailIfValidateFails(t *testing.T) {
	var testSubject = &StepIPAddress{
		get:   func(string, string) (string, error) { return "", fmt.Errorf("!! Unit Test FAIL !!") },
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepIPAddress()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%s'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepIPAddressShouldPassIfValidatePasses(t *testing.T) {
	var testSubject = &StepIPAddress{
		get:   func(string, string) (string, error) { return "", nil },
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepIPAddress()

	var result = testSubject.Run(stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%s'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepIPAddressShouldTakeValidateArgumentsFromStateBag(t *testing.T) {
	var actualResourceGroupName string
	var actualIPAddressName string

	var testSubject = &StepIPAddress{
		get: func(resourceGroupName string, ipAddressName string) (string, error) {
			actualResourceGroupName = resourceGroupName
			actualIPAddressName = ipAddressName

			return "127.0.0.1", nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepIPAddress()
	var result = testSubject.Run(stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%s'.", result)
	}

	var expectedResourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)
	var expectedIPAddressName = stateBag.Get(constants.ArmPublicIPAddressName).(string)

	if actualIPAddressName != expectedIPAddressName {
		t.Fatalf("Expected StepValidateTemplate to source 'constants.ArmIPAddressName' from the state bag, but it did not.")
	}

	if actualResourceGroupName != expectedResourceGroupName {
		t.Fatalf("Expected StepValidateTemplate to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}

	expectedIPAddress, ok := stateBag.GetOk(constants.SSHHost)
	if !ok {
		t.Fatalf("Expected the state bag to have a value for '%s', but it did not.", constants.SSHHost)
	}

	if expectedIPAddress != "127.0.0.1" {
		t.Fatalf("Expected the value of stateBag[%s] to be '127.0.0.1', but got '%s'.", constants.SSHHost, expectedIPAddress)
	}
}

func createTestStateBagStepIPAddress() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmPublicIPAddressName, "Unit Test: PublicIPAddressName")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")

	return stateBag
}
