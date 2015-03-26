package retry

import (
	"fmt"
	"time"

	"github.com/MSOpenTech/azure-sdk-for-go/management"
)

// ExecuteAsyncOperation blocks until the provided asyncOperation is
// no longer in the InProgress state. Any known retryiable transient
// errors are retried and additional retry rules can be specified.
// If the operation was successful, nothing is returned, otherwise
// an error is returned.
func ExecuteAsyncOperation(client management.Client, asyncOperation func() (management.OperationId, error), extraRules ...RetryRule) error {
	if asyncOperation == nil {
		return fmt.Errorf("Parameter not specified: %s", "asyncOperation")
	}

	retryPolicy := append(newDefaultRetryPolicy(), extraRules...)

	for { // retry loop for azure errors, call continue for retryable errors

		operationId, err := asyncOperation()
		if err == nil {
			err = client.WaitAsyncOperation(operationId)
		}
		if err != nil {
			// need to remove the pointer receiver in Azure SDK to make these *'s go away
			if azureError, ok := err.(*management.AzureError); ok {
				if shouldRetry, backoff := retryPolicy.ShouldRetry(*azureError); shouldRetry {
					time.Sleep(backoff)
					continue // retry asyncOperation
				}
			}
		}
		return err
	}
}
