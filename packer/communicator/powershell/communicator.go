// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package powershell

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"os"
	"path/filepath"
)

type comm struct {
	config *Config
}

type Config struct {
	Driver        Driver
	Username      string
	Password      string
	RemoteHostUrl string
	Ui            packer.Ui
}

// Creates a new packer.Communicator implementation over SSH. This takes
// an already existing TCP connection and SSH configuration.
func New(config *Config) (result *comm, err error) {
	// Establish an initial connection and connect
	result = &comm{
		config: config,
	}

	return
}

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
	username := c.config.Username
	password := c.config.Password
	remoteHostUrl := c.config.RemoteHostUrl
	driver := c.config.Driver

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock { ")
	blockBuffer.WriteString("$uri = '" + remoteHostUrl + "';")
	blockBuffer.WriteString("$username = '" + username + "';")
	blockBuffer.WriteString("$password = '" + password + "';")
	blockBuffer.WriteString("$secPassword = ConvertTo-SecureString $password -AsPlainText -Force;")
	blockBuffer.WriteString("$credential = New-Object -typename System.Management.Automation.PSCredential -argumentlist $username, $secPassword;")
	blockBuffer.WriteString("$opt = New-PSSessionOption -OperationTimeout 540000;")
	blockBuffer.WriteString("$sess = New-PSSession -ConnectionUri $uri -Credential $credential -SessionOption $opt;")
	blockBuffer.WriteString("Invoke-Command -Session $sess ")
	blockBuffer.WriteString(cmd.Command)
	blockBuffer.WriteString("; Remove-PSSession -session $sess;")
	blockBuffer.WriteString("}")

	var cmdCopy packer.RemoteCmd
	cmdCopy.Stdout = cmd.Stdout
	cmdCopy.Stderr = cmd.Stderr

	cmdCopy.Command = blockBuffer.String()
	err = driver.ExecRemote(&cmdCopy)

	return
}

func (c *comm) Upload(string, io.Reader, *os.FileInfo) error {
	panic("not implemented for powershell")
}

func (c *comm) UploadDir(dst string, src string, excl []string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	ui := c.config.Ui

	if info.IsDir() {
		ui.Say(fmt.Sprintf("Uploading folder to the VM '%s' => '%s'...", src, dst))
		err := c.uploadFolder(dst, src)
		if err != nil {
			return err
		}
	} else {
		target_file := filepath.Join(dst, filepath.Base(src))
		ui.Say(fmt.Sprintf("Uploading file to the VM '%s' => '%s'...", src, target_file))
		err := c.uploadFile(target_file, src)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *comm) Download(string, io.Writer) error {
	panic("not implemented yet")
}

// region private helpers

func (c *comm) uploadFile(dscPath string, srcPath string) error {
	driver := c.config.Driver

	dscPath = filepath.FromSlash(dscPath)
	srcPath = filepath.FromSlash(srcPath)

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("$srcPath = '" + srcPath + "';")
	blockBuffer.WriteString("$dstDir = '" + dscPath + "';")
	blockBuffer.WriteString("$uri = '" + c.config.RemoteHostUrl + "';")
	blockBuffer.WriteString("$username = '" + c.config.Username + "';")
	blockBuffer.WriteString("$password = '" + c.config.Password + "';")
	blockBuffer.WriteString("$secPassword = ConvertTo-SecureString $password -AsPlainText -Force;")
	blockBuffer.WriteString("$credential = New-Object -typename System.Management.Automation.PSCredential -argumentlist $username, $secPassword;")
	blockBuffer.WriteString("$sess = New-PSSession -ConnectionUri $uri -Credential $credential;")
	blockBuffer.WriteString("function CopyFileToAzureVm($srcPath, $dstDir, $sess) { try {  $filename = gci $srcPath ;  $remotePath = Join-Path $dstDir $filename.Name;  $localPath = $srcPath;  [IO.FileStream]$filestream = [IO.File]::OpenRead( $localPath );  Invoke-Command -Session $sess -ScriptBlock { Param($remFile) New-Item $remFile -type file -force | Out-Null;[IO.FileStream]$filestream = [IO.File]::OpenWrite( $remFile ) } -ArgumentList $remotePath;  $content = [Io.File]::ReadAllBytes( $localPath );  $contentsizeMB = $content.Count / 1MB + 1MB;  $chunksize = 1MB;  [byte[]]$contentchunk = New-Object byte[] $chunksize;  while (($bytesread = $filestream.Read( $contentchunk, 0, $chunksize )) -ne 0) {   Invoke-Command -Session $sess -ScriptBlock {    Param($data, $bytes);    $filestream.Write( $data, 0, $bytes );   } -ArgumentList $contentchunk, $bytesread;  }  Invoke-Command -Session $sess -ScriptBlock {$filestream.Close()};  $filestream.Close(); } catch {  if($filestream -ne $null){$filestream.Close()};  Invoke-Command -Session $sess -ScriptBlock {if($filestream -ne $null){$filestream.Close()}};  Write-Error $_.Exception.Message; }};")
	blockBuffer.WriteString("CopyFileToAzureVm -srcPath $srcPath -dstDir $dstDir -sess $sess;")
	blockBuffer.WriteString("Remove-PSSession -session $sess;")

	err := driver.Exec(blockBuffer.String())

	return err
}

func (c *comm) uploadFolder(dscPath string, srcPath string) error {

	driver := c.config.Driver

	dscPath = filepath.FromSlash(dscPath)
	srcPath = filepath.FromSlash(srcPath)

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("$srcDir = '" + srcPath + "';")
	blockBuffer.WriteString("$dstDir = '" + dscPath + "';")
	blockBuffer.WriteString("$uri = '" + c.config.RemoteHostUrl + "';")
	blockBuffer.WriteString("$username = '" + c.config.Username + "';")
	blockBuffer.WriteString("$password = '" + c.config.Password + "';")
	blockBuffer.WriteString("$secPassword = ConvertTo-SecureString $password -AsPlainText -Force;")
	blockBuffer.WriteString("$credential = New-Object -typename System.Management.Automation.PSCredential -argumentlist $username, $secPassword;")
	blockBuffer.WriteString("$sess = New-PSSession -ConnectionUri $uri -Credential $credential;")
	blockBuffer.WriteString("function CopyDirToAzureVm($srcDir, $dstDir, $sess) { try {  Set-Location $srcDir;  Get-ChildItem -Path $srcDir -recurse | ? {($_.psiscontainer -ne $true)} | foreach {   $relPath = $_ | Resolve-Path -Relative;   $remotePath = Join-Path $dstDir $relPath;   $localPath = $_.FullName;   [IO.FileStream]$filestream = [IO.File]::OpenRead( $localPath );   Invoke-Command -Session $sess -ScriptBlock { Param($remFile) New-Item $remFile -type file -force | Out-Null;[IO.FileStream]$filestream = [IO.File]::OpenWrite( $remFile ) } -ArgumentList $remotePath;   $content = [Io.File]::ReadAllBytes( $localPath );   $contentsizeMB = $content.Count / 1MB + 1MB;   $chunksize = 1MB;   [byte[]]$contentchunk = New-Object byte[] $chunksize;   while (($bytesread = $filestream.Read( $contentchunk, 0, $chunksize )) -ne 0) {    Invoke-Command -Session $sess -ScriptBlock {     Param($data, $bytes);     $filestream.Write( $data, 0, $bytes );    } -ArgumentList $contentchunk, $bytesread;   }   Invoke-Command -Session $sess -ScriptBlock {$filestream.Close()};   $filestream.Close();  } } catch {  if($filestream -ne $null){$filestream.Close()};  Invoke-Command -Session $sess -ScriptBlock {if($filestream -ne $null){$filestream.Close()}};  Write-Error $_.Exception.Message; }};")
	blockBuffer.WriteString("CopyDirToAzureVm -srcDir $srcDir -dstDir $dstDir -sess $sess;")
	blockBuffer.WriteString("Remove-PSSession -session $sess;")

	err := driver.Exec(blockBuffer.String())

	return err
}
