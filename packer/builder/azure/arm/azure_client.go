// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	armStorage "github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/packer-azure/packer/builder/azure/common"
)

type AzureClient struct {
	storage.BlobStorageClient
	resources.DeploymentsClient
	resources.GroupsClient
	network.PublicIPAddressesClient
	compute.VirtualMachinesClient
	common.VaultClient
	armStorage.AccountsClient
}

func NewAzureClient(subscriptionID, resourceGroupName, storageAccountName string,
	servicePrincipalToken, servicePrincipalTokenVault *azure.ServicePrincipalToken) (*AzureClient, error) {

	var azureClient = &AzureClient{}

	azureClient.DeploymentsClient = resources.NewDeploymentsClient(subscriptionID)
	azureClient.DeploymentsClient.Authorizer = servicePrincipalToken

	azureClient.GroupsClient = resources.NewGroupsClient(subscriptionID)
	azureClient.GroupsClient.Authorizer = servicePrincipalToken

	azureClient.PublicIPAddressesClient = network.NewPublicIPAddressesClient(subscriptionID)
	azureClient.PublicIPAddressesClient.Authorizer = servicePrincipalToken

	azureClient.VirtualMachinesClient = compute.NewVirtualMachinesClient(subscriptionID)
	azureClient.VirtualMachinesClient.Authorizer = servicePrincipalToken

	azureClient.AccountsClient = armStorage.NewAccountsClient(subscriptionID)
	azureClient.AccountsClient.Authorizer = servicePrincipalToken

	azureClient.VaultClient = common.VaultClient{}
	azureClient.VaultClient.Authorizer = servicePrincipalTokenVault

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
