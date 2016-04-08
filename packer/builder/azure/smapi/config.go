package azure

import (
	"fmt"
	azureCommon "github.com/Azure/packer-azure/packer/builder/azure/common"
	"github.com/Azure/packer-azure/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"log"
	"math"
	"os"
	"regexp"
	"time"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	SubscriptionName    string `mapstructure:"subscription_name"`
	PublishSettingsPath string `mapstructure:"publish_settings_path"`

	StorageAccount   string        `mapstructure:"storage_account"`
	StorageContainer string        `mapstructure:"storage_account_container"`
	Location         string        `mapstructure:"location"`
	InstanceSize     string        `mapstructure:"instance_size"`
	DataDisks        []interface{} `mapstructure:"data_disks"`
	UserImageLabel   string        `mapstructure:"user_image_label"`

	OSType                string `mapstructure:"os_type"`
	OSImageLabel          string `mapstructure:"os_image_label"`
	OSImageName           string `mapstructure:"os_image_name"`
	RemoteSourceImageLink string `mapstructure:"remote_source_image_link"`
	ResizeOSVhdGB         *int   `mapstructure:"resize_os_vhd_gb"`

	ProvisionTimeoutInMinutes uint `mapstructure:"provision_timeout_in_minutes"`

	VNet   string `mapstructure:"vnet"`
	Subnet string `mapstructure:"subnet"`

	UserName         string `mapstructure:"username"`
	tmpVmName        string
	tmpServiceName   string
	tmpContainerName string
	userImageName    string

	Comm communicator.Config `mapstructure:",squash"`

	ctx *interpolate.Context
}

func newConfig(raws ...interface{}) (*Config, []string, error) {
	var c Config

	// Default provision timeout
	c.ProvisionTimeoutInMinutes = 120

	c.ctx = &interpolate.Context{}
	err := config.Decode(&c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: c.ctx,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Defaults
	log.Println(fmt.Sprintf("%s: %v", "PackerUserVars", c.PackerConfig.PackerUserVars))

	if c.StorageContainer == "" {
		c.StorageContainer = "vhds"
	}

	if c.UserName == "" {
		c.UserName = "packer"
	}

	c.Comm.SSHUsername = c.UserName
	if c.Comm.SSHTimeout == 0 {
		c.Comm.SSHTimeout = 20 * time.Minute
	}

	randSuffix := azureCommon.RandomString("0123456789abcdefghijklmnopqrstuvwxyz", 10)
	c.tmpVmName = "PkrVM" + randSuffix
	c.tmpServiceName = "PkrSrv" + randSuffix
	c.tmpContainerName = "packer-provision-" + randSuffix

	// Check values
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(c.ctx)...)

	if c.SubscriptionName == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("subscription_name must be specified"))
	}

	if c.PublishSettingsPath == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("publish_settings_path must be specified"))
	}

	if c.StorageAccount == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("storage_account must be specified"))
	}

	if _, err := os.Stat(c.PublishSettingsPath); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("publish_settings_path is not a valid path: %s", err))
	}

	if !(c.OSType == constants.Target_Linux || c.OSType == constants.Target_Windows) {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("os_type is not valid, must be one of: %s, %s", constants.Target_Windows, constants.Target_Linux))
	}

	count := 0
	if c.RemoteSourceImageLink != "" {
		count += 1
	}
	if c.OSImageLabel != "" {
		count += 1
	}
	if c.OSImageName != "" {
		count += 1
	}

	if count != 1 {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("One source and only one among os_image_label, os_image_label or remote_source_image_link has to be specified"))
	}

	if c.Location == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("location must be specified"))
	}

	sizeIsValid := false

	for _, instanceSize := range allowedVMSizes {
		if c.InstanceSize == instanceSize {
			sizeIsValid = true
			break
		}
	}

	if !sizeIsValid {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("instance_size is not valid, must be one of: %v", allowedVMSizes))
	}

	for n := 0; n < len(c.DataDisks); n++ {
		switch v := c.DataDisks[n].(type) {
		case string:
		case int:
		case float64:
			if v != math.Floor(v) {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("Data disk # %d is a fractional number, needs to be integer", n))
			}
			c.DataDisks[n] = int(v)
		default:
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Data disk # %d is not a string to an existing VHD nor an integer number, but a %T", n, v))
		}
	}

	if c.UserImageLabel == "" {
		log.Println(fmt.Sprintf("Using dynamically generated user_image_label [%s]", c.tmpVmName))
		c.UserImageLabel = c.tmpVmName
	}

	const userLabelRegex = "^[A-Za-z][A-Za-z0-9-_.]*[A-Za-z0-9]$"
	if !regexp.MustCompile(userLabelRegex).MatchString(c.UserImageLabel) {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("user_image_label [%s] is not valid, it should follow the pattern %s", c.UserImageLabel, userLabelRegex))
	}

	c.userImageName = fmt.Sprintf("%s_%s", c.UserImageLabel, time.Now().Format("2006-01-02_15-04"))

	if (c.VNet != "" && c.Subnet == "") || (c.Subnet != "" && c.VNet == "") {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("vnet and subnet need to either both be set or both be empty"))
	}

	log.Println(common.ScrubConfig(c))

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return &c, nil, nil
}

// Target sizes
var allowedVMSizes = []string{
	"ExtraSmall",
	"Small",
	"Medium",
	"Large",
	"ExtraLarge",
	"A5",
	"A6",
	"A7",
	"A8",
	"A9",

	"Standard_D1",
	"Standard_D2",
	"Standard_D3",
	"Standard_D4",
	"Standard_D11",
	"Standard_D12",
	"Standard_D13",
	"Standard_D14",

	"Standard_DS1",
	"Standard_DS2",
	"Standard_DS3",
	"Standard_DS4",
	"Standard_DS11",
	"Standard_DS12",
	"Standard_DS13",
	"Standard_DS14",

	"Standard_G1",
	"Standard_G2",
	"Standard_G3",
	"Standard_G4",
	"Standard_G5",
}
