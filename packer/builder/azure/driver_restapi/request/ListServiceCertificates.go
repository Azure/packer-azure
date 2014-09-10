package request

import (
	"fmt"
)

func (m *Manager) ListServiceCertificates(serviceName string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/certificates",  m.SubscrId, serviceName)

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
