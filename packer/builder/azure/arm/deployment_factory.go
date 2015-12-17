// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package arm

import (
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/arm/resources"
)

type DeploymentFactory struct {
	template string
}

func NewDeploymentFactory(template string) DeploymentFactory {
	return DeploymentFactory{
		template: template,
	}
}

func (f *DeploymentFactory) create(templateParameters TemplateParameters) (*resources.Deployment, error) {
	template, err := f.getTemplate(templateParameters)
	if err != nil {
		return nil, err
	}

	parameters, err := f.getTemplateParameters(templateParameters)
	if err != nil {
		return nil, err
	}

	return &resources.Deployment{
		Properties: &resources.DeploymentProperties{
			Mode:       "Incremental",
			Template:   template,
			Parameters: parameters,
		},
	}, nil
}

func (f *DeploymentFactory) getTemplate(templateParameters TemplateParameters) (*map[string]interface{}, error) {
	var t map[string]interface{}
	err := json.Unmarshal([]byte(f.template), &t)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (f *DeploymentFactory) getTemplateParameters(templateParameters TemplateParameters) (*map[string]interface{}, error) {
	b, err := json.Marshal(templateParameters)
	if err != nil {
		return nil, err
	}

	var t map[string]interface{}
	err = json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
