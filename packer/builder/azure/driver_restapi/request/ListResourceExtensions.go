// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
)

func (m *Manager) ListResourceExtensions() *Data {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/resourceextensions", m.SubscrId)

	data := &Data{
		Verb: "GET",
		Uri:  uri,
	}

	return data
}
