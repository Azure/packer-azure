package azure

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

func findSubscriptionID(publishSettingsPath, subscriptionName string) (string, error) {
	data, err := ioutil.ReadFile(publishSettingsPath)
	if err != nil {
		return "", fmt.Errorf("Error reading publishsettings (%s): %v", publishSettingsPath, err)
	}

	var pubsettings struct {
		Subscriptions []struct {
			ID   string `xml:"Id,attr"`
			Name string `xml:",attr"`
		} `xml:"PublishProfile>Subscription"`
	}
	err = xml.Unmarshal(data, &pubsettings)
	if err != nil {
		return "", fmt.Errorf("Error deserializing publishsettings (%s): %v", publishSettingsPath, err)
	}

	for _, subscription := range pubsettings.Subscriptions {
		if subscription.Name == subscriptionName {
			return subscription.ID, nil
		}
	}

	return "", fmt.Errorf("Subscription with name %q not found in %s", subscriptionName, publishSettingsPath)
}
