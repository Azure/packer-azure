package model

import "encoding/xml"

type Operation struct {
	XMLName   			xml.Name 	`xml:"Operation"`
	Xmlns	  			string 		`xml:"xmlns,attr"`
	ID 					string
	Status 				string
	HttpStatusCode 		string
	Error 				Error 		`xml:"Error"`
}

type Error struct {
	Code 		string
	Message 	string
}
