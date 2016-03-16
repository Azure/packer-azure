// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package common

import (
	"time"
)

type InterruptibleTaskResult struct {
	Err         error
	IsCancelled bool
}

type InterruptibleTask struct {
	IsCancelled func() bool
	Task        func() error
}

func NewInterruptibleTask(isCancelled func() bool, task func() error) *InterruptibleTask {
	return &InterruptibleTask{
		IsCancelled: isCancelled,
		Task:        task,
	}
}

func StartInterruptibleTask(isCancelled func() bool, task func() error) InterruptibleTaskResult {
	t := NewInterruptibleTask(isCancelled, task)
	return t.Run()
}

func (s *InterruptibleTask) Run() InterruptibleTaskResult {
	completeCh := make(chan error)
	defer close(completeCh)

	go func() {
		err := s.Task()
		completeCh <- err
	}()

	resultCh := make(chan InterruptibleTaskResult)
	defer close(resultCh)

	for {
		if s.IsCancelled() {
			return InterruptibleTaskResult{Err: nil, IsCancelled: true}
		}

		select {
		case err := <-completeCh:
			return InterruptibleTaskResult{Err: err, IsCancelled: false}
		case <-time.After(100 * time.Millisecond):
		}
	}
}
