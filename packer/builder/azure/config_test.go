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

	tcs := []struct {
		cfgmod func(map[string]interface{})
		err    bool
	}{
		{func(cfg map[string]interface{}) { cfg["remote_source_image_link"] = "http://www.microsoft.com/" }, true},
		{func(cfg map[string]interface{}) {
			cfg["remote_source_image_link"] = "http://www.microsoft.com/"
			delete(cfg, "os_image_label")
		}, false},
		{func(cfg map[string]interface{}) { delete(cfg, "os_image_label") }, true},
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
