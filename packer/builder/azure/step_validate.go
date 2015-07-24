package azure

import (
	"encoding/xml"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/location"
	"github.com/Azure/azure-sdk-for-go/management/osimage"
	"github.com/Azure/azure-sdk-for-go/management/storageservice"
	vmdisk "github.com/Azure/azure-sdk-for-go/management/virtualmachinedisk"
	vmimage "github.com/Azure/azure-sdk-for-go/management/virtualmachineimage"
	"github.com/Azure/azure-sdk-for-go/management/vmutils"
	"github.com/Azure/packer-azure/packer/builder/azure/constants"
	"github.com/Azure/packer-azure/packer/builder/azure/utils"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"net/http"
	"strings"
)

type StepValidate struct{}

func (*StepValidate) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get(constants.RequestManager).(management.Client)
	ui := state.Get(constants.Ui).(packer.Ui)
	config := state.Get(constants.Config).(*Config)

	ui.Say("Validating Azure options...")

	locationsResponse, err := location.NewClient(client).ListLocations()
	if err != nil {
		err = fmt.Errorf("Error checking location: %v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if err := func() error {
		for _, l := range locationsResponse.Locations {
			if config.Location == l.Name {
				ui.Message("Checking instance size availability...")
				if !func() bool {
					for _, size := range l.VirtualMachineRoleSizes {
						if size == config.InstanceSize {
							return true
						}
					}
					return false
				}() {
					sizes := strings.Join(l.VirtualMachineRoleSizes, ",")
					return fmt.Errorf("Instance size %q not available in location %q not available for this subscription, valid instance sizes: %s",
						config.InstanceSize, config.Location, sizes)
				}
				return nil
			}
		}
		return fmt.Errorf("Location %q not available for this subscription, valid locations: %s", config.Location, locationsResponse.String())
	}(); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	role := vmutils.NewVMConfiguration(config.tmpVmName, config.InstanceSize)

	ui.Message("Checking storage account...")
	destinationVhd, err := validateStorageAccount(config, client)
	if err != nil {
		err = fmt.Errorf("Error checking storage account: %v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Message(fmt.Sprintf("Destination VHD: %s", destinationVhd))

	if err := func() error {
		if config.RemoteSourceImageLink != "" {
			ui.Message("Checking remote image source link...")
			response, err := http.DefaultClient.Head(config.RemoteSourceImageLink)
			if response != nil && response.Body != nil {
				defer response.Body.Close()
			}
			if err != nil {
				log.Printf("HTTP client returned error: %s", err)
				return fmt.Errorf("error checking remote image source link: %v", err)
			}
			if response.StatusCode != 200 {
				return fmt.Errorf("Unexpected status while retrieving remote image source at %s: %d %s", config.RemoteSourceImageLink, response.StatusCode, response.Status)
			}
			size := float64(response.ContentLength) / 1024 / 1024 / 1024
			ui.Say(fmt.Sprintf("Remote image size: %.1f GiB", size))

			vmutils.ConfigureDeploymentFromRemoteImage(&role, config.RemoteSourceImageLink, config.OSType, fmt.Sprintf("%s-OSDisk", config.tmpVmName), destinationVhd, "")
			if config.ResizeOSVhdGB != nil {
				if float64(*config.ResizeOSVhdGB) < size {
					return fmt.Errorf("new OS VHD size of %d GiB is smaller than current size of %.1f GiB", *config.ResizeOSVhdGB, size)
				}
				ui.Say(fmt.Sprintf("Remote image will be resized to %d GiB", *config.ResizeOSVhdGB))
				role.OSVirtualHardDisk.ResizedSizeInGB = *config.ResizeOSVhdGB
			}

		} else {
			ui.Message("Checking image source...")
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
				if config.ResizeOSVhdGB != nil {
					if float64(*config.ResizeOSVhdGB) < osImage.LogicalSizeInGB {
						return fmt.Errorf("new OS VHD size of %d GiB is smaller than current size of %.1f GiB", *config.ResizeOSVhdGB, osImage.LogicalSizeInGB)
					}
					ui.Say(fmt.Sprintf("OS image will be resized to %d GiB", *config.ResizeOSVhdGB))
					role.OSVirtualHardDisk.ResizedSizeInGB = *config.ResizeOSVhdGB
				}
			} else {
				imageList, err := vmimage.NewClient(client).ListVirtualMachineImages()
				if err != nil {
					log.Printf("VM image client returned error: %s", err)
					return err
				}

				if vmImage, found := FindVmImage(imageList.VMImages, "", config.OSImageLabel, config.Location); found {
					if config.ResizeOSVhdGB != nil {
						return fmt.Errorf("VM images cannot be resized")
					}

					vmutils.ConfigureDeploymentFromVMImage(&role, vmImage.Name, destinationVhd, true)
					ui.Message(fmt.Sprintf("Image source is VM image %q", vmImage.Name))
					if vmImage.OSDiskConfiguration.OS != config.OSType {
						return fmt.Errorf("VM image type (%q) does not match config (%q)", vmImage.OSDiskConfiguration.OS, config.OSType)
					}
				} else {
					return fmt.Errorf("Can't find VM or OS image '%s' Located at '%s'", config.OSImageLabel, config.Location)
				}
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
		vmutils.ConfigureForLinux(&role, config.tmpVmName, config.UserName, "", certThumbprint)
		vmutils.ConfigureWithPublicSSH(&role)
	} else if config.OSType == constants.Target_Windows {
		password := utils.RandomPassword()
		state.Put("password", password)
		vmutils.ConfigureForWindows(&role, config.tmpVmName, config.UserName, password, true, "")
		vmutils.ConfigureWithPublicRDP(&role)
		vmutils.ConfigureWithPublicPowerShell(&role)
	}

	if config.VNet != "" && config.Subnet != "" {
		ui.Message("Checking VNet...")
		if err := checkVirtualNetworkConfiguration(client, config.VNet, config.Subnet, config.Location); err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		vmutils.ConfigureWithSubnet(&role, config.Subnet)
	}

	for n, d := range config.DataDisks {
		switch d := d.(type) {
		case int:
			ui.Message(fmt.Sprintf("Configuring datadisk %d: new disk with size %d GB...", n, d))
			destination := fmt.Sprintf("%s-data-%d.vhd", destinationVhd[:len(destinationVhd)-4], n)
			ui.Message(fmt.Sprintf("Destination VHD for data disk %d: %s", destinationVhd, n))
			vmutils.ConfigureWithNewDataDisk(&role, "", destination, d, vmdisk.HostCachingTypeNone)
		case string:
			ui.Message(fmt.Sprintf("Configuring datadisk %d: existing blob (%s)...", n, d))
			vmutils.ConfigureWithVhdDataDisk(&role, d, vmdisk.HostCachingTypeNone)
		default:
			err := fmt.Errorf("Datadisk %d is not a string nor a number", n)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

	}

	state.Put("role", &role)

	return multistep.ActionContinue
}

func (*StepValidate) Cleanup(multistep.StateBag) {}

func validateStorageAccount(config *Config, client management.Client) (string, error) {
	ssc := storageservice.NewClient(client)

	sa, err := ssc.GetStorageService(config.StorageAccount)
	if err != nil {
		return "", err
	}

	if sa.StorageServiceProperties.Location != config.Location {
		return "", fmt.Errorf("Storage account %q is not in location %q, but in location %q.",
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

func checkVirtualNetworkConfiguration(client management.Client, vnetname, subnetname, location string) error {
	const getVNetConfig = "services/networking/media"
	d, err := client.SendAzureGetRequest(getVNetConfig)
	if err != nil {
		return err
	}

	var vnetConfig struct {
		VNets []struct {
			Name          string `xml:"name,attr"`
			AffinityGroup string `xml:",attr"`
			Location      string `xml:",attr"`
			Subnets       []struct {
				Name          string `xml:"name,attr"`
				AddressPrefix string
			} `xml:"Subnets>Subnet"`
		} `xml:"VirtualNetworkConfiguration>VirtualNetworkSites>VirtualNetworkSite"`
	}
	err = xml.Unmarshal(d, &vnetConfig)
	if err != nil {
		return err
	}

	for _, vnet := range vnetConfig.VNets {
		if vnet.Name == vnetname {
			if vnet.AffinityGroup != "" {
				vnet.Location, err = getAffinityGroupLocation(client, vnet.AffinityGroup)
				if err != nil {
					return err
				}
			}
			if vnet.Location != location {
				return fmt.Errorf("VNet %q is not in location %q, but in %q", vnet.Name, location, vnet.Location)
			}

			for _, sn := range vnet.Subnets {
				if sn.Name == subnetname {
					return nil
				}
			}
		}
	}

	return fmt.Errorf("Could not find vnet %q and subnet %q in network configuration: %v", vnetname, subnetname, vnetConfig)
}

func getAffinityGroupLocation(client management.Client, affinityGroup string) (string, error) {
	const getAffinityGroupProperties = "affinitygroups/%s"
	d, err := client.SendAzureGetRequest(fmt.Sprintf(getAffinityGroupProperties, affinityGroup))
	if err != nil {
		return "", err
	}

	var afGroup struct {
		Location string
	}
	err = xml.Unmarshal(d, &afGroup)

	return afGroup.Location, err
}
