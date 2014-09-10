// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model

import (
	"encoding/xml"
	"regexp"
	"sort"
	"strings"
)

type OsImageList struct {
	XMLName   xml.Name `xml:"Images"`
	Xmlns	  	string `xml:"xmlns,attr"`
	OSImages []OSImage `xml:"OSImage"`
}

type OSImage struct {
	AffinityGroup  		string
	Category			string
	Label				string
	Location			string
	LogicalSizeInGB		string
	MediaLink			string
	Name				string
	OS					string
	Eula				string
	Description			string
	ImageFamily			string
	ShowInGui			string
	PublishedDate		string
	IsPremium			string
	PrivacyUri			string
	RecommendedVMSize	string
	PublisherName		string
	PricingDetailLink	string
	SmallIconUri		string
	Language			string
}

func (l *OsImageList) Filter(label, location string) []OSImage {
	origLen := len(l.OSImages)
	filtered  := make([]OSImage, 0, origLen)

	pattern := label
	for _, im := range(l.OSImages){
		matchImageLocation := false
		for _, loc := range strings.Split(im.Location, ";") { if loc == location{ matchImageLocation = true; break } }
		if !matchImageLocation { continue }
		matchImageLabel, _ := regexp.MatchString(pattern, im.Label)
		matchImageFamily, _ := regexp.MatchString(pattern, im.ImageFamily)
		if( (matchImageLabel || matchImageFamily) && matchImageLabel ) {
			filtered = append( filtered, im)

		}
	}

	return filtered[:len(filtered)]
}

type ByDateDesc []OSImage

func (a ByDateDesc) Len() int           { return len(a) }
func (a ByDateDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDateDesc) Less(i, j int) bool { return a[i].PublishedDate > a[j].PublishedDate }

func (l *OsImageList) SortByDateDesc(imageList []OSImage) {
	if len(imageList) == 0 {
		sort.Sort(ByDateDesc(l.OSImages))
	} else {
		sort.Sort(ByDateDesc(imageList))
	}
}
