// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package azureVmCustomScriptExtension

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"path/filepath"
	"os"
	"io/ioutil"
	"encoding/base64"
	"log"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	azureservice "github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/request"
	storageservice "github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/storage_service/request"
	"time"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/utils"
)

type comm struct {
	config *Config
	uris string
}

type Config struct {
	ServiceName string
	VmName	string
	StorageServiceDriver *storageservice.StorageServiceDriver
	AzureServiceRequestManager *azureservice.Manager
	ContainerName string
	Ui packer.Ui
}

func New(config *Config) (result *comm, err error) {
	result = &comm{
		config: config,
	}

	return
}

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
	ui := c.config.Ui

	reqManager := c.config.AzureServiceRequestManager

	ui.Message("Requesting resource extentions...")
	requestData := reqManager.ListResourceExtensions()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		return
	}

	list, err := response.ParseResourceExtensionList(resp.Body)

	if err != nil {
		return
	}

	ui.Message("Searching for CustomScriptExtension...")
	ext := list.FirstOrNull("CustomScriptExtension")
	log.Printf("CustomScriptExtension: %v\n\n", ext)

	if ext == nil {
		err = fmt.Errorf("CustomScriptExtension is nil")
		return
	}

	serviceName := c.config.ServiceName
	vmName := c.config.VmName
	nameOfReference := "PackerCSE"
	nameOfPublisher := ext.Publisher
	nameOfExtension := ext.Name
	versionOfExtension := ext.Version
	state := "enable"

	storageAccountName, storageAccountKey := c.config.StorageServiceDriver.GetProps()

	account := "{\"storageAccountName\":\"" + storageAccountName + "\",\"storageAccountKey\": \"" + storageAccountKey + "\"}";
	runScript := cmd.Command

	scriptfile := "{\"fileUris\": [" + c.uris + "], \"commandToExecute\":\"powershell -ExecutionPolicy Unrestricted -file " + runScript + "\"}"

	params := []azureservice.ResourceExtensionParameterValue {
		azureservice.ResourceExtensionParameterValue{
			Key: "CustomScriptExtensionPublicConfigParameter",
			Value: base64.StdEncoding.EncodeToString([]byte(scriptfile)),
			Type: "Public",
		},
		azureservice.ResourceExtensionParameterValue{
			Key: "CustomScriptExtensionPrivateConfigParameter",
			Value: base64.StdEncoding.EncodeToString([]byte(account)),
			Type: "Private",
		},
	}
	ui.Message("Updating Role Resource Extension...")
	requestData = reqManager.UpdateRoleResourceExtensionReference(serviceName, vmName, nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state, params)
	err = reqManager.ExecuteSync(requestData)

	if err != nil {
		return
	}

	ui.Message("Polling the VM is ready. It may take some time...")

	var deployment *model.Deployment
	var res *model.ResourceExtensionStatus
	const statusSuccess = "Success"
	const statusError = "Error"
	var stdOutBuff, stdErrBuff string

	needUpdateStatus := true
//	errorIgnoreCount := 10

	for needUpdateStatus {
		repeatCount := 30
		for ; repeatCount > 0; repeatCount-- {
			requestData = reqManager.GetDeployment(serviceName, vmName)
			resp, err = reqManager.Execute(requestData)

			if err != nil {
/*
				log.Printf("Checking Result error: '%s'", err.Error())
				pattern := "Request needs to have a x-ms-version header"
				errString := err.Error()
				// Sometimes server returns strange error - ignore it
				match, _ := regexp.MatchString(pattern, errString)
				if match {
					log.Println("Checking Result ignore error: " + errString)
					errorIgnoreCount--
					if errorIgnoreCount == 0 {
						return
					}
					continue
				}
*/

				return
			}

			deployment, err = response.ParseDeployment(resp.Body)

			if err != nil {
				return
			}

			if len(deployment.RoleInstanceList[0].ResourceExtensionStatusList) > 0 {
				break
			}

			var d time.Duration = 30
			log.Printf("Sleep for %v", d)
			time.Sleep(time.Second*d)
		}

		if repeatCount == 0 {
			err = fmt.Errorf("CustomScriptExtension ResourceExtensionStatusList is empty")
			return
		}

//		log.Printf("ResourceExtensionStatusList: %v", deployment.RoleInstanceList[0].ResourceExtensionStatusList)

		extHandlerName := ext.Publisher + "." + ext.Name

		for _, s := range deployment.RoleInstanceList[0].ResourceExtensionStatusList {
			if s.HandlerName == extHandlerName {
				res = &s
			}
		}

		if res == nil {
			err = fmt.Errorf("CustomScriptExtension status not found")
			return
		}

//		log.Printf("CustomScriptExtension status: %v", res)

		extensionSettingStatus := res.ExtensionSettingStatus

		if extensionSettingStatus.Status == statusError {
			err = fmt.Errorf("CustomScriptExtension operation '%s' status: %s", extensionSettingStatus.Operation, extensionSettingStatus.Status )
			return
		}

		log.Printf("CustomScriptExtension INFO: operation '%s' status: %s",extensionSettingStatus.Operation, extensionSettingStatus.Status)

		var stdOut, stdErr string

		for _, subStatus := range res.ExtensionSettingStatus.SubStatusList {
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
			stdOutBuff = utils.Clue(stdOutBuff, stdOut)
		}

//		log.Printf("stdOutBuff: '%s'\n", stdOutBuff)

		if len(stdErrBuff) == 0 {
			stdErrBuff = stdErr
		} else {
			stdErrBuff = utils.Clue(stdErrBuff, stdErr)
		}

		if extensionSettingStatus.Status == statusSuccess {
			needUpdateStatus = false
			break
		}

		var d time.Duration = 30
		log.Printf("Sleep for %v", d)
		time.Sleep(time.Second*d)
	}

	_, err = cmd.Stdout.Write([]byte(stdOutBuff))
	if res == nil {
		err = fmt.Errorf("cmd.Stdout error: %s", err.Error())
		return
	}

	_, err = cmd.Stderr.Write([]byte(stdErrBuff))
	if res == nil {
		err = fmt.Errorf("cmd.Stdout error: %s", err.Error())
		return
	}

	return
}

func (c *comm)Upload(string, io.Reader, *os.FileInfo) error {
	return fmt.Errorf("Upload is not supported for azureVmCustomScriptExtension")
}

func (c *comm) UploadDir(dst string, src string, excl []string) error {

	src = filepath.FromSlash(src)
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	containerName := c.config.ContainerName

	if info.IsDir() {
		log.Println(fmt.Sprintf("Uploading files (only!) in the folder to Azure storage container '%s' => '%s'...",  src, containerName))
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

func (c *comm) uploadFile(dscPath string, srcPath string) error {

	srcPath = filepath.FromSlash(srcPath)

	_, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("Check file path is correct: %s", srcPath)
	}

	ui := c.config.Ui
	sa := c.config.StorageServiceDriver

	storageAccountName, _ := c.config.StorageServiceDriver.GetProps()
	containerName := c.config.ContainerName

	fileName := filepath.Base(srcPath)
	uri := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", storageAccountName, containerName, fileName)

	if len(c.uris) == 0 {
		c.uris = fmt.Sprintf("\"%s\"", uri)
	} else {
		c.uris += fmt.Sprintf(", \"%s\"", uri)
	}

	log.Println("uris: '" + c.uris + "'")

	ui.Message(fmt.Sprintf("Uploading file to to Azure storage container '%s' => '%s'...", srcPath, containerName))

	_, err = sa.PutBlob(containerName, srcPath)

	return err
}

func (c *comm) uploadFolder(dscPath string, srcPath string ) error {

	srcPath = filepath.FromSlash(srcPath)

	files, err := ioutil.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if (f.IsDir()) {
			continue
		}

		err := c.uploadFile("", filepath.Join(srcPath,f.Name()))
		if err != nil {
			return err
		}
	}

	return err
}

