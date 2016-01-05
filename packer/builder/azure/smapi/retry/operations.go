package retry

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/management"
)

// ExecuteAsyncOperation blocks until the provided asyncOperation is
// no longer in the InProgress state. Any known retryiable transient
// errors are retried and additional retry rules can be specified.
// If the operation was successful, nothing is returned, otherwise
// an error is returned.
func ExecuteAsyncOperation(client management.Client, asyncOperation func() (management.OperationID, error), extraRules ...RetryRule) error {
	if asyncOperation == nil {
		return fmt.Errorf("Parameter not specified: %s", "asyncOperation")
	}

	retryPolicy := append(newDefaultRetryPolicy(), extraRules...)

	for { // retry loop for azure errors, call continue for retryable errors

		operationId, err := asyncOperation()
		if err == nil && operationId != "" {
			log.Printf("Waiting for operation: %s", operationId)
			err = client.WaitForOperation(operationId, nil)
		}
		if err != nil {
			log.Printf("Caught error (%T) during retryable operation: %v", err, err)
			// need to remove the pointer receiver in Azure SDK to make these *'s go away
			if azureError, ok := err.(management.AzureError); ok {
				log.Printf("Error is Azure error, checking if we should retry...")
				if shouldRetry, backoff := retryPolicy.ShouldRetry(azureError); shouldRetry {
					log.Printf("Error needs to be retried, sleeping %v", backoff)
					time.Sleep(backoff)
					continue // retry asyncOperation
				}
			}
		}
		return err
	}
}

// ExecuteOperation calls the provided syncOperation.
// Any known retryiable transient errors are retried and
// additional retry rules can be specified.
// If the operation was successful, nothing is returned, otherwise
// an error is returned.
func ExecuteOperation(syncOperation func() error, extraRules ...RetryRule) error {
	return ExecuteAsyncOperation(nil, func() (management.OperationID, error) { return "", syncOperation() }, extraRules...)
}
