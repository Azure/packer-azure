package model

import (
	"encoding/xml"
	"regexp"
)

type VmImageList struct {
	XMLName   xml.Name `xml:"VMImages"`
	Xmlns	  	string `xml:"xmlns,attr"`
	VMImages []VMImage `xml:"VMImage"`
}

type VMImage struct {
	Name  						string
	Label  						string
	Category  					string
	Description  				string
	OSDiskConfiguration  		OSDiskConfiguration
	DataDiskConfigurations		[]DataDiskConfiguration	`xml:"DataDiskConfigurations > DataDiskConfiguration"`
	ServiceName					string
	DeploymentName				string
	RoleName					string
	Location					string
	AffinityGroup				string
	CreatedTime					string
	ModifiedTime				string
	Language					string
	ImageFamily					string
	RecommendedVMSize			string
	IsPremium					string
	Eula						string
	IconUri						string
	SmallIconUri				string
	PrivacyUri					string
	PublisherName				string
	PublishedDate				string
	ShowInGui					string
	PricingDetailLink			string
}

type OSDiskConfiguration struct {
	Name  					string
	HostCaching  			string
	OSState  				string
	OS  					string
	MediaLink  				string
	LogicalDiskSizeInGB  	string
}

type DataDiskConfiguration struct {
	Name  					string
	HostCaching  			string
	Lun  					string
	MediaLink  				string
	LogicalDiskSizeInGB  	string
}


func (l *VmImageList) First(name string) *VMImage {
	pattern := name
	for _, im := range(l.VMImages){
		matchName, _ := regexp.MatchString(pattern, im.Name)
		if( matchName ) {
			return &im
		}
	}

	return nil
}
