// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/to"
	"github.com/Azure/azure-sdk-for-go/arm/compute"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const (
	DefaultUserName = "packer"
	DefaultVMSize   = "Standard_A1"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Authentication via OAUTH
	ClientID       string `mapstructure:"client_id"`
	ClientSecret   string `mapstructure:"client_secret"`
	TenantID       string `mapstructure:"tenant_id"`
	SubscriptionID string `mapstructure:"subscription_id"`

	// Capture
	CaptureNamePrefix    string `mapstructure:"capture_name_prefix"`
	CaptureContainerName string `mapstructure:"capture_container_name"`

	// Compute
	ImagePublisher string `mapstructure:"image_publisher"`
	ImageOffer     string `mapstructure:"image_offer"`
	ImageSku       string `mapstructure:"image_sku"`
	Location       string `mapstructure:"location"`
	VMSize         string `mapstructure:"vm_size"`

	// Deployment
	ResourceGroupName string `mapstructure:"resource_group_name"`
	StorageAccount    string `mapstructure:"storage_account"`

	// Packer
	UserName string `mapaccount:"username"`

	// Runtime Values
	tmpAdminPassword     string
	tmpResourceGroupName string
	tmpComputeName       string
	tmpDeploymentName    string
	tmpOSDiskName        string

	// Authentication with the VM via SSH
	sshAuthorizedKey string
	sshPrivateKey    string

	Comm communicator.Config `mapstructure:",squash"`
	ctx  *interpolate.Context
}

// If we ever feel the need to support more templates consider moving this
// method to its own factory class.
func (c *Config) toTemplateParameters() *TemplateParameters {
	return &TemplateParameters{
		AdminUsername:      &TemplateParameter{c.UserName},
		AdminPassword:      &TemplateParameter{c.tmpAdminPassword},
		DnsNameForPublicIP: &TemplateParameter{c.tmpComputeName},
		ImageOffer:         &TemplateParameter{c.ImageOffer},
		ImagePublisher:     &TemplateParameter{c.ImagePublisher},
		ImageSku:           &TemplateParameter{c.ImageSku},
		Location:           &TemplateParameter{c.Location},
		OSDiskName:         &TemplateParameter{c.tmpOSDiskName},
		SshAuthorizedKey:   &TemplateParameter{c.sshAuthorizedKey},
		StorageAccountName: &TemplateParameter{c.StorageAccount},
		VMSize:             &TemplateParameter{c.VMSize},
		VMName:             &TemplateParameter{c.tmpComputeName},
	}
}

func (c *Config) toVirtualMachineCaptureParameters() *compute.VirtualMachineCaptureParameters {
	return &compute.VirtualMachineCaptureParameters{
		DestinationContainerName: &c.CaptureContainerName,
		VhdPrefix:                &c.CaptureNamePrefix,
		OverwriteVhds:            to.BoolPtr(false),
	}
}

func newConfig(raws ...interface{}) (*Config, []string, error) {
	var c Config

	err := config.Decode(&c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: c.ctx,
	}, raws...)

	if err != nil {
		return nil, nil, err
	}

	provideDefaultValues(&c)
	setRuntimeValues(&c)

	err = setSshValues(&c)
	if err != nil {
		return nil, nil, err
	}

	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(c.ctx)...)

	assertRequiredParametersSet(&c, errs)
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return &c, nil, nil
}

func setSshValues(c *Config) error {
	c.Comm.SSHUsername = c.UserName

	if c.Comm.SSHTimeout == 0 {
		c.Comm.SSHTimeout = 20 * time.Minute
	}

	sshKeyPair, err := NewOpenSshKeyPair()
	if err != nil {
		return err
	}

	c.sshAuthorizedKey = sshKeyPair.AuthorizedKey()
	c.sshPrivateKey = sshKeyPair.PrivateKey()

	return nil
}

func setRuntimeValues(c *Config) {
	var tempName = NewTempName()

	c.tmpAdminPassword = tempName.AdminPassword
	c.tmpComputeName = tempName.ComputeName
	c.tmpDeploymentName = tempName.DeploymentName
	c.tmpResourceGroupName = tempName.ResourceGroupName
	c.tmpOSDiskName = tempName.OSDiskName
}

func provideDefaultValues(c *Config) {
	if c.UserName == "" {
		c.UserName = DefaultUserName
	}

	if c.VMSize == "" {
		c.VMSize = DefaultVMSize
	}
}

func assertRequiredParametersSet(c *Config, errs *packer.MultiError) {
	/////////////////////////////////////////////
	// Authentication via OAUTH

	if c.ClientID == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A client_id must be specified"))
	}

	if c.ClientSecret == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A client_secret must be specified"))
	}

	if c.TenantID == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A tenant_id must be specified"))
	}

	if c.SubscriptionID == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A subscription_id must be specified"))
	}

	/////////////////////////////////////////////
	// Capture
	if c.CaptureContainerName == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("An capture_container_name must be specified"))
	}

	if c.CaptureNamePrefix == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("An capture_name_prefix must be specified"))
	}

	/////////////////////////////////////////////
	// Compute

	if c.ImagePublisher == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A image_publisher must be specified"))
	}

	if c.ImageOffer == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A image_offer must be specified"))
	}

	if c.ImageSku == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A image_sku must be specified"))
	}

	if c.Location == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A location must be specified"))
	}

	/////////////////////////////////////////////
	// Deployment

	if c.StorageAccount == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A storage_account must be specified"))
	}
}
