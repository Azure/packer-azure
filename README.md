packer-azure
=============

Packer is an open source tool for creating identical machine images for multiple platforms from a single source configuration. For an introduction to Packer, check out documentation at http://www.packer.io/intro/index.html.

This is an Azure plugin for Packer.io to enable Microsoft Azure users to build custom images given an Azure image. 

You must have an Azure subscription to begin using Azure. http://azure.microsoft.com

You can build Linux and Windows Azure images (targets) with this plugin. 

You can execute the plugin from both Windows and Linux dev-boxes (clients). 

**The bin directory contains binaries and example configurations for Packer-Azure.**

#### Packer version the plug-ins were tested is 0.7.1

### Windows dev-box

* packer-azure for Windows implemented as a **PowerShell Azure** wrapper and consists of two plug-ins: **builder-azure.exe** and **provisioner-powershell-azure.exe** (for Windows targets); 
* To build the builder use this command: **go install  -tags 'powershell' github.com\MSOpenTech\packer-azure\packer\plugin\builder-azure**;
* To build the provisioner (for Windows targets) use this command: **go install github.com\MSOpenTech\packer-azure\packer\plugin\provisioner-powershell-azure**;

### Linux dev-box

* packer-azure for Linux utilizes **Service Management REST API** and **Storage Services REST API** and consists of two plug-ins: **builder-azure** and **provisioner-azure-custom-script-extension** (for Windows targets); 
* To build the builder use this command: **go install -tags 'restapi' github.com\MSOpenTech\packer-azure\packer\plugin\builder-azure**;
* To build the provisioner (for Windows targets) use this command: **go install github.com\MSOpenTech\packer-azure\packer\plugin\provisioner-azure-custom-script-extension**. Visit http://msdn.microsoft.com/en-us/library/dn781373.aspx to understand how the provisioner works;
* To manage certificates packer-azure uses **openssl**;
* To start using the plugin you will need to get **PublishSetting profile**. Visit one of the links bellow to get the profile:
  * https://windows.azure.com/download/publishprofile.aspx
  * http://go.microsoft.com/fwlink/?LinkId=254432

* Easy steps to build the plugin on Ubunty
  * sudo apt-get install git
  * sudo apt-get install mercurial meld
  * go get github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi
  * go get github.com/hashicorp/yamux
  * go get github.com/ugorji/go/codec
  * go install -tags 'restapi' github.com/MSOpenTech/packer-azure/packer/plugin/builder-azure
  * go install github.com/MSOpenTech/packer-azure/packer/plugin/provisioner-azure-custom-script-extension

### Dependencies

*	code.google.com/p/go.crypto
*	github.com/mitchellh/go-fs
*	github.com/mitchellh/iochan
*	github.com/mitchellh/mapstructure
*	github.com/mitchellh/multistep
*	github.com/mitchellh/packer
*	github.com/hashicorp/go-version
*	github.com/hashicorp/yamux
*	github.com/ugorji/go/codec


