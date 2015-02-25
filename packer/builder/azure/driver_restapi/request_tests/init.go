// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request_tests

import (
	"os"
	"testing"

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

	g_reqManager, err := request.NewManager(psPath, "PackerTestSubscription")
	if err != nil {
		panic(err)
	}

	return g_reqManager, err
}
