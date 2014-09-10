package request

import (
	"fmt"
)

func (m *Manager) GetVmImages() (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/vmimages",  m.SubscrId)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2014-05-01",
	}

	data := &Data{
		Verb: "GET",
		Uri : uri,
		Headers: headers,
		Body : nil,
	}

	return data
}
