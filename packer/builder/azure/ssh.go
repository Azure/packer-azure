// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package azure

import (
	gossh "code.google.com/p/gosshold/ssh"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
)

// SSHAddress returns a function that can be given to the SSH communicator
func SSHAddress(state multistep.StateBag) (string, error) {
	port := state.Get("port").(string)
	tmpServiceName := state.Get("tmpServiceName").(string)
	return fmt.Sprintf("%s.cloudapp.net:%s", tmpServiceName, port), nil
}

// SSHConfig returns a function that can be used for the SSH communicator
//func SSHConfig(username string, password string) func(multistep.StateBag) (*gossh.ClientConfig, error) {
//	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
//
//		privateKey := state.Get("privateKey").(string)
//
//		keyring := new(ssh.SimpleKeychain)
//		if err := keyring.AddPEMKey(privateKey); err != nil {
//			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
//		}
//
//		auth := []gossh.ClientAuth{
//			gossh.ClientAuthPassword(ssh.Password(password)),
//			gossh.ClientAuthKeyboardInteractive(ssh.PasswordKeyboardInteractive(password)),
//			gossh.ClientAuthKeyring(keyring),
//		}
//
//		return &gossh.ClientConfig{
//			User: username,
//			Auth: auth,
//		}, nil
//	}
//}

func SSHConfig(username string) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		privateKey := state.Get("privateKey").(string)

		keyring := new(ssh.SimpleKeychain)
		if err := keyring.AddPEMKey(privateKey); err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		return &gossh.ClientConfig{
			User: username,
			Auth: []gossh.ClientAuth{
				gossh.ClientAuthKeyring(keyring),
			},
		}, nil
	}
}

