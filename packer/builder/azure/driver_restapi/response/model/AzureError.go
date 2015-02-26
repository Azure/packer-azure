// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model

import (
	"encoding/xml"
	"fmt"
)

type AzureError struct {
	XMLName xml.Name `xml:"Error"`
	Code    string
	Message string
}

func (err AzureError) Error() string {
	return fmt.Sprintf("Azure error (%v): %v", err.Code, err.Message)
}
