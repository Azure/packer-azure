packer-azure
=============

Packer is an open source tool for creating identical machine images for multiple platforms from a single source configuration. For an introduction to Packer, check out documentation at http://www.packer.io/intro/index.html.

This is an Azure plugin for Packer.io to enable Microsoft Azure users to build custom images given an Azure image. 

You must have an Azure subscription to begin using Azure. http://azure.microsoft.com

You can build Linux and Windows Azure images with this plugin. 

You can execute the plugin from both Windows and Linux dev-boxes 

**The bin directory contains binaries and example configurations for Packer-Azure.**

### Windows dev-box

* packer-azure for Windows implemented as a **PowerShell Azure** wrapper and consists of two plug-ins: **builder-azure.exe** and **provisioner-powershell-azure.exe**  
* To build the builder use this command: **go install  -tags 'powershell' github.com\MSOpenTech\packer-azure\packer\plugin\builder-azure**
* To build the provisioner use this command: **go install github.com\MSOpenTech\packer-azure\packer\plugin\provisioner-powershell-azure**

### Linux dev-box

* packer-azure for Linux utilizes **Service Management REST API** and **Storage Services REST API** and consists of two plug-ins: **builder-azure** and **provisioner-azure-custom-script-extension**  
* To build the builder use this command: **go install -tags 'restapi' github.com\MSOpenTech\packer-azure\packer\plugin\builder-azure**
* To build the provisioner use this command: **go install github.com\MSOpenTech\packer-azure\packer\plugin\provisioner-azure-custom-script-extension** 
* To manage certificates packer-azure uses **openssl**
