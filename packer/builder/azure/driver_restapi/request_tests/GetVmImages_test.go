package request_tests

import (
	"testing"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response"
)

func _TestGetVmImages(t *testing.T) {

	errMassage := "TestCaptureVmImage: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	requestData := reqManager.GetVmImages()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	vmImageList, err := response.ParseVmImageList(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	t.Logf("vmImageList:\n\n")

//	for _, image := range(vmImageList.VMImages){
//		t.Logf("%v\n\n", image)
//	}

	userImageName := "PackerMade_Ubuntu_Server_14_04_2014-August-21_16-12"

	first := vmImageList.First(userImageName)
	if first != nil {
		t.Logf("%v\n\n", first)
	}

	t.Error("eom")
}
