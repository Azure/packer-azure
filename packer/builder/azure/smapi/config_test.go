package azure

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"testing"
)

func getDefaultTestConfig(publishSettingsFileName string) map[string]interface{} {
	return map[string]interface{}{
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
		cfgmod func(map[string]interface{})
		err    bool
	}{
		{func(cfg map[string]interface{}) { cfg["vnet"] = "vnetName" }, true},
		{func(cfg map[string]interface{}) { cfg["subnet"] = "subnetName" }, true},
		{func(cfg map[string]interface{}) { cfg["vnet"] = "vnetName"; cfg["subnet"] = "subnetName" }, false},
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

	m := make(map[string]string)
	m["name"] = "Ubuntu-14_04_3-LTS-amd64-server-20160119-en-us-30GB"
	m["link"] = "http://www.microsoft.com/"

	tcs := []struct {
		cfgmod func(map[string]interface{})
		msg    string
		err    bool
	}{
		{func(cfg map[string]interface{}) { // none
			delete(cfg, "os_image_label")
		}, "No defined source", true},
		{func(cfg map[string]interface{}) { // label and link
			cfg["remote_source_image_link"] = m["link"]
		}, "Both label and link defined", true},
		{func(cfg map[string]interface{}) { // label and name
			cfg["os_image_name"] = m["name"]
		}, "Both label and name defined", true},
		{func(cfg map[string]interface{}) { // label and name and link
			cfg["os_image_name"] = m["name"]
			cfg["remote_source_image_link"] = m["link"]
		}, "Both name and link defined", true},
		{func(cfg map[string]interface{}) { // only os_image_name set
			delete(cfg, "os_image_label")
			cfg["os_image_name"] = m["name"]
		}, "Only name", false},
		{func(cfg map[string]interface{}) { // only remote_source_image_link set
			delete(cfg, "os_image_label")
			cfg["remote_source_image_link"] = m["link"]
		}, "Only link", false},
	}
	// Test default config
	_, _, err := newConfig(getDefaultTestConfig(f))
	if err != nil {
		t.Fatalf("TestCase Label only : unexpected error value: %v", err)
	}
	for _, tc := range tcs {
		cfgmap := getDefaultTestConfig(f)
		tc.cfgmod(cfgmap)
		_, _, err := newConfig(cfgmap)
		if (err != nil) != tc.err {
			t.Fatalf("TestCase %s : unexpected error value: %v", tc.msg, err)
		}
	}
}

// a short test that shows how a mixed-type array is processed by
// mapstructure
func TestMapStructureMixedArray(t *testing.T) {
	jsonstring := `{"array": ["string", 1, 2, "three", 4] }`
	var raw interface{}
	err := json.Unmarshal([]byte(jsonstring), &raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var cfg struct {
		Array []interface{}
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &cfg})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = decoder.Decode(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cfg.Array[0].(string); !ok {
		t.Errorf("value 0 is not a string: %v = %T", cfg.Array[0], cfg.Array[0])
	}
	if _, ok := cfg.Array[1].(float64); !ok {
		t.Errorf("value 1 is not a float64: %v = %T", cfg.Array[1], cfg.Array[1])
	}
	if _, ok := cfg.Array[2].(float64); !ok {
		t.Errorf("value 2 is not a float64: %v = %T", cfg.Array[2], cfg.Array[2])
	}
	if _, ok := cfg.Array[3].(string); !ok {
		t.Errorf("value 3 is not a string: %v = %T", cfg.Array[3], cfg.Array[3])
	}
	if _, ok := cfg.Array[4].(float64); !ok {
		t.Errorf("value 4 is not a float64: %v = %T", cfg.Array[4], cfg.Array[4])
	}
}

func TestConfig_Datadisks(t *testing.T) {
	f := getTempFile(t)
	defer os.Remove(f)
	log.SetOutput(testLogger{t}) // hide log if test is succes

	tcs := []struct {
		cfgmod func(map[string]interface{})
		err    bool
	}{
		{func(cfg map[string]interface{}) {
			cfg["data_disks"] = []interface{}{"http://blob/container/disk.vhd", float64(20), float64(1023)}
		}, false},
		{func(cfg map[string]interface{}) {
			cfg["data_disks"] = []interface{}{1.1}
		}, true},
		{func(cfg map[string]interface{}) {
			cfg["data_disks"] = map[string]interface{}{"key": "value"}
		}, true},
		{func(cfg map[string]interface{}) {
			cfg["data_disks"] = []interface{}{map[string]interface{}{"key": "value"}}
		}, true},
	}

	for n, tc := range tcs {
		t.Logf("============== starting test case %d =================", n)
		cfgmap := getDefaultTestConfig(f)
		tc.cfgmod(cfgmap)
		_, _, err := newConfig(cfgmap)
		if (err != nil) != tc.err {
			t.Fatalf("unexpected error value for test case %d: %v", n, err)
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
