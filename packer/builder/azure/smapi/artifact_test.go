// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	. "gopkg.in/check.v1"
)

type ArtifactSuite struct{}

var _ = Suite(&ArtifactSuite{})

func (s *ArtifactSuite) Test_State(c *C) {
	a := artifact{
		publishSettingsPath: "publishSettingsPath",
		subscriptionID:      "subscriptionID",
	}

	c.Check(a.State("publishSettingsPath").(string), Equals, "publishSettingsPath")
	c.Check(a.State("subscriptionID").(string), Equals, "subscriptionID")
}

func (s *ArtifactSuite) Test_BuilderId(c *C) {
	a := artifact{}
	c.Check(a.BuilderId(), Equals, "Azure.ServiceManagement.VMImage")
}
