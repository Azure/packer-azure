package driver_restapi

import (
	"fmt"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/targets"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"regexp"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	SubscriptionName    string `mapstructure:"subscription_name"`
	PublishSettingsPath string `mapstructure:"publish_settings_path"`
	StorageAccount      string `mapstructure:"storage_account"`
	StorageContainer    string `mapstructure:"storage_account_container"`
	OSType              string `mapstructure:"os_type"`
	OSImageLabel        string `mapstructure:"os_image_label"`
	Location            string `mapstructure:"location"`
	InstanceSize        string `mapstructure:"instance_size"`
	UserImageLabel      string `mapstructure:"user_image_label"`

	userName         string `mapstructure:"username"`
	tmpVmName        string
	tmpServiceName   string
	tmpContainerName string
	userImageName    string
}

func newConfig(raws ...interface{}) (*Config, []string, error) {
	var c Config

	var md mapstructure.Metadata
	err := config.Decode(&c, &config.DecodeOpts{
		Metadata:    &md,
		Interpolate: true,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Defaults
	log.Println(fmt.Sprintf("%s: %v", "PackerUserVars", c.PackerConfig.PackerUserVars))

	if c.StorageContainer == "" {
		c.StorageContainer = "vhds"
	}

	if c.userName == "" {
		c.userName = "packer"
	}

	randSuffix := utils.RandomString("0123456789abcdefghijklmnopqrstuvwxyz", 10)
	c.tmpVmName = "PkrVM" + randSuffix
	c.tmpServiceName = "PkrSrv" + randSuffix
	c.tmpContainerName = "packer-provision-" + randSuffix

	// Check values
	var errs *packer.MultiError

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

	if !(c.OSType == targets.Linux || c.OSType == targets.Windows) {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("os_type is not valid, must be one of: %s, %s", targets.Windows, targets.Linux))
	}

	if c.OSImageLabel == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("os_image_label must be specified"))
	}

	if c.Location == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("location must be specified"))
	}

	sizeIsValid := false

	for _, instanceSize := range targets.VMSizes {
		if c.InstanceSize == instanceSize {
			sizeIsValid = true
			break
		}
	}

	if !sizeIsValid {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("instance_size is not valid, must be one of: %v", targets.VMSizes))
	}

	if c.UserImageLabel == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("user_image_label must be specified"))
	}

	const userLabelRegex = "^[A-Za-z][A-Za-z0-9-_.]*[A-Za-z0-9]$"
	if !regexp.MustCompile(userLabelRegex).MatchString(c.UserImageLabel) {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("user_image_label is not valid, it should follow the pattern %s", userLabelRegex))
	}

	c.userImageName = utils.DecorateImageName(c.UserImageLabel)

	log.Println(fmt.Sprintf("%s: %v", "subscription_name", c.SubscriptionName))
	log.Println(fmt.Sprintf("%s: %v", "publish_settings_path", c.PublishSettingsPath))
	log.Println(fmt.Sprintf("%s: %v", "storage_account", c.StorageAccount))
	log.Println(fmt.Sprintf("%s: %v", "storage_account_container", c.StorageContainer))
	log.Println(fmt.Sprintf("%s: %v", "os_type", c.OSType))
	log.Println(fmt.Sprintf("%s: %v", "os_image_label", c.OSImageLabel))
	log.Println(fmt.Sprintf("%s: %v", "location", c.Location))
	log.Println(fmt.Sprintf("%s: %v", "instance_size", c.InstanceSize))
	log.Println(fmt.Sprintf("%s: %v", "user_image_name", c.userImageName))
	log.Println(fmt.Sprintf("%s: %v", "user_image_label", c.UserImageLabel))
	log.Println(fmt.Sprintf("%s: %v", "tmpContainerName", c.tmpContainerName))
	log.Println(fmt.Sprintf("%s: %v", "tmpVmName", c.tmpVmName))
	log.Println(fmt.Sprintf("%s: %v", "tmpServiceName", c.tmpServiceName))
	log.Println(fmt.Sprintf("%s: %v", "username", c.userName))

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return &c, nil, nil
}
