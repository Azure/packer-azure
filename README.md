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

* packer-azure for Linux utilizes **Service Management REST API** and **Storage Services REST API** and consists of two plug-ins: **builder-azure** and **provisioner-azure-custom-script-extension** (for Windows targets). For Linux targets use well known "shell" provisioner; 
* To build the builder use this command: **go install -tags 'restapi' github.com\MSOpenTech\packer-azure\packer\plugin\builder-azure**;
* To build the provisioner (for Windows targets) use this command: **go install github.com\MSOpenTech\packer-azure\packer\plugin\provisioner-azure-custom-script-extension**.<br/><i>Visit http://msdn.microsoft.com/en-us/library/dn781373.aspx to understand how the provisioner works</i>;
* To manage certificates packer-azure uses **openssl**;
* To start using the plugin you will need to get **PublishSetting profile**. Visit one of the links bellow to get the profile:
  * https://windows.azure.com/download/publishprofile.aspx
  * http://go.microsoft.com/fwlink/?LinkId=254432

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
  * go get github.com/ugorji/go/codec
  * go install -tags 'restapi' github.com/MSOpenTech/packer-azure/packer/plugin/builder-azure
  * go install github.com/MSOpenTech/packer-azure/packer/plugin/provisioner-azure-custom-script-extension
  * copy built plugins from $GOPATH/bin to you Packer folder
  * install the plugins, find details here: https://github.com/MSOpenTech/packer-azure/tree/master/bin/driver_restapi/lin
   
* Quick Packer configuration examples:
 <table border="1" style="width:100%;font-size:medium;">
     <tr>
		<th>Linux target</th> 
		<th>Windows target</th>
     </tr>
     <tr>
		<td valign="top"  style="font-size:medium;" >
			{<br>
				<table align="left" border="0" style="font-size:medium;">
				  <tr>
					<td colspan=3>"builders":[{</td>
				  </tr>
				  <tr>
					<td>&nbsp;</td>
					<td>"type":</td> 
					<td>"azure",</td>
				  </tr>
				  <tr>
					<td>&nbsp;</td>
					<td>"publish_settings_path":</td> 
					<td>"your_path",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"subscription_name":</td> 
					<td>"your_name",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"storage_account":</td> 
					<td>"your_storage_account",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"storage_account_container":</td> 
					<td>"my_images",</td>
				  </tr> 
			  
				  <tr>
					<td>&nbsp;</td>
					<td>"os_type":</td> 
					<td>"Linux",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"os_image_label":</td> 
					<td>"Ubuntu Server 14.04 LTS",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"location":</td> 
					<td>"West US",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"instance_size":</td> 
					<td>"Small",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"user_image_label":</td> 
					<td>"PackerMade_Ubuntu_Serv14"</td>
				  </tr> 
				  <tr>
					<td colspan=3>}],</td>
				  </tr>
				  <tr>
					<td colspan=3>"provisioners":[{</td>
				  </tr>
				  <tr>
					<td>&nbsp;</td>
					<td>"type":</td> 
					<td>"shell",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"execute_command":</td> 
					<td>"chmod +x {{ .Path }}; {{ .Vars }} sudo -E sh '{{ .Path }}'",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td valign="top">"inline": [</td> 
					<td>
						"sudo apt-get update",<br>
						"sudo apt-get install -y mc",<br>
						"sudo apt-get install -y nodejs",<br>
						"sudo apt-get install -y npm",<br>
						"sudo npm install azure-cli -g"
					</td>
				  </tr> 			  <tr>
					<td>&nbsp;</td>
					<td colspan=2>],</td> 
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"inline_shebang":</td> 
					<td>"/bin/sh -x"</td>
				  </tr> 

				  <tr>
					<td colspan=3>}]</td>
				  </tr>
				</table> 
			}
		</td>
		<td>
			{<br>
				<table align="left" border="0" style="font-size:medium;">
				  <tr>
					<td colspan=3>"builders":[{</td>
				  </tr>
				  <tr>
					<td>&nbsp;</td>
					<td>"type":</td> 
					<td>"azure",</td>
				  </tr>
				  <tr>
					<td>&nbsp;</td>
					<td>"publish_settings_path":</td> 
					<td>"your_path",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"subscription_name":</td> 
					<td>"your_name",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"storage_account":</td> 
					<td>"your_storage_account",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"storage_account_container":</td> 
					<td>"my_images",</td>
				  </tr> 
			  
				  <tr>
					<td>&nbsp;</td>
					<td>"os_type":</td> 
					<td>"Windows",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"os_image_label":</td> 
					<td>"Windows Server 2012 R2 Datacenter",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"location":</td> 
					<td>"West US",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"instance_size":</td> 
					<td>"Small",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td>"user_image_label":</td> 
					<td>"PackerMade_Windows2012R2DC"</td>
				  </tr> 
				  <tr>
					<td colspan=3>}],</td>
				  </tr>
				  <tr>
					<td colspan=3>"provisioners":[{</td>
				  </tr>
				  <tr>
					<td>&nbsp;</td>
					<td>"type":</td> 
					<td>"azure-custom-script-extension",</td>
				  </tr> 
				  <tr>
					<td>&nbsp;</td>
					<td valign="top">"inline": [</td> 
					<td>
						"Write-Host 'Inline script!'",<br>
						"Write-Host 'Installing Mozilla Firefox...'",<br>
						"$filename = 'Firefox Setup 31.0.exe'",<br>
						"$link = 'https://download.mozilla.org/?product=firefox-31.0-SSL&os=win&lang=en-US'",<br>
						"$dstDir = 'c:/MyFileFolder'",<br>
						"New-Item $dstDir -type directory -force | Out-Null",<br>
						"$remotePath = Join-Path $dstDir $filename",<br>
						"(New-Object System.Net.Webclient).downloadfile($link, $remotePath)",<br>
						"Start-Process $remotePath -NoNewWindow -Wait -Argument '/S'",<br>
						"Write-Host 'Inline script finished!'"
					</td>
				  </tr> 			  
				  <tr>
					<td>&nbsp;</td>
					<td colspan=2>]</td> 
				  </tr> 
				  <tr>
					<td colspan=3>}]</td>
				  </tr>
				</table> 
			}
		</td>
		
	</tr>
</table>

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


