package response

import (
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseResourceExtensionList(body io.ReadCloser) (*model.ResourceExtensionList, error ) {
	data, err := toModel(body, &model.ResourceExtensionList{})

	if err != nil {
		return nil, err
	}
	m := data.(*model.ResourceExtensionList)

	return m, nil
}
