package azure

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"testing"
)

func getDefaultTestConfig(publishSettingsFileName string) map[string]string {
	return map[string]string{
		"subscription_name":         "My subscription",
		"publish_settings_path":     publishSettingsFileName,
		"storage_account":           "mysa",
		"storage_account_container": "vhdz",
		"os_type":                   "Linux",
		"os_image_label":            "Ubuntu_14.04",
		"location":                  "Central US",
		"instance_size":             "Large",
		"user_image_label":          "boo",
	}
}

func TestConfig_newConfig(t *testing.T) {
	f := getTempFile(t)
	defer os.Remove(f)
	log.SetOutput(testLogger{t}) // hide log if test is succes

	cfg, _, err := newConfig(getDefaultTestConfig(f))
	if err != nil {
		t.Fatal(err)
	}

	expectedPattern := `^boo_\d{4}-\d{2}-\d{2}_\d{2}-\d{2}$`
	if !regexp.MustCompile(expectedPattern).MatchString(cfg.userImageName) {
		t.Errorf("expected %q to match %q", cfg.userImageName, expectedPattern)
	}
}

func TestConfig_VNet(t *testing.T) {
	f := getTempFile(t)
	defer os.Remove(f)
	log.SetOutput(testLogger{t}) // hide log if test is succes

	tcs := []struct {
		cfgmod func(map[string]string)
		err    bool
	}{
		{func(cfg map[string]string) { cfg["vnet"] = "vnetName" }, true},
		{func(cfg map[string]string) { cfg["subnet"] = "subnetName" }, true},
		{func(cfg map[string]string) { cfg["vnet"] = "vnetName"; cfg["subnet"] = "subnetName" }, false},
	}

	for _, tc := range tcs {
		cfgmap := getDefaultTestConfig(f)
		tc.cfgmod(cfgmap)
		_, _, err := newConfig(cfgmap)
		if (err != nil) != tc.err {
			t.Fatalf("unexpected error value: %v", err)
		}
	}
}

func TestConfig_VMImageSource(t *testing.T) {
	f := getTempFile(t)
	defer os.Remove(f)
	log.SetOutput(testLogger{t}) // hide log if test is succes

	tcs := []struct {
		cfgmod func(map[string]string)
		err    bool
	}{
		{func(cfg map[string]string) { cfg["remote_source_image_link"] = "http://www.microsoft.com/" }, true},
		{func(cfg map[string]string) {
			cfg["remote_source_image_link"] = "http://www.microsoft.com/"
			delete(cfg, "os_image_label")
		}, false},
		{func(cfg map[string]string) { delete(cfg, "os_image_label") }, true},
	}

	for _, tc := range tcs {
		cfgmap := getDefaultTestConfig(f)
		tc.cfgmod(cfgmap)
		_, _, err := newConfig(cfgmap)
		if (err != nil) != tc.err {
			t.Fatalf("unexpected error value: %v", err)
		}
	}
}

func getTempFile(t *testing.T) string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

type testLogger struct{ *testing.T }

func (t testLogger) Write(d []byte) (int, error) {
	t.Logf("log: %s", string(d))
	return len(d), nil
}
