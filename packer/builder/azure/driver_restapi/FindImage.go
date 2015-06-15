// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package driver_restapi

import (
	osi "github.com/Azure/azure-sdk-for-go/management/osimage"
	vmi "github.com/Azure/azure-sdk-for-go/management/virtualmachineimage"
	"sort"
	"strings"
)

func FindVmImage(imageList []vmi.VMImage, name, label, location string) (vmi.VMImage, bool) {
	matches := make([]vmi.VMImage, 0)
	for _, im := range imageList {
		for _, loc := range strings.Split(im.Location, ";") {
			if loc == location &&
				(label != "" && im.Label == label) &&
				(name != "" && im.Name == name) {
				matches = append(matches, im)
			}
		}
	}

	if len(matches) > 0 {
		sort.Sort(vmImageByPublishDate(matches))
		return matches[0], true
	}
	return vmi.VMImage{}, false
}

type vmImageByPublishDate []vmi.VMImage

func (a vmImageByPublishDate) Len() int           { return len(a) }
func (a vmImageByPublishDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a vmImageByPublishDate) Less(i, j int) bool { return a[i].PublishedDate > a[j].PublishedDate }

func FindOSImage(imageList []osi.OSImage, label, location string) (osi.OSImage, bool) {
	matches := make([]osi.OSImage, 0)
	for _, im := range imageList {
		for _, loc := range strings.Split(im.Location, ";") {
			if loc == location && im.Label == label {
				matches = append(matches, im)
			}
		}
	}

	if len(matches) > 0 {
		sort.Sort(osImageByPublishDate(matches))
		return matches[0], true
	}
	return osi.OSImage{}, false
}

type osImageByPublishDate []osi.OSImage

func (a osImageByPublishDate) Len() int           { return len(a) }
func (a osImageByPublishDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a osImageByPublishDate) Less(i, j int) bool { return a[i].PublishedDate > a[j].PublishedDate }
