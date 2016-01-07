// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package azure

import (
	vmi "github.com/Azure/azure-sdk-for-go/management/virtualmachineimage"
	"testing"
)

func Test_MatchUserImage(t *testing.T) {

	imageList := []vmi.VMImage{
		{Name: "first", Label: "one"},
		{Name: "something", Label: "label"},
		{Name: "name", Label: "not matching"},
		{Name: "name", Label: "label"},
	}

	tests := []struct {
		name          string
		label         string
		expectFound   bool
		expectedName  string
		expectedLabel string
	}{
		{"name", "label", true, "name", "label"},   // both specified
		{"name", "", true, "name", "not matching"}, // one specified, first in list
		{"", "label", true, "something", "label"},  // one specified, first in list
		{"", "", true, "first", "one"},             // none specified, always match
		{"not", "exist", false, "", ""},            // no match
	}

	for _, tc := range tests {
		image, found := FindVmImage(imageList, tc.name, tc.label)
		if found != tc.expectFound {
			t.Fatalf("tc failed to match 'found': %+v", tc)
		}
		if found {
			if image.Name != tc.expectedName {
				t.Fatalf("tc failed to match 'name': %+v", tc)
			}
			if image.Label != tc.expectedLabel {
				t.Fatalf("tc failed to match 'label': %+v", tc)
			}
		}
	}
}
