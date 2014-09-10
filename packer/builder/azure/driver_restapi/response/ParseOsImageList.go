package response

import (
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseOsImageList(body io.ReadCloser) (*model.OsImageList, error ) {
	data, err := toModel(body, &model.OsImageList{})

	if err != nil {
		return nil, err
	}

	m := data.(*model.OsImageList)

	return m, nil
}

