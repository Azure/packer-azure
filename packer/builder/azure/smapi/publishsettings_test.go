package azure

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_findSubscriptionID(t *testing.T) {
	f, err := ioutil.TempFile("", "test")
	defer f.Close()
	if err != nil {
		t.Fatal(err)
	}
	filename := f.Name()
	defer os.Remove(filename)

	_, err = f.WriteString(`<PublishData>
  <PublishProfile
    SchemaVersion="2.0"
    PublishMethod="AzureServiceManagementAPI">
    <Subscription
      ServiceManagementUrl="https://management.core.windows.net"
      Id="2a6a0fd6-0fc1-11e5-96b2-1cc1de3246e5"
      Name="test sub 1"
      ManagementCertificate="MIIKPAIBAzCCCfwGC"/>
  </PublishProfile>
  <PublishProfile
    SchemaVersion="2.0"
    PublishMethod="AzureServiceManagementAPI">
    <Subscription
      ServiceManagementUrl="https://management.core.windows.net"
      Id="4ff520d8-0fc1-11e5-82fd-1cc1de3246e5"
      Name="test sub 2"
      ManagementCertificate="MIIKPAIBAzCCCfwGC"/>
    <Subscription
      ServiceManagementUrl="https://management.core.windows.net"
      Id="504ced04-0fc1-11e5-b46d-1cc1de3246e5"
      Name="test sub 3"
      ManagementCertificate="MIIKPAIBAzCCCfwGC"/>
  </PublishProfile>
</PublishData>`)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		in, out string
		err     bool
	}{
		{in: "test sub 1", out: "2a6a0fd6-0fc1-11e5-96b2-1cc1de3246e5"},
		{in: "test sub 2", out: "4ff520d8-0fc1-11e5-82fd-1cc1de3246e5"},
		{in: "test sub 3", out: "504ced04-0fc1-11e5-b46d-1cc1de3246e5"},
		{in: "test sub 4", err: true},
		{in: "", err: true},
	}

	for _, tc := range tcs {
		id, err := findSubscriptionID(filename, tc.in)
		if !(err == nil || tc.err) {
			t.Fatalf("Failed for %s: %v", tc.in, err)
		}
		if !tc.err && id != tc.out {
			t.Errorf("For %s: Expected %s, got %s", tc.in, tc.out, id)
		}
	}
}
