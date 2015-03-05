packer-azure [![Build Status](https://travis-ci.org/MSOpenTech/packer-azure.svg)](https://travis-ci.org/MSOpenTech/packer-azure)
=============

Packer is an open source tool for creating identical machine images for multiple platforms from a single source configuration. For an introduction to Packer, check out documentation at http://www.packer.io/intro/index.html.

This is an Azure plugin for Packer.io to enable Microsoft Azure users to build custom images given an Azure image. 

You must have an Azure subscription to begin using Azure. http://azure.microsoft.com

You can build Linux and Windows Azure images (targets) with this plugin. 

You can execute the plugin from both Windows and Linux dev-boxes (clients). 

**The bin directory contains binaries and example configurations for Packer-Azure.**

#### Packer version the plug-ins were tested is 0.7.2

### Dependencies

*	code.google.com/p/go.crypto
*	code.google.com/p/go-uuid/uuid
*	github.com/mitchellh/go-fs
*	github.com/mitchellh/iochan
*	github.com/mitchellh/mapstructure
*	github.com/mitchellh/multistep
*	github.com/mitchellh/packer
*	github.com/hashicorp/go-version
*	github.com/hashicorp/yamux
*	github.com/hashicorp/go-msgpack/codec

### Windows dev-box

* packer-azure for Windows implemented as a **PowerShell Azure** wrapper and consists of two plug-ins: **packer-builder-azure.exe** and **packer-provisioner-powershell-azure.exe** (for Windows targets); 
* To build the builder use this command: **go install  -tags 'powershell' github.com\MSOpenTech\packer-azure\packer\plugin\packer-builder-azure**;
* To build the provisioner (for Windows targets) use this command: **go install github.com\MSOpenTech\packer-azure\packer\plugin\packer-provisioner-powershell-azure**;

### Linux dev-box

* packer-azure for Linux utilizes **Service Management REST API** and **Storage Services REST API** and consists of two plug-ins: **packer-builder-azure** and **packer-provisioner-azure-custom-script-extension** (for Windows targets). For Linux targets use well known "shell" provisioner; 
* To build the builder use this command: **go install -tags 'restapi' github.com\MSOpenTech\packer-azure\packer\plugin\packer-builder-azure**;
* To build the provisioner (for Windows targets) use this command: **go install github.com\MSOpenTech\packer-azure\packer\plugin\packer-provisioner-azure-custom-script-extension**.<br/><i>Visit http://msdn.microsoft.com/en-us/library/dn781373.aspx to understand how the provisioner works</i>;
* To manage certificates packer-azure uses **openssl**;
* To start using the plugin you will need to get **PublishSetting profile**. Visit one of the links bellow to get the profile:
  * http://go.microsoft.com/fwlink/?LinkId=254432
  * <del>https://windows.azure.com/download/publishprofile.aspx</del> (dead link)

* Easy steps to build the plugin on Ubunty
  * install go 1.3, visit https://golang.org/doc/install for details. Possible steps to install go 1.3:
  	* wget -P $HOME/downloads  https://storage.googleapis.com/golang/go1.3.1.linux-amd64.tar.gz
  	* sudo tar -C /usr/local -xzf $HOME/downloads/go1.3.1.linux-amd64.tar.gz
  	* mkdir $HOME/go
  	* export PATH=$PATH:/usr/local/go/bin
	* export GOROOT=/usr/local/go
	* export GOPATH=$HOME/go
	* export PATH=$PATH:$GOPATH/bin
  * sudo apt-get install git
  * sudo apt-get install mercurial meld
  * go get github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi
  * go get github.com/hashicorp/yamux
  * go get github.com/hashicorp/go-msgpack/codec
  * go install -tags 'restapi' github.com/MSOpenTech/packer-azure/packer/plugin/packer-builder-azure
  * go install github.com/MSOpenTech/packer-azure/packer/plugin/packer-provisioner-azure-custom-script-extension
  * copy built plugins from $GOPATH/bin to you Packer folder 
   
### Mac OSX dev-box
To build and install on a OS X dev machine you will need to install Go and the Mercurial packages, download dependencies, then build. 

* Install Go 1.3 from https://golang.org/dl/
* Install the Mercurial package (required to fetch dependencies) http://mercurial.selenic.com/downloads
* Open a Terminal session and run:
	* mkdir $HOME/go
  	* export PATH=$PATH:/usr/local/go/bin
	* export GOROOT=/usr/local/go
	* export GOPATH=$HOME/go
	* export PATH=$PATH:$GOPATH/bin
	* go get github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi
	* go get github.com/hashicorp/yamux
    * go get github.com/hashicorp/go-msgpack/codec
  	* go install -tags 'restapi' github.com/MSOpenTech/packer-azure/packer/plugin/packer-builder-azure
  	* go install github.com/MSOpenTech/packer-azure/packer/plugin/packer-provisioner-azure-custom-script-extension
 * copy built plugins from $GOPATH/bin to you Packer folder


### Quick Packer configuration examples:

To launch a Linux instance in Azure:

```
{
"builders":[{
 	"type":	"azure",
 	"publish_settings_path":	"your_path",
 	"subscription_name":	"your_name",
 	"storage_account":	"your_storage_account",
 	"storage_account_container":	"my_images",
 	"os_type":	"Linux",
 	"os_image_label":	"Ubuntu Server 14.04 LTS",
 	"location":	"West US",
 	"instance_size":	"Small",
 	"user_image_label":	"PackerMade_Ubuntu_Serv14"
}],
"provisioners":[{
 	"type":	"shell",
 	"execute_command":	"chmod +x {{ .Path }}; {{ .Vars }} sudo -E sh '{{ .Path }}'",
 	"inline": [	"sudo apt-get update",
				"sudo apt-get install -y mc",
				"sudo apt-get install -y nodejs",
				"sudo apt-get install -y npm",
				"sudo npm install azure-cli -g"
 	],
 	"inline_shebang":	"/bin/sh -x"
}] }
```

To launch a Windows instance in Azure:

```
{
"builders":[{
 	"type":	"azure",
 	"publish_settings_path":	"your_path",
 	"subscription_name":	"your_name",
 	"storage_account":	"your_storage_account",
 	"storage_account_container":	"my_images",
 	"os_type":	"Windows",
 	"os_image_label":	"Windows Server 2012 R2 Datacenter",
 	"location":	"West US",
 	"instance_size":	"Small",
 	"user_image_label":	"PackerMade_Windows2012R2DC"
}],
"provisioners":[{
 	"type":	"azure-custom-script-extension",
 	"inline": [	"Write-Host 'Inline script!'",
				"Write-Host 'Installing Mozilla Firefox...'",
				"$filename = 'Firefox Setup 31.0.exe'",
				"$link = 'https://download.mozilla.org/?product=firefox-31.0-SSL&os=win&lang=en-US'",
				"$dstDir = 'c:/MyFileFolder'",
				"New-Item $dstDir -type directory -force | Out-Null",
				"$remotePath = Join-Path $dstDir $filename",
				"(New-Object System.Net.Webclient).downloadfile($link, $remotePath)",
				"Start-Process $remotePath -NoNewWindow -Wait -Argument '/S'",
				"Write-Host 'Inline script finished!'"
 	]
}] }
```

### Quick steps to get Packer on Ubunty
  * wget -P $HOME/downloads https://dl.bintray.com/mitchellh/packer/packer_0.7.2_linux_amd64.zip
  * unzip $HOME/downloads/packer_0.7.2_linux_amd64.zip -d $HOME/packer/
  * export PATH=$PATH:$HOME/packer/
  * export PACKER_LOG=1
  * export PACKER_LOG_PATH=$HOME/packer.log

### Known Issues
  * It was discovered that some Linux distributions behave strangely as a target. In particular, if a user 
    1. creates a VM from an **OpenLogic image** using **certificate authentication only**;
    2. captures the VM to an user image;
    3. creates a VM from the captured image using **password and certificate authentication** - password authentication won't work.

    - Since the Packer plugin uses the same scenario (steps 1-2) to provision images - be ready to use Packer created images with certificate authentication only.
    - **All Ubuntu distributions work fine**. 	


