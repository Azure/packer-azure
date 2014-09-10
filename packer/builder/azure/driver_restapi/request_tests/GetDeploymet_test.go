package request_tests

import (
	"testing"
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response"
	"fmt"
)

func _TestGetDeploymet(t *testing.T) {

	errMassage := "TestGetDeploymet: %s\n"

	reqManager, err := getRequestManager()
	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	serviceName := "pkrsrvpakd9ma4yb"
	vmName := "shchVm1"

	requestData := reqManager.GetDeployment(serviceName, vmName)
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

/*
	defer resp.Body.Close()
	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)

	t.Logf("resp.Body: %s\n", string(respBody))
*/

	deployment, err := response.ParseDeployment(resp.Body)

	if err != nil {
		t.Errorf(errMassage, err.Error())
	}

	fmt.Printf("\ndeployment:\n\n %v", deployment.RoleInstanceList[0].GuestAgentStatus)
	fmt.Printf("\ndeployment:\n\n %v", deployment.RoleInstanceList[0].ResourceExtensionStatusList[1])

	t.Error("eom")
}
