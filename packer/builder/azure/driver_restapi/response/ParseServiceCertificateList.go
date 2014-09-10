package response

import (
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseServiceCertificateList(body io.ReadCloser) (*model.ServiceCertificateList, error ) {
	data, err := toModel(body, &model.ServiceCertificateList{})

	if err != nil {
		return nil, err
	}
	m := data.(*model.ServiceCertificateList)

	return m, nil
}
