package request

import (
	"fmt"
)

func (m *Manager) GetOperationStatus(requestId string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/operations/%s",  m.SubscrId, requestId)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2009-10-01",
	}

	data := &Data{
		Verb: "GET",
		Uri : uri,
		Headers: headers,
		Body : nil,
	}

	return data
}
