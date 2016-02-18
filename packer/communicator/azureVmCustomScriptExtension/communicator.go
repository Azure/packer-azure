// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azureVmCustomScriptExtension

import (
	"encoding/base64"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pborman/uuid"

	"github.com/Azure/packer-azure/packer/builder/azure/common"
	"github.com/Azure/packer-azure/packer/builder/azure/smapi/retry"

	"github.com/Azure/azure-sdk-for-go/management"
	vm "github.com/Azure/azure-sdk-for-go/management/virtualmachine"
	"github.com/Azure/azure-sdk-for-go/management/vmutils"
	"github.com/Azure/azure-sdk-for-go/storage"
)

const extPublisher = "Microsoft.Compute"
const extName = "CustomScriptExtension"

type comm struct {
	config Config
	uris   string
}

type Config struct {
	ServiceName               string
	VmName                    string
	StorageAccountName        string
	StorageAccountKey         string
	ContainerName             string
	Ui                        packer.Ui
	ProvisionTimeoutInMinutes uint
	ManagementClient          management.Client

	blobClient storage.BlobStorageClient
}

func New(config Config) (result *comm, err error) {
	storageClient, err := storage.NewBasicClient(config.StorageAccountName, config.StorageAccountKey)
	if err != nil {
		return nil, err
	}
	config.blobClient = storageClient.GetBlobService()

	result = &comm{
		config: config,
	}
	return
}

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
	ext, err := c.requestCustomScriptExtension()
	if err != nil {
		return
	}

	nameOfReference := fmt.Sprintf("PackerCustomScriptExtension-%s", uuid.New())
	nameOfPublisher := extPublisher
	nameOfExtension := extName
	versionOfExtension := ext.Version

	log.Println("Installing CustomScriptExtension...")
	state := "enable"
	params := c.buildParams(cmd.Command)

	err = c.updateRoleResourceExtension(nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state, params)
	if err != nil {
		return
	}

	stdOutBuff, stdErrBuff, err := c.pollCustomScriptExtensionIsReady()
	if err != nil {
		return
	}

	_, err = cmd.Stdout.Write([]byte(stdOutBuff))
	if err != nil {
		err = fmt.Errorf("cmd.Stdout error: %s", err.Error())
		return
	}

	_, err = cmd.Stderr.Write([]byte(stdErrBuff))
	if err != nil {
		err = fmt.Errorf("cmd.Stdout error: %s", err.Error())
		return
	}

	log.Println("Uninstalling CustomScriptExtension...")

	state = "uninstall"
	params = nil
	err = c.updateRoleResourceExtension(nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state, params)
	if err != nil {
		return
	}

	c.sleepSec(20)

	err = c.pollCustomScriptIsUninstalled()
	if err != nil {
		return
	}

	return
}

func (c *comm) sleepSec(d time.Duration) {
	log.Printf("Sleep for %v sec", uint(d))
	time.Sleep(time.Second * d)
}

func (c *comm) requestCustomScriptExtension() (*vm.ResourceExtension, error) {

	log.Println("Requesting resource extensions...")

	list, err := vm.NewClient(c.config.ManagementClient).GetResourceExtensions()
	if err != nil {
		return nil, err
	}

	log.Println("Searching for CustomScriptExtension...")
	for _, ext := range list {
		if ext.Name == "CustomScriptExtension" {
			log.Printf("CustomScriptExtension: %v\n\n", ext)
			return &ext, nil
		}
	}

	return nil, fmt.Errorf("Couldn't find CustomScriptExtension, am I too old or is Azure broken?")
}

func (c *comm) updateRoleResourceExtension(
	nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state string,
	params []vm.ResourceExtensionParameter) error {

	client := c.config.ManagementClient

	serviceName := c.config.ServiceName
	vmName := c.config.VmName

	log.Println("Updating Role Resource Extension...")

	role := vm.Role{}
	vmutils.AddAzureVMExtensionConfiguration(&role,
		nameOfExtension, nameOfPublisher, versionOfExtension, nameOfReference, state, []byte{}, []byte{})
	// HACK-paulmey: clean up later
	(*role.ResourceExtensionReferences)[0].ParameterValues = params

	if err := retry.ExecuteAsyncOperation(client, func() (management.OperationID, error) {
		return vm.NewClient(client).UpdateRole(serviceName, vmName, vmName, role)
	}); err != nil {
		return err
	}

	return nil
}

func (c *comm) buildParams(runScript string) []vm.ResourceExtensionParameter {

	scriptfile := "{\"fileUris\": [" + c.uris + "], \"commandToExecute\":\"powershell -ExecutionPolicy Unrestricted -file " + runScript + "\"}"
	account := "{\"storageAccountName\":\"" + c.config.StorageAccountName + "\",\"storageAccountKey\": \"" + c.config.StorageAccountKey + "\"}"

	return []vm.ResourceExtensionParameter{
		vm.ResourceExtensionParameter{
			Key:   "CustomScriptExtensionPublicConfigParameter",
			Value: base64.StdEncoding.EncodeToString([]byte(scriptfile)),
			Type:  "Public",
		},
		vm.ResourceExtensionParameter{
			Key:   "CustomScriptExtensionPrivateConfigParameter",
			Value: base64.StdEncoding.EncodeToString([]byte(account)),
			Type:  "Private",
		},
	}
}

func (c *comm) pollCustomScriptExtensionIsReady() (stdOutBuff, stdErrBuff string, err error) {
	client := c.config.ManagementClient
	log.Println("Polling CustomScriptExtension is ready. It may take some time...")

	var deployment vm.DeploymentResponse
	var extStatus *vm.ResourceExtensionStatus // paulmey-BUG #58: should not be pointer
	const statusSuccess = "Success"
	const statusError = "Error"

	//	needUpdateStatus := true

	serviceName := c.config.ServiceName
	vmName := c.config.VmName
	var timeout int64 = int64(c.config.ProvisionTimeoutInMinutes * 60)

	startTime := time.Now().Unix()
	timeoutState := false

	for {
		if timeout != 0 && time.Now().Unix()-startTime > timeout {
			timeoutState = true
			break
		}

		for {
			if timeout != 0 && time.Now().Unix()-startTime > timeout {
				timeoutState = true
				break
			}

			deployment, err = vm.NewClient(client).GetDeployment(serviceName, vmName)

			if err != nil {
				return
			}

			if deployment.RoleInstanceList[0].InstanceStatus == vm.InstanceStatusReadyRole {
				if len(deployment.RoleInstanceList[0].ResourceExtensionStatusList) > 0 {
					break
				}
			}

			c.sleepSec(45)
		}

		if timeoutState {
			err = fmt.Errorf("InstanceStatus is not 'ReadyRole' or CustomScriptExtension ResourceExtensionStatusList is empty after %d minutes", c.config.ProvisionTimeoutInMinutes)
			return
		}

		extHandlerName := extPublisher + "." + extName

		for _, s := range deployment.RoleInstanceList[0].ResourceExtensionStatusList {
			if s.HandlerName == extHandlerName {
				extStatus = &s
			}
		}

		if extStatus == nil {
			err = fmt.Errorf("CustomScriptExtension status not found")
			return
		}

		log.Printf("CustomScriptExtension status: %v", extStatus)

		extensionSettingStatus := extStatus.ExtensionSettingStatus

		if extensionSettingStatus.Status == statusError {
			err = fmt.Errorf("CustomScriptExtension operation '%s' status: %s", extensionSettingStatus.Operation, extensionSettingStatus.FormattedMessage.Message)
			return
		}

		log.Printf("CustomScriptExtension INFO: operation '%s' status: %s", extensionSettingStatus.Operation, extensionSettingStatus.Status)

		var stdOut, stdErr string

		for _, subStatus := range extStatus.ExtensionSettingStatus.SubStatusList {
			if subStatus.Name == "StdOut" {
				if subStatus.Status != statusSuccess {
					stdOut = fmt.Sprintf("StdOut failed with message: '%s'", subStatus.FormattedMessage.Message)
				} else {
					stdOut = subStatus.FormattedMessage.Message
				}
				continue
			}

			if subStatus.Name == "StdErr" {
				if subStatus.Status != statusSuccess {
					stdErr = fmt.Sprintf("StdErr failed with message: '%s'", subStatus.FormattedMessage.Message)
				} else {
					stdErr = subStatus.FormattedMessage.Message
				}
				continue
			}
		}

		log.Printf("StdOut: '%s'\n", stdOut)

		if len(stdOutBuff) == 0 {
			stdOutBuff = stdOut
		} else {
			stdOutBuff = common.GlueStrings(stdOutBuff, stdOut)
		}

		if len(stdErrBuff) == 0 {
			stdErrBuff = stdErr
		} else {
			stdErrBuff = common.GlueStrings(stdErrBuff, stdErr)
		}

		if extensionSettingStatus.Status == statusSuccess {
			break
		}

		c.sleepSec(40)
	}

	if timeoutState {
		err = fmt.Errorf("extensionSettingStatus.Status in not 'Success' after %d minutes", c.config.ProvisionTimeoutInMinutes)
		return
	}

	return
}

func (c *comm) pollCustomScriptIsUninstalled() error {
	client := c.config.ManagementClient
	log.Println("Polling CustomScript is uninstalled. It may take some time...")

	serviceName := c.config.ServiceName
	vmName := c.config.VmName

	const attemptLimit uint = 30
	repeatCount := attemptLimit
	for ; repeatCount > 0; repeatCount-- {
		deployment, err := vm.NewClient(client).GetDeployment(serviceName, vmName)
		if err != nil {
			return err
		}

		if deployment.RoleInstanceList[0].InstanceStatus == vm.InstanceStatusReadyRole {
			if len(deployment.RoleInstanceList[0].ResourceExtensionStatusList) == 0 {
				break
			}
		}

		c.sleepSec(45)
	}

	if repeatCount == 0 {
		err := fmt.Errorf("InstanceStatus is not 'ReadyRole' or ResourceExtensionStatusList is not empty after %d attempts", attemptLimit)
		return err
	}

	return nil
}

func (c *comm) Upload(string, io.Reader, *os.FileInfo) error {
	return fmt.Errorf("Upload is not supported for azureVmCustomScriptExtension")
}

func (c *comm) UploadDir(skipped string, src string, excl []string) error {

	src = filepath.FromSlash(src)
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	containerName := c.config.ContainerName

	if info.IsDir() {
		log.Println(fmt.Sprintf("Uploading files (only!) in the folder to Azure storage container '%s' => '%s'...", src, containerName))
		err := c.uploadFolder("", src)
		if err != nil {
			return err
		}
	} else {
		err := c.uploadFile("", src)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *comm) Download(string, io.Writer) error {
	return fmt.Errorf("Download is not supported for azureVmCustomScriptExtension")
}

// region private helpers

// Helper method to upload a blob, removed from azure-sdk (and not added back as util)
// See https://github.com/Azure/azure-sdk-for-go/commit/bf6723e87ca7c26e523b13c18f29383da60ab64f

// PutBlockBlob uploads given stream into a block blob by splitting
// data stream into chunks and uploading as blocks. Commits the block
// list at the end. This is a helper method built on top of PutBlock
// and PutBlockList methods with sequential block ID counting logic.
func putBlockBlob(b storage.BlobStorageClient, container, name string, blob io.Reader, chunkSize int) error {
	if chunkSize <= 0 || chunkSize > storage.MaxBlobBlockSize {
		chunkSize = storage.MaxBlobBlockSize
	}

	chunk := make([]byte, chunkSize)
	n, err := blob.Read(chunk)
	if err != nil && err != io.EOF {
		return err
	}

	blockList := []storage.Block{}

	for blockNum := 0; ; blockNum++ {
		id := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%011d", blockNum)))
		data := chunk[:n]
		err = b.PutBlock(container, name, id, data)
		if err != nil {
			return err
		}

		blockList = append(blockList, storage.Block{ID: id, Status: storage.BlockStatusLatest})

		// Read next block
		n, err = blob.Read(chunk)
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}
	}

	return b.PutBlockList(container, name, blockList)
}

func (c *comm) uploadFile(dscPath string, srcPath string) error {

	srcPath = filepath.FromSlash(srcPath)

	_, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("Check file path is correct: %s", srcPath)
	}

	ui := c.config.Ui
	sa := c.config.blobClient

	containerName := c.config.ContainerName

	fileName := filepath.Base(srcPath)
	uri := sa.GetBlobURL(containerName, fileName)

	if len(c.uris) == 0 {
		c.uris = fmt.Sprintf("\"%s\"", uri)
	} else {
		c.uris += fmt.Sprintf(", \"%s\"", uri)
	}

	log.Println("uris: '" + c.uris + "'")

	ui.Message(fmt.Sprintf("Uploading file to to Azure storage container '%s' => '%s'...", srcPath, containerName))

	f, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %v", srcPath, err)
	}

	defer f.Close()

	err = sa.CreateBlockBlob(containerName, fileName)
	if err != nil {
		return fmt.Errorf("Error creating block blob in container %s for file %s: %v", containerName, fileName, err)
	}

	err = putBlockBlob(sa, containerName, fileName, f, storage.MaxBlobBlockSize)
	if err != nil {
		return fmt.Errorf("Error uploading block blob in container %s for file %s: %v", containerName, fileName, err)
	}

	return nil
}

func (c *comm) uploadFolder(dscPath string, srcPath string) error {

	srcPath = filepath.FromSlash(srcPath)

	files, err := ioutil.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		err := c.uploadFile("", filepath.Join(srcPath, f.Name()))
		if err != nil {
			return err
		}
	}

	return err
}

func (*comm) DownloadDir(string, string, []string) error {
	return fmt.Errorf("Not implemented")
}
