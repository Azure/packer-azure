packer-azure [![Build Status](https://travis-ci.org/Azure/packer-azure.svg)](https://travis-ci.org/Azure/packer-azure) [![Code Coverage](https://githubpackerazure.blob.core.windows.net/codecoverage/coverage.svg)](https://githubpackerazure.blob.core.windows.net/codecoverage/coverage.html)
=============

[Packer](http://www.packer.io/intro/index.html) is an open source tool for creating identical machine images for multiple platforms from a single source configuration.
This is an Azure plugin for Packer.io to enable Microsoft Azure users to build custom images given an Azure image.

You must have an Azure subscription to begin using [Azure](http://azure.microsoft.com).
You can build both Linux and Windows Azure images (targets) from both Windows and Linux dev-boxes (clients) with these plugins.

#### Packer version the plug-ins were tested is 0.7.5

### Usage
packer-azure utilizes **Service Management REST API** and **Storage Services REST API** and consists of two plug-ins: **packer-builder-azure** and **packer-provisioner-azure-custom-script-extension** (for Windows targets). For Linux targets use well known "shell" provisioner; More information about the custom script provisioner can be found at http://msdn.microsoft.com/en-us/library/dn781373.aspx
To start using the plugin you will need to get **PublishSetting profile** for your azure subscriptions. Visit  https://manage.windowsazure.com/publishsettings to download the publish profile for the currently logged in user.

You can download binaries from the [releases](https://github.com/Azure/packer-azure/releases) for this project and drop them in your [packer install](https://packer.io/docs/installation.html) directory or you can build the plugins from source (see below). Configuration examples can be found in the [config_examples](https://github.com/Azure/packer-azure/tree/master/config_examples) directory.

## Building from source
* [Install Go](https://golang.org/doc/install) 1.5.3 or newer. Go and packer-azure dependencies require [git](http://git-scm.com/) to be installed as well.
* [Install packer](https://packer.io/docs/installation.html)
* Set the environment variable GO15VENDOREXPERIMENT to 1 when using Go 1.5.
* To build the plugins set the GOBIN environment variable to your packer install directory and run `go get github.com/Azure/packer-azure/packer/plugin/...`

### Dependencies (as of 2016-03-15)

 * github.com/Azure/azure-sdk-for-go
 * github.com/Azure/go-autorest
 * github.com/Azure/go-ntlmssp
 * github.com/Azure/packer-azure
 * github.com/mitchellh/go-homedir
 * github.com/mitchellh/multistep
 * github.com/mitchellh/packer
 * github.com/pborman/uuid
 * golang.org/x/crypto

Please see the [ARM README](https://github.com/Azure/packer-azure/blob/master/packer/builder/azure/arm/README.md) for
getting started with packer-azure for [ARM](https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/).
