package retry

import (
	"github.com/Azure/azure-sdk-for-go/management"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestAzureConflictInUseMatches(c *C) {
	rule := newRetryRuleConflictInUse()
	err := returnConflictErrorLikeAzureSDK()

	// type assertion from ExecuteAsyncOperation (operations.go)
	if azureError, ok := err.(*management.AzureError); ok {
		shouldRetry, _ := rule(*azureError)
		c.Check(shouldRetry, Equals, true)
	} else {
		c.Error("err is not AzureError")
	}
}

// returnConflictErrorLikeAzureSDK returns the AzureError exactly like the SDK client would.
// See https://github.com/Azure/azure-sdk-for-go/blob/47c70157399d900fab8e5863f0fb2851eab0cb15/management/operations.go#L73
func returnConflictErrorLikeAzureSDK() error {

	op := management.GetOperationStatusResponse{
		Status:         management.OperationStatusFailed,
		HTTPStatusCode: "409",
		Error: &management.AzureError{
			Code:    "ConflictError",
			Message: "Windows Azure is currently performing an operation on deployment 'xxx' that requires exclusive access. Please try again later for operation on disk or image 'xxx' associated with the deployment.",
		},
	}

	return op.Error
}
