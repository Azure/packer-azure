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

func TestStepValidateTemplateShouldFailIfValidateFails(t *testing.T) {

	var testSubject = &StepValidateTemplate{
		validate: func(string, string, *TemplateParameters) error { return fmt.Errorf("!! Unit Test FAIL !!") },
		say:      func(message string) {},
		error:    func(e error) {},
	}

	stateBag := createTestStateBagStepValidateTemplate()

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

func TestStepValidateTemplateShouldPassIfValidatePasses(t *testing.T) {
	var testSubject = &StepValidateTemplate{
		validate: func(string, string, *TemplateParameters) error { return nil },
		say:      func(message string) {},
		error:    func(e error) {},
	}

	stateBag := createTestStateBagStepValidateTemplate()

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

func TestStepValidateTemplateShouldTakeValidateArgumentsFromStateBag(t *testing.T) {
	var actualResourceGroupName string
	var actualDeploymentName string
	var actualTemplateParameters *TemplateParameters

	var testSubject = &StepValidateTemplate{
		validate: func(resourceGroupName string, deploymentName string, templateParameter *TemplateParameters) error {
			actualResourceGroupName = resourceGroupName
			actualDeploymentName = deploymentName
			actualTemplateParameters = templateParameter

			return nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepValidateTemplate()
	var result = testSubject.Run(stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%s'.", result)
	}

	var expectedDeploymentName = stateBag.Get(constants.ArmDeploymentName).(string)
	var expectedResourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)
	var expectedTemplateParameters = stateBag.Get(constants.ArmTemplateParameters).(*TemplateParameters)

	if actualDeploymentName != expectedDeploymentName {
		t.Fatalf("Expected the step to source 'constants.ArmDeploymentName' from the state bag, but it did not.")
	}

	if actualResourceGroupName != expectedResourceGroupName {
		t.Fatalf("Expected the step to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}

	if actualTemplateParameters != expectedTemplateParameters {
		t.Fatalf("Expected the step to source 'constants.ArmTemplateParameters' from the state bag, but it did not.")
	}
}

func createTestStateBagStepValidateTemplate() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmDeploymentName, "Unit Test: DeploymentName")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")
	stateBag.Put(constants.ArmTemplateParameters, &TemplateParameters{})

	return stateBag
}
