package response

import (
	"github.com/MSOpenTech/packer-azure/packer/builder/azure/driver_restapi/response/model"
	"io"
)

func ParseAvailabilityResponse(body io.ReadCloser) (*model.AvailabilityResponse, error ) {
	data, err := toModel(body, &model.AvailabilityResponse{})

	if err != nil {
		return nil, err
	}

	m := data.(*model.AvailabilityResponse)

	return m, nil
}

