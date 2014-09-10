package model

import "encoding/xml"

type AvailabilityResponse struct {
	XMLName   			xml.Name 	`xml:"AvailabilityResponse"`
	Xmlns	  			string 		`xml:"xmlns,attr"`
	Result 				string
	Reason 				string
}
