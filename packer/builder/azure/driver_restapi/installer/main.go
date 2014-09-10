// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package main

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"encoding/json"
	"path/filepath"
	"os/user"
)

const (
	srcPath string = "config-template"
	dstPath string = ".packerconfig"
)

type Config struct {
	Builders   map[string]string			`json:"builders"`
	Provisioners   map[string]string		`json:"provisioners"`
}

func main () {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("Error getting working directory: " + err.Error())
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting home directory: " + err.Error())
	}

	usrHome := usr.HomeDir

	err = ModifyCoreConfig(filepath.Join(wd, srcPath), filepath.Join(usrHome, dstPath))
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
}

func ModifyCoreConfig(srcPath, dstPath string) error {
	var err error

	if _, err = os.Stat(srcPath); err != nil {
		err = fmt.Errorf("Config template path: '%v' check the path is correct.", srcPath)
		return err
	}

	if _, err = os.Stat(dstPath); err != nil { // no packer core config
		err = CopyFile(srcPath, dstPath)
		if err != nil {
			return fmt.Errorf("Can't copy config template: %s", err.Error())
		}
		fmt.Printf("Installer created a new Packer core config file '%s' in your home directory\n", dstPath)
		return nil
	}

	// create a copy of a packer core config
	dstPathCopy := filepath.Join(filepath.Dir(dstPath), filepath.Base(dstPath) + ".orig")
	err = CopyFile(dstPath, dstPathCopy)
	if err != nil {
		return fmt.Errorf("Can't create a backup copy of file: %s", err.Error())
	}

	srcData, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return err
	}
	srcConf := Config{}
	json.Unmarshal(srcData, &srcConf)

	dstData, err := ioutil.ReadFile(dstPath)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile returned a error: %s", err.Error())

	}

	dstConf := Config{}
	json.Unmarshal(dstData, &dstConf)

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

	if mod {
		dstDataMod, err := json.Marshal(dstConf)
		if err != nil {
			return fmt.Errorf("json.Marshal returned a error: %s", err.Error())
		}

		err = ioutil.WriteFile(dstPath, dstDataMod, 0600)
		if err != nil {
			return fmt.Errorf("ioutil.WriteFile returned a error: %s", err.Error())
		}
		fmt.Printf("Installer modified Packer core config file '%s' in your home directory\n", dstPath)
	} else {
		fmt.Printf("Installer found Packer core config file '%s' in your home directory already contains current plugin\n", dstPath)
	}

	return nil
}

func CopyFile(srcPath, dstPath string) error{
	in, err := os.Open(srcPath)
	if err != nil { return err }
	defer in.Close()
	out, err := os.Create(dstPath)
	if err != nil { return err }
	defer out.Close()
	_, err = io.Copy(out, in)
	err = out.Close()
	if err != nil { return err }
	return nil
}
