// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package arm

import (
	"bytes"
	"crypto/rsa"
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/ssh"
	"encoding/pem"
	"crypto/x509"
	"fmt"
	"time"
)

const (
	KeySize = 2048
)

type OpenSshKeyPair struct {
	privateKey *rsa.PrivateKey
	publicKey ssh.PublicKey
}

func NewOpenSshKeyPair() (*OpenSshKeyPair, error) {
	return NewOpenSshKeyPairWithSize(KeySize)
}

func NewOpenSshKeyPairWithSize(keySize int) (*OpenSshKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, err
	}

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return &OpenSshKeyPair {
		privateKey: privateKey,
		publicKey: publicKey,
	}, nil
}


func (s *OpenSshKeyPair) AuthorizedKey() string {
	b := bytes.NewBufferString("")
	b.WriteString(s.publicKey.Type())
	b.WriteString(" ")
	b.WriteString(base64.StdEncoding.EncodeToString(s.publicKey.Marshal()))
	b.WriteString(" ")
	b.WriteString(fmt.Sprintf("packer Azure Deployment@%s", time.Now().Format(time.RFC3339)))
	return b.String()
}

func (s *OpenSshKeyPair) PrivateKey() string {
	privateKey := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(s.privateKey),
	}))

	return privateKey
}

