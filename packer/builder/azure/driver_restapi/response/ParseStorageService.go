package response

import (
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseStorageService(body io.ReadCloser) (*model.StorageService, error ) {
	data, err := toModel(body, &model.StorageService{})

	if err != nil {
		return nil, err
	}

	m := data.(*model.StorageService)

	return m, nil
}
