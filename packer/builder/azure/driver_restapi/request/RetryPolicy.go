// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"log"
	"strings"
	"time"

	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
)

func defaultRetryPolicy() retryPolicy {
	return retryPolicy{
		exponentialBackoff("Throttling", func(err model.AzureError) bool {
			return err.Code == "TooManyRequests"
		}, 5*time.Second, 2*time.Minute, 0),
		constantBackoff("InternalError", func(err model.AzureError) bool {
			return err.Code == "InternalError"
		}, 10*time.Second, 100),
		constantBackoff("Conflict/InUse", func(err model.AzureError) bool {
			return (err.Code == "BadRequest" && strings.Contains(err.Message, "is currently in use by")) ||
				(err.Code == "ConflictError" && strings.Contains(err.Message, "that requires exclusive access"))
		}, 10*time.Second, 100),
	}
}

type retryRule func(model.AzureError) (bool, time.Duration)
type retryPolicy []retryRule
type matchRule func(model.AzureError) bool

func (rules retryPolicy) ShouldRetry(err model.AzureError) (shouldRetry bool, backoff time.Duration) {
	for _, rule := range rules {
		shouldRetry, backoff = rule(err)
		if shouldRetry {
			return
		}
	}
	return false, 0
}

func constantBackoff(name string, match matchRule, backoff time.Duration, maxRetries int) retryRule {
	indefinitely := maxRetries == 0
	retries := 0
	return func(err model.AzureError) (bool, time.Duration) {
		if match(err) {
			if indefinitely || retries < maxRetries {
				retries++
				log.Printf("Retry %d for rule '%s' with %v backoff", retries, name, backoff)
				return true, backoff
			} else {
				log.Printf("Retries for rule '%s' exhausted (%d)", name, retries)
			}
		}
		return false, 0
	}
}

func exponentialBackoff(name string, match matchRule, initialBackoff time.Duration, maximumBackoff time.Duration, maxRetries int) retryRule {
	indefinitely := maxRetries == 0
	retries := 0
	backoff := initialBackoff
	return func(err model.AzureError) (bool, time.Duration) {
		if match(err) {
			if indefinitely || retries < maxRetries {
				retries++
				thisBackoff := backoff
				backoff *= 2
				if backoff > maximumBackoff {
					backoff = maximumBackoff
				}
				log.Printf("Retry %d for rule '%s' with %v backoff", retries, name, thisBackoff)
				return true, thisBackoff
			} else {
				log.Printf("Retries for rule '%s' exhausted (%d)", name, retries)
			}
		}
		return false, 0
	}
}
