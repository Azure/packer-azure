// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package powershell

import "github.com/mitchellh/packer/packer"

// A driver is able to talk to PowerShell Azure and perform certain
// operations with it. Some of the operations on here may seem overly
// specific, but they were built specifically in mind to handle features
// of the HyperV builder for Packer, and to abstract differences in
// versions out of the builder steps, so sometimes the methods are
// extremely specific.
type Driver interface {

	// Exec executes the given PowerShell command
	Exec(string) error

	// ExecRet executes the given PowerShell command and returns a value
	ExecRet(string) (string, error)

	ExecRemote(cmd *packer.RemoteCmd) error

	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error
}
