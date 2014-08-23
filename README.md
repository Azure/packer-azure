packer-azure
=============

Packer is an open source tool for creating identical machine images for multiple platforms from a single source configuration. For an introduction to Packer, check out documentation at http://www.packer.io/intro/index.html.

This is an Azure plugin for Packer.io to enable Microsoft Azure users to build custom images given an Azure image. 

You must have an Azure subscription to begin using Azure. http://azure.microsoft.com

The Packer-Azure plugin is still in development and the current version can be launched from windows workstations only. You can build Linux and Windows Azure images with this plugin. 

The bin directory contains example configurations for Packer-Azure.

To build plug-in use this command: 
	go install  -tags 'powershell' github.com\MSOpenTech\packer-azure\packer\plugin\builder-azure
