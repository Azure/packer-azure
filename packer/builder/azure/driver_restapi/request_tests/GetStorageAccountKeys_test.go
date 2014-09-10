package request_tests

import (
	"testing"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response"
	"fmt"
)

func _TestGetStorageAccountKeys(t *testing.T) {

	errMassage := "TestGetStorageAccountKeys: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	storageAccountName := "packervhds"
	requestData := reqManager.GetStorageAccountKeys(storageAccountName)
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	storageService, err := response.ParseStorageService(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	fmt.Printf("storageService: %v\n\n", storageService)

	t.Error("eom")
}
