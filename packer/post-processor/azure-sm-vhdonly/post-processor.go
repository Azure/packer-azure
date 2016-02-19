// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azuresmvhdonly

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"

	azure "github.com/Azure/packer-azure/packer/builder/azure/smapi"
	"github.com/Azure/packer-azure/packer/builder/azure/smapi/retry"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/virtualmachineimage"

	"github.com/mitchellh/packer/packer"
)

var _ packer.PostProcessor = &PostProcessor{}

const stateError = "Error: Could not retrieve %s for this artifact. Make sure you used the same version of the builder as the post-processor."

type PostProcessor struct{}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact /*keep*/, bool, error) {
	ui.Say("Validating artifact")
	if artifact.BuilderId() != azure.BuilderId {
		return nil, false, fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from Azure builder artifacts (%s).",
			artifact.BuilderId(), azure.BuilderId)
	}

	publishSettingsPath, ok := artifact.State("publishSettingsPath").(string)
	if !ok || publishSettingsPath == "" {
		return nil, false, fmt.Errorf(stateError, "publishSettingsPath")
	}
	subscriptionID, ok := artifact.State("subscriptionID").(string)
	if !ok || subscriptionID == "" {
		return nil, false, fmt.Errorf(stateError, "subscriptionID")
	}

	name := artifact.Id()

	ui.Message("Creating Azure Service Management client...")
	client, err := management.ClientFromPublishSettingsFile(publishSettingsPath, subscriptionID)
	if err != nil {
		return nil, false, fmt.Errorf("Error creating new Azure client: %v", err)
	}
	client = azure.GetLoggedClient(client)
	vmic := virtualmachineimage.NewClient(client)

	ui.Message("Retrieving VM image...")
	var image virtualmachineimage.VMImage
	if err = retry.ExecuteOperation(func() error {
		imageList, err := vmic.ListVirtualMachineImages(
			virtualmachineimage.ListParameters{
				Category: virtualmachineimage.CategoryUser,
			})
		if err != nil {
			return err
		}

		for _, i := range imageList.VMImages {
			if i.Name == name {
				image = i
				break
			}
		}
		return nil
	}); err != nil {
		log.Printf("VM image client returned error: %s", err)
		return nil, false, err
	}
	if image.Name != name {
		return nil, false, fmt.Errorf("Could not find image: %s", name)
	}

	ui.Message(fmt.Sprintf("Deleting VM image (keeping VHDs) %s: %s...", image.Name, image.Label))
	err = retry.ExecuteOperation(func() error { return vmic.DeleteVirtualMachineImage(image.Name, false) })
	if err != nil {
		log.Printf("Error deleting VM image: %s", err)
		return nil, false, err
	}

	blobs := VMBlobListArtifact{
		OSDisk:    image.OSDiskConfiguration.MediaLink,
		DataDisks: make([]string, len(image.DataDiskConfigurations))}

	for i, ddc := range image.DataDiskConfigurations {
		blobs.DataDisks[i] = ddc.MediaLink
	}

	return blobs, false, nil
}

type VMBlobListArtifact struct {
	OSDisk    string
	DataDisks []string
}

const BuilderID = azure.BuilderId + "-vhdonly"

func (VMBlobListArtifact) BuilderId() string { return BuilderID }
func (VMBlobListArtifact) Destroy() error    { return nil }
func (VMBlobListArtifact) Files() []string   { return nil }
func (a VMBlobListArtifact) Id() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(a.String())))
}
func (VMBlobListArtifact) State(string) interface{} { return nil }
func (a VMBlobListArtifact) String() string {
	d, err := json.Marshal(&a)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return string(d)
}
