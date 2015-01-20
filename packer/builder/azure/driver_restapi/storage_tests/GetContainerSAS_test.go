// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package storage_tests

import (
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/storage_service/request"
	"testing"
	"time"
)

func TestGetContainerSas(t *testing.T) {

	errMassage := "TestGetContainerSas: %s\n"

	sa := request.NewStorageServiceDriver(g_accountName, g_secret)

	ts := time.Now().UTC()
	t.Logf("ts: " + ts.String())
	te := ts.Add(time.Hour * 24)
	t.Logf("te: " + te.String())

	signedstart := ts.Format(time.RFC3339)
	t.Logf("signedstart: " + signedstart)

	containerName := "images"
	sas, err := sa.GetContainerSAS(containerName)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	t.Logf("sas: " + sas)
}
