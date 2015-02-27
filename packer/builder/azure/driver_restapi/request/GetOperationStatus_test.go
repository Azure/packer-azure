// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MSOpenTech/packer-azure/mod/pkg/net/http"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
)

func TestWaitForOperation_returns_when_succeeded(t *testing.T) {
	logs := new(bytes.Buffer)
	log.SetOutput(logs)
	defaultInterval = 1 * time.Microsecond
	responseBody := "<Operation><Status>InProgress</Status></Operation>"
	m := Manager{httpClient: httpClientStub{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			var rv = http.Response{
				Body: newStringReader(responseBody),
			}
			t.Logf("Exec returning %v", responseBody)
			return &rv, nil
		},
	}}
	b := newBarrier(1)

	var op *model.Operation
	var err error
	go func() {
		op, err = m.WaitForOperation("bla")
		t.Logf("WaitForOperation returned with %v,%v", op, err)
		b.RemoveParticipant()
		t.Logf("Barrier removed")
	}()

	if b.WaitFor(0 * time.Second) {
		t.Fatalf("Expected WaitForOperation to not have returned yet, but it has?")
	}

	responseBody = "<Operation><Status>Succeeded</Status><ID>test</ID></Operation>"

	if !b.WaitFor(defaultInterval + 15*time.Millisecond) {
		t.Fatalf("Expected WaitForOperation to have returned, but it hasn't?")
	}

	if err != nil {
		t.Fatalf("unexpected err: %T:%v", err, err)
	}
	if op == nil || op.ID != "test" {
		t.Fatalf("unexpected op: %v", op)
	}
}

func TestWaitForOperation_returns_when_failed(t *testing.T) {
	logs := new(bytes.Buffer)
	log.SetOutput(logs)
	defaultInterval = 1 * time.Microsecond
	responseBody := "<Operation><Status>InProgress</Status></Operation>"
	m := Manager{httpClient: httpClientStub{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			var rv = http.Response{
				Body: newStringReader(responseBody),
			}
			t.Logf("Exec returning %v", responseBody)
			return &rv, nil
		},
	}}
	b := newBarrier(1)

	var op *model.Operation
	var err error
	go func() {
		op, err = m.WaitForOperation("bla")
		t.Logf("WaitForOperation returned with %v,%v", op, err)
		b.RemoveParticipant()
		t.Logf("Barrier removed")
	}()

	if b.WaitFor(0 * time.Second) {
		t.Fatalf("Expected WaitForOperation to not have returned yet, but it has?")
	}

	responseBody = "<Operation><Status>Failed</Status><ID>test</ID><Error><Code>E1234</Code><Message>some error message</Message></Error></Operation>"

	if !b.WaitFor(defaultInterval + 15*time.Millisecond) {
		t.Fatalf("Expected WaitForOperation to have returned, but it hasn't?")
	}

	if err == nil || fmt.Sprintf("%T", err) != "model.AzureError" || err.Error() != "Azure error (E1234): some error message" {
		t.Fatalf("unexpected err: %T:%v", err, err)
	}
	if op == nil || op.ID != "test" {
		t.Fatalf("unexpected op: %v", op)
	}
}

type httpClientStub struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (s httpClientStub) Do(req *http.Request) (resp *http.Response, err error) {
	if s.DoFunc != nil {
		return s.DoFunc(req)
	}
	fmt.Printf("huh?")
	panic("stub not set up")
}

type barrier struct {
	wg sync.WaitGroup
}

func newBarrier(numParticipants int) *barrier {
	b := new(barrier)
	b.wg.Add(numParticipants)
	return b
}

func (b *barrier) RemoveParticipant() {
	b.wg.Done()
}

func (b *barrier) WaitFor(timeout time.Duration) bool {
	c := make(chan bool)
	go func() {
		b.wg.Wait()
		c <- true
	}()
	t := time.After(timeout)
	select {
	case _ = <-c:
		return true
	case _ = <-t:
		return false
	}
}

type stringReader struct {
	reader *strings.Reader
}

func newStringReader(s string) *stringReader {
	sr := stringReader{reader: strings.NewReader(s)}
	return &sr
}

func (stringReader) Close() error {
	return nil
}

func (sr stringReader) Read(p []byte) (n int, err error) {
	return sr.reader.Read(p)
}
