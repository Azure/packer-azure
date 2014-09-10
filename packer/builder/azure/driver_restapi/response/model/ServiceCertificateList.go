package model

import (
	"encoding/xml"
)

type ServiceCertificateList struct {
	XMLName   xml.Name `xml:"Certificates"`
	Certificates []Certificate `xml:"Certificate"`
}

type Certificate struct {
	CertificateUrl string
	Thumbprint string
	ThumbprintAlgorithm string
	Data string
}
