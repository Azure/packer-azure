// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	"github.com/Azure/azure-sdk-for-go/management"
	"log"
	"os"
	"strconv"
	"time"
)

// This environment variable determines if we log Azure HTTP
// requests and after how many characters the request/response data is chopped.
const logLengthKey = "PACKER_LOG_AZURE_MAXLEN"

const logPrefix = "[AZURE]"

func GetLoggedClient(client management.Client) management.Client {
	logLengthStr := os.Getenv(logLengthKey)
	if logLengthStr == "" {
		log.Printf("%s %s not set, not logging", logPrefix, logLengthKey)
		return client
	}
	maxlen, err := strconv.ParseInt(logLengthStr, 10, 64)
	if err != nil {
		log.Printf("%s WARNING: Found %s in environment, but %s is not an integer?", logPrefix, logLengthKey, logLengthStr)
		return client
	}

	log.Printf("%s Azure requests will be logged", logPrefix)

	return loggedAzureClient{client, maxlen}
}

type loggedAzureClient struct {
	management.Client
	maxlen int64
}

func (loggedAzureClient) logRequest(method, url string, err error) {
	if err == nil {
		log.Printf("%s %s %s", logPrefix, method, url)
	} else {
		// keep lines together
		log.Printf("%s %s %s\n%s ERROR: %v", logPrefix, method, url, logPrefix, err)
	}
}

func (c loggedAzureClient) logRequestBody(data []byte) {
	log.Printf("%s >>> %s", logPrefix, c.chop(data))
}

func (c loggedAzureClient) logResponseBody(data []byte) {
	log.Printf("%s <<< %s", logPrefix, c.chop(data))
}

func (c loggedAzureClient) chop(data []byte) string {
	s := string(data)
	if int64(len(s)) > c.maxlen {
		s = s[:c.maxlen-3] + "..."
	}
	return s
}

func (c loggedAzureClient) logOperationID(oid management.OperationID) {
	log.Printf("%s <<< operation id: %s", logPrefix, oid)
}

func (c loggedAzureClient) SendAzureGetRequest(url string) ([]byte, error) {
	d, err := c.Client.SendAzureGetRequest(url)
	c.logRequest("GET", url, err)
	c.logResponseBody(d)
	return d, err
}

func (c loggedAzureClient) SendAzurePostRequest(url string, data []byte) (management.OperationID, error) {
	oid, err := c.Client.SendAzurePostRequest(url, data)
	c.logRequest("POST", url, err)
	c.logRequestBody(data)
	c.logOperationID(oid)
	return oid, err
}

func (c loggedAzureClient) SendAzurePutRequest(url, contentType string, data []byte) (management.OperationID, error) {
	oid, err := c.Client.SendAzurePutRequest(url, contentType, data)
	c.logRequest("PUT", url, err)
	c.logRequestBody(data)
	c.logOperationID(oid)
	return oid, err
}

func (c loggedAzureClient) SendAzureDeleteRequest(url string) (management.OperationID, error) {
	oid, err := c.Client.SendAzureDeleteRequest(url)
	c.logRequest("DELETE", url, err)
	c.logOperationID(oid)
	return oid, err
}

// func (c loggedAzureClient) GetOperationStatus(operationID management.OperationID) (management.GetOperationStatusResponse, error) {
// 	response, err := c.Client.GetOperationStatus(operationID)
// 	return response, err
// }

func (c loggedAzureClient) WaitForOperation(operationID management.OperationID, cancel chan struct{}) error {
	log.Printf("%s WaitForOperation( %s ) - begin", logPrefix, operationID)
	start := time.Now()
	err := c.Client.WaitForOperation(operationID, cancel)
	log.Printf("%s WaitForOperation( %s ) - end, duration: %v, err: %v", logPrefix, operationID, time.Now().Sub(start), err)
	return err
}
