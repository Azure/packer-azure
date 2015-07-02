package azure

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"testing"
)

func TestConfig_newConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	log.SetOutput(testLogger{t}) // hide log if test is succes
	cfg, _, err := newConfig(map[string]string{
		"subscription_name":         "My subscription",
		"publish_settings_path":     f.Name(),
		"storage_account":           "mysa",
		"storage_account_container": "vhdz",
		"os_type":                   "Linux",
		"os_image_label":            "Ubuntu_14.04",
		"location":                  "Central US",
		"instance_size":             "Large",
		"user_image_label":          "boo",
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedPattern := `^boo_\d{4}-\d{2}-\d{2}_\d{2}-\d{2}$`
	if !regexp.MustCompile(expectedPattern).MatchString(cfg.userImageName) {
		t.Errorf("expected %q to match %q", cfg.userImageName, expectedPattern)
	}
}

type testLogger struct{ *testing.T }

func (t testLogger) Write(d []byte) (int, error) {
	t.Logf("log: %s", string(d))
	return len(d), nil
}
