// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/driver"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/request"
)

var g_reqManager *request.Manager

func getRequestManager(t *testing.T) (*request.Manager, error) {
	psPath := os.Getenv("PUBLISHSETTINGS")
	if psPath == "" {
		t.Skip("PUBLISHSETTINGS environment variable not set, skipping this test.")
	}
	if g_reqManager != nil {
		return g_reqManager, nil
	}

	var d driver.IDriverRest
	var err error

	subscriptionInfo, err := driver_restapi.ParsePublishSettings(psPath, "PackerTestSubscription")

	if err != nil {
		return nil, fmt.Errorf("ParsePublishSettings error: %s\n", err.Error())
	}

	fmt.Println("id: " + subscriptionInfo.Id)

	d, err = driver.NewTlsDriver(subscriptionInfo.CertData)

	if err != nil {
		return nil, fmt.Errorf("NewTlsDriver error: %s\n", err.Error())
	}

	g_reqManager = &request.Manager{
		SubscrId: subscriptionInfo.Id,
		Driver:   d,
	}

	return g_reqManager, err
}
