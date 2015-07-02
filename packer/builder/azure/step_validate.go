package azure

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/osimage"
	"github.com/Azure/azure-sdk-for-go/management/storageservice"
	vmimage "github.com/Azure/azure-sdk-for-go/management/virtualmachineimage"
	"github.com/Azure/azure-sdk-for-go/management/vmutils"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/constants"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/utils"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"strings"
)

type StepValidate struct{}

func (*StepValidate) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	ui := state.Get(constants.Ui).(packer.Ui)
	config := state.Get(constants.Config).(Config)

	ui.Say("Validating Azure options...")

	role := vmutils.NewVMConfiguration(config.tmpVmName, config.InstanceSize)

	ui.Message("Checking Storage Account...")
	destinationVhd, err := validateStorageAccount(config, client)
	if err != nil {
		err = fmt.Errorf("Error checking storage account: %v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Message(fmt.Sprintf("Destination VHD: %s", destinationVhd))

	// Check image exists
	if err := func() error {
		imageList, err := osimage.NewClient(client).ListOSImages()
		if err != nil {
			log.Printf("OS image client returned error: %s", err)
			return err
		}

		if osImage, found := FindOSImage(imageList.OSImages, config.OSImageLabel, config.Location); found {
			vmutils.ConfigureDeploymentFromPlatformImage(&role, osImage.Name, destinationVhd, "")
			ui.Message(fmt.Sprintf("Image source is OS image %q", osImage.Name))
			if osImage.OS != config.OSType {
				return fmt.Errorf("OS image type (%q) does not match config (%q)", osImage.OS, config.OSType)
			}
		} else {
			imageList, err := vmimage.NewClient(client).ListVirtualMachineImages()
			if err != nil {
				log.Printf("VM image client returned error: %s", err)
				return err
			}

			if vmImage, found := FindVmImage(imageList.VMImages, "", config.OSImageLabel, config.Location); found {
				vmutils.ConfigureDeploymentFromVMImage(&role, vmImage.Name, destinationVhd, true)
				ui.Message(fmt.Sprintf("Image source is VM image %q", vmImage.Name))
				if vmImage.OSDiskConfiguration.OS != config.OSType {
					return fmt.Errorf("VM image type (%q) does not match config (%q)", vmImage.OSDiskConfiguration.OS, config.OSType)
				}
			} else {
				return fmt.Errorf("Can't find VM or OS image '%s' Located at '%s'", config.OSImageLabel, config.Location)
			}
		}
		return nil
	}(); err != nil {
		err = fmt.Errorf("Error determining deployment source: %v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if config.OSType == constants.Target_Linux {
		certThumbprint := state.Get(constants.Thumbprint).(string)
		if len(certThumbprint) == 0 {
			err := fmt.Errorf("Certificate Thumbprint is empty")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		vmutils.ConfigureForLinux(&role, config.tmpVmName, config.userName, "", certThumbprint)
		vmutils.ConfigureWithPublicSSH(&role)
	} else if config.OSType == constants.Target_Windows {
		password := utils.RandomPassword()
		state.Put("password", password)
		vmutils.ConfigureForWindows(&role, config.tmpVmName, config.userName, password, true, "")
		vmutils.ConfigureWithPublicRDP(&role)
		vmutils.ConfigureWithPublicPowerShell(&role)
	}

	state.Put("role", &role)

	return multistep.ActionContinue
}

func validateStorageAccount(config Config, client management.Client) (string, error) {
	ssc := storageservice.NewClient(client)

	sa, err := ssc.GetStorageService(config.StorageAccount)
	if err != nil {
		return "", err
	}

	if sa.StorageServiceProperties.Location != config.Location {
		return "", fmt.Errorf("Storage Account %q is not in location %q, but in location %q.",
			config.StorageAccount, sa.StorageServiceProperties.Location, config.Location)
	}

	var blobEndpoint string
	for _, uri := range sa.StorageServiceProperties.Endpoints {
		if strings.Contains(uri, ".blob.") {
			blobEndpoint = uri
		}
	}
	if blobEndpoint == "" {
		return "", fmt.Errorf("Could not find blob endpoint for account %q in %v",
			sa.ServiceName, sa.StorageServiceProperties.Endpoints)
	}
	log.Printf("Blob endpoint: %s", blobEndpoint)

	return fmt.Sprintf("%s%s/%s.vhd", blobEndpoint, config.StorageContainer, config.tmpVmName), nil
}

func (*StepValidate) Cleanup(multistep.StateBag) {}
