// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azuresmvhdonly

import (
	"errors"
	azure "github.com/Azure/packer-azure/packer/builder/azure/smapi"
	"github.com/mitchellh/packer/packer"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) Test_BuilderId(c *C) {
	a := packer.MockArtifact{BuilderIdValue: "bla"}

	sut := PostProcessor{}
	_, _, err := sut.PostProcess(testUi(c), &a)
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "Can only import from Azure builder"), Equals, true)

	a.BuilderIdValue = azure.BuilderId
	_, _, err = sut.PostProcess(testUi(c), &a)
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "Can only import from Azure builder"), Equals, false)
}

func (s *MySuite) Test_ArtifactCredentials(c *C) {
	a := packer.MockArtifact{
		BuilderIdValue: azure.BuilderId,
		StateValues:    make(map[string]interface{}),
	}

	checkErr := func() {
		sut := PostProcessor{}
		_, _, err := sut.PostProcess(testUi(c), &a)
		c.Assert(err, NotNil)
		c.Assert(err, ErrorMatches, ".* for this artifact. Make sure you used the same version of the builder as the post-processor.")
	}
	checkErr() // check for nil interface{}

	for _, creds := range []struct { // check for one or both values empty
		PublishSettingsPath string
		SubscriptionID      string
	}{
		{"", ""},
		{"one", ""},
		{"", "two"},
	} {
		a.StateValues["publishSettingsPath"] = creds.PublishSettingsPath
		a.StateValues["subscriptionID"] = creds.SubscriptionID
		checkErr()
	}
}

func testUi(c *C) *packer.BasicUi {
	return &packer.BasicUi{
		Reader:      nilReader{},
		Writer:      logFunc(func(s string) { c.Logf("UI: %s", s) }),
		ErrorWriter: logFunc(func(s string) { c.Logf("ERR: %s", s) }),
	}
}

type logFunc func(s string)

func (w logFunc) Write(d []byte) (int, error) {
	w(string(d))
	return len(d), nil
}

type nilReader struct{}

func (nilReader) Read([]byte) (int, error) {
	return 0, errors.New("Nothing to read here, go away.")
}
