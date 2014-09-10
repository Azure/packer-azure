package request

import (
	"fmt"
)

func (m *Manager) CheckStorageAccountNameAvailability(storageAccountName string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/storageservices/operations/isavailable/%s",  m.SubscrId, storageAccountName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2013-03-01",
	}

	data := &Data {
		Verb: "GET",
		Uri : uri,
		Headers: headers,
		Body : nil,
	}

	return data
}
