// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package utils

// removes overlap between the end of a and the start of b and
// glues them together
func GlueStrings(a, b string) string {
	shift := 0
	for shift < len(a) {
		i := 0
		for (i+shift < len(a)) && (i < len(b)) && (a[i+shift] == b[i]) {
			i++
		}
		if i+shift == len(a) {
			break
		}
		shift++
	}

	return string(a[:shift]) + b
}
