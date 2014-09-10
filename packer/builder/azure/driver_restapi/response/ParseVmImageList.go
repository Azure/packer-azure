package response

import (
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseVmImageList(body io.ReadCloser) (*model.VmImageList, error ) {
	data, err := toModel(body, &model.VmImageList{})

	if err != nil {
		return nil, err
	}

	m := data.(*model.VmImageList)

	return m, nil
}

