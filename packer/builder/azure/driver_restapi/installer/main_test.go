package main

import (
	"testing"
	"io/ioutil"
	"fmt"
	"encoding/json"
)

const (
	srcPathTest string = "d:/Packer.io/PackerLinux/installer/config-template"
	dstPathTest string = "d:/Packer.io/PackerLinux/installer/.packerconfig"
	dstPathModTest string = "d:/Packer.io/PackerLinux/installer/.packerconfigMod"
)

func TestModifyCoreConfig(t *testing.T) {
	err := ModifyCoreConfig(srcPathTest, dstPathTest)
	if err != nil {
		t.Error(err.Error())
	}
	t.Error("eom")
}

func _TestJSON(t *testing.T) {

	srcData, err := ioutil.ReadFile(srcPathTest)
	if err != nil {
		t.Error(err.Error())
	}
	srcConf := Config{}
	json.Unmarshal(srcData, &srcConf)
	fmt.Println(srcConf)
	fmt.Println(srcConf.Builders["azure1"])

	dstData, err := ioutil.ReadFile(dstPathTest)
	if err != nil {
		t.Error(err.Error())

	}
	dstConf := Config{}
	json.Unmarshal(dstData, &dstConf)
	fmt.Println(dstConf)
	fmt.Println(dstConf.Builders["hyperv-iso"])

	mod := false
	for sk, sv := range srcConf.Builders {
		dv, ok := dstConf.Builders[sk]
		if ok && sv == dv {
			continue
		}
		if dstConf.Builders == nil{
			dstConf.Builders = map[string]string{sk:sv}
		} else {
			dstConf.Builders[sk] = sv
		}
		mod = true
	}

	for sk, sv := range srcConf.Provisioners {
		dv, ok := dstConf.Provisioners[sk]
		if ok && sv == dv {
			continue
		}
		if dstConf.Provisioners == nil{
			dstConf.Provisioners = map[string]string{sk:sv}
		} else {
			dstConf.Provisioners[sk] = sv
		}
		mod = true
	}

	fmt.Println(dstConf)

	if mod {
		dstDataMod, err := json.Marshal(dstConf)
		if err != nil {
			t.Error(err.Error())
		}

		err = ioutil.WriteFile(dstPathModTest, dstDataMod, 0600)
		if err != nil {
			t.Error(err.Error())
		}
	}

	t.Error("eom")
}
