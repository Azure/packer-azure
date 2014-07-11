// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package azure

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	msbldcommon "github.com/MSOpenTech/packer-azure/packer/builder/common"
)

type StepSelectSubscription struct {
	SubscriptionName string
	StorageAccount string
}

func (s *StepSelectSubscription) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(msbldcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error Setting Azure Subscription: %s"

	ui.Say("Setting Azure Subscription...")

	var err error

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$subscriptionName = '" + s.SubscriptionName + "';")
	blockBuffer.WriteString("$storageAccount = '" + s.StorageAccount + "';")
	blockBuffer.WriteString("Set-AzureSubscription -SubscriptionName $subscriptionName –CurrentStorageAccount $storageAccount")
	blockBuffer.WriteString("}")

	err = driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$subscriptionName = '" + s.SubscriptionName + "';")
	blockBuffer.WriteString("Select-AzureSubscription -SubscriptionName $subscriptionName -Default")
	blockBuffer.WriteString("}")

	err = driver.Exec( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepSelectSubscription) Cleanup(state multistep.StateBag) {
	// do nothing
}
