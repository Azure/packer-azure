package request

import (
	"fmt"
)

func (m *Manager) DeleteRole(serviceName, vmName string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/deployments/%s/roles/%s", m.SubscrId, serviceName, vmName, vmName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2012-03-01",
	}

	data := &Data{
		Verb: "DELETE",
		Uri : uri,
		Headers: headers,
		Body : nil,
	}

	return data
}
