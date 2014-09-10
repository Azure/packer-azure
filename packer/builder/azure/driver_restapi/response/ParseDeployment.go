package response

import (
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseDeployment(body io.ReadCloser) (*model.Deployment, error ) {
	data, err := toModel(body, &model.Deployment{})

	if err != nil {
		return nil, err
	}

	m := data.(*model.Deployment)

	return m, nil
}
