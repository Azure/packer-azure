// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	osi "github.com/Azure/azure-sdk-for-go/management/osimage"
	vmi "github.com/Azure/azure-sdk-for-go/management/virtualmachineimage"
	"regexp"
	"sort"
	"strings"
)

func GetImageNameRegexp(name string) *regexp.Regexp {
	return regexp.MustCompile(strings.Replace(name, ".", `\.`, -1) + "$")
}

func FindVmImage(imageList []vmi.VMImage, name, label string) (vmi.VMImage, bool) {

	imageNameRegexp := GetImageNameRegexp(name)
	matches := make([]vmi.VMImage, 0)
	for _, im := range imageList {
		if (len(label) == 0 || im.Label == label) &&
			(len(name) == 0 || imageNameRegexp.MatchString(im.Name)) {
			matches = append(matches, im)
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

func FindOSImage(imageList []osi.OSImage, name, label, location string) (osi.OSImage, bool) {

	imageNameRegexp := GetImageNameRegexp(name)
	matches := make([]osi.OSImage, 0)
	for _, im := range imageList {
		for _, loc := range strings.Split(im.Location, ";") {
			if loc == location &&
				(len(label) == 0 || im.Label == label) &&
				(len(name) == 0 || imageNameRegexp.MatchString(im.Name)) {
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
