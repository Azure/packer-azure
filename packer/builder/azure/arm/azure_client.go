// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	armStorage "github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/packer-azure/packer/builder/azure/common"
	"github.com/Azure/packer-azure/version"
	"math"
	"os"
	"strconv"
)

const (
	EnvPackerLogAzureMaxLen = "PACKER_LOG_AZURE_MAXLEN"
)

var (
	packerUserAgent = fmt.Sprintf(";packer/%s", version.Version)
)

type AzureClient struct {
	storage.BlobStorageClient
	resources.DeploymentsClient
	resources.GroupsClient
	network.PublicIPAddressesClient
	compute.VirtualMachinesClient
	common.VaultClient
	armStorage.AccountsClient

	InspectorMaxLength int
}

func NewAzureClient(subscriptionID, resourceGroupName, storageAccountName string,
	servicePrincipalToken, servicePrincipalTokenVault *azure.ServicePrincipalToken) (*AzureClient, error) {

	var azureClient = &AzureClient{}

	maxlen := getInspectorMaxLength()

	azureClient.DeploymentsClient = resources.NewDeploymentsClient(subscriptionID)
	azureClient.DeploymentsClient.Authorizer = servicePrincipalToken
	azureClient.DeploymentsClient.RequestInspector = withInspection(maxlen)
	azureClient.DeploymentsClient.ResponseInspector = byInspecting(maxlen)
	azureClient.DeploymentsClient.UserAgent += packerUserAgent

	azureClient.GroupsClient = resources.NewGroupsClient(subscriptionID)
	azureClient.GroupsClient.Authorizer = servicePrincipalToken
	azureClient.GroupsClient.RequestInspector = withInspection(maxlen)
	azureClient.GroupsClient.ResponseInspector = byInspecting(maxlen)
	azureClient.GroupsClient.UserAgent += packerUserAgent

	azureClient.PublicIPAddressesClient = network.NewPublicIPAddressesClient(subscriptionID)
	azureClient.PublicIPAddressesClient.Authorizer = servicePrincipalToken
	azureClient.PublicIPAddressesClient.RequestInspector = withInspection(maxlen)
	azureClient.PublicIPAddressesClient.ResponseInspector = byInspecting(maxlen)
	azureClient.PublicIPAddressesClient.UserAgent += packerUserAgent

	azureClient.VirtualMachinesClient = compute.NewVirtualMachinesClient(subscriptionID)
	azureClient.VirtualMachinesClient.Authorizer = servicePrincipalToken
	azureClient.VirtualMachinesClient.RequestInspector = withInspection(maxlen)
	azureClient.VirtualMachinesClient.ResponseInspector = byInspecting(maxlen)
	azureClient.PublicIPAddressesClient.UserAgent += packerUserAgent

	azureClient.AccountsClient = armStorage.NewAccountsClient(subscriptionID)
	azureClient.AccountsClient.Authorizer = servicePrincipalToken
	azureClient.AccountsClient.RequestInspector = withInspection(maxlen)
	azureClient.AccountsClient.ResponseInspector = byInspecting(maxlen)
	azureClient.AccountsClient.UserAgent += packerUserAgent

	azureClient.VaultClient = common.VaultClient{}
	azureClient.VaultClient.Authorizer = servicePrincipalTokenVault
	azureClient.VaultClient.RequestInspector = withInspection(maxlen)
	azureClient.VaultClient.ResponseInspector = byInspecting(maxlen)
	azureClient.VaultClient.UserAgent += packerUserAgent

	accountKeys, err := azureClient.AccountsClient.ListKeys(resourceGroupName, storageAccountName)
	if err != nil {
		return nil, err
	}

	storageClient, err := storage.NewBasicClient(storageAccountName, *accountKeys.Key1)
	if err != nil {
		return nil, err
	}

	azureClient.BlobStorageClient = storageClient.GetBlobService()
	return azureClient, nil
}

func getInspectorMaxLength() int64 {
	value, ok := os.LookupEnv(EnvPackerLogAzureMaxLen)
	if !ok {
		return math.MaxInt64
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	if i < 0 {
		return 0
	}

	return i
}
