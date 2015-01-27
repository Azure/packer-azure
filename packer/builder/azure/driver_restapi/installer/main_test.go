// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

const (
	srcPathTest    string = "/config-template"
	dstPathTest    string = "/.packerconfig"
	dstPathModTest string = "/.packerconfigMod"
)

func TestModifyCoreConfig(t *testing.T) {
	path := os.Getenv("CONFIGTESTPATH")
	if path == "" {
		t.Skip("CONFIGTESTPATH environment variable not set, skipping this test.")
	}

	err := ModifyCoreConfig(path+srcPathTest, path+dstPathTest)
	if err != nil {
		t.Error(err.Error())
	}
}

func _TestJSON(t *testing.T) {
	path := os.Getenv("CONFIGTESTPATH")
	if path == "" {
		t.Skip("CONFIGTESTPATH environment variable not set, skipping this test.")
	}

	srcData, err := ioutil.ReadFile(path + srcPathTest)
	if err != nil {
		t.Error(err.Error())
	}
	srcConf := Config{}
	json.Unmarshal(srcData, &srcConf)
	fmt.Println(srcConf)
	fmt.Println(srcConf.Builders["azure1"])

	dstData, err := ioutil.ReadFile(path + dstPathTest)
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
		if dstConf.Builders == nil {
			dstConf.Builders = map[string]string{sk: sv}
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
		if dstConf.Provisioners == nil {
			dstConf.Provisioners = map[string]string{sk: sv}
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

		err = ioutil.WriteFile(path+dstPathModTest, dstDataMod, 0600)
		if err != nil {
			t.Error(err.Error())
		}
	}
}
