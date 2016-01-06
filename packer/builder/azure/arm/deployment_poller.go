// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package arm

import (
	"time"
)

const (
	DeployCanceled  = "Canceled"
	DeployFailed    = "Failed"
	DeployDeleted   = "Deleted"
	DeploySucceeded = "Succeeded"
)

type DeploymentPoller struct {
	getProvisioningState func() (string, error)
	pause                func()
}

func NewDeploymentPoller(getProvisioningState func() (string, error)) *DeploymentPoller {
	pollDuration := time.Second * 15

	return &DeploymentPoller{
		getProvisioningState: getProvisioningState,
		pause:                func() { time.Sleep(pollDuration) },
	}
}

func (t *DeploymentPoller) PollAsNeeded() (string, error) {
	for {
		res, err := t.getProvisioningState()

		if err != nil {
			return res, err
		}

		switch res {
		case DeployCanceled, DeployDeleted, DeployFailed, DeploySucceeded:
			return res, nil
		default:
			break
		}

		t.pause()
	}
}
