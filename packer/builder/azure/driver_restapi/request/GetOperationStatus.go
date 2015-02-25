// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
	"log"
	"time"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
)

func (m *Manager) GetOperationStatus(requestId string) (*model.Operation, error) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/operations/%s", m.SubscrId, requestId)

	data := &Data{
		Verb: "GET",
		Uri:  uri,
	}

	res, err := m.Execute(data)
	if err != nil {
		return nil, err
	}

	return response.ParseOperation(res.Body)
}

var defaultInterval = time.Duration(2 * time.Second)

func (m *Manager) WaitForOperation(requestId string) (operation *model.Operation, err error) {
	log.Printf("Manager.WaitForOperation (%s)", requestId)

	starttime := time.Now()

	for {
		nextRequestAt := time.Now().Add(defaultInterval)

		operation, err = m.GetOperationStatus(requestId)
		if err != nil {
			return nil, err
		}

		if operation.Status != "InProgress" {
			log.Printf("Manager.WaitForOperation (%s) took %v to complete (%v)", requestId, time.Now().Sub(starttime), operation.Status)

			if operation.Status == "Failed" {
				return operation, operation.Error
			}

			return operation, nil
		}

		if wait := nextRequestAt.Sub(time.Now()); wait > 0 {
			log.Printf("Waiting %v before sending the next request...", wait)
			time.Sleep(wait)
		}
	}
}
