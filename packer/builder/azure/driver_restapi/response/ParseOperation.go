package response

import (
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseOperation(body io.ReadCloser) (*model.Operation, error ) {
	data, err := toModel(body, &model.Operation{})

	if err != nil {
		return nil, err
	}

	m := data.(*model.Operation)

	return m, nil
}

