package retry

import (
	"log"
	"strings"
	"time"

	"github.com/MSOpenTech/azure-sdk-for-go/management"
)

func newDefaultRetryPolicy() retryPolicy {
	return retryPolicy{
		ExponentialBackoffRule("Throttling", func(err management.AzureError) bool {
			return err.Code == "TooManyRequests"
		}, 5*time.Second, 2*time.Minute, 0),
		ConstantBackoffRule("InternalError", func(err management.AzureError) bool {
			return err.Code == "InternalError"
		}, 10*time.Second, 100),
		ConstantBackoffRule("Conflict/InUse", func(err management.AzureError) bool {
			return (err.Code == "BadRequest" && strings.Contains(err.Message, "is currently in use by")) ||
				(err.Code == "ConflictError" && strings.Contains(err.Message, "that requires exclusive access"))
		}, 10*time.Second, 100),
	}
}

type RetryRule func(management.AzureError) (bool, time.Duration)
type retryPolicy []RetryRule
type matchRule func(management.AzureError) bool

func (rules retryPolicy) ShouldRetry(err management.AzureError) (bool, time.Duration) {
	for _, rule := range rules {
		if shouldRetry, backoff := rule(err); shouldRetry {
			return shouldRetry, backoff
		}
	}
	return false, 0
}

func ConstantBackoffRule(name string, match matchRule, backoff time.Duration, maxRetries int) RetryRule {
	indefinitely := maxRetries == 0
	retries := 0
	return func(err management.AzureError) (bool, time.Duration) {
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

func ExponentialBackoffRule(name string, match matchRule, initialBackoff time.Duration, maximumBackoff time.Duration, maxRetries int) RetryRule {
	indefinitely := maxRetries == 0
	retries := 0
	backoff := initialBackoff
	return func(err management.AzureError) (bool, time.Duration) {
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
