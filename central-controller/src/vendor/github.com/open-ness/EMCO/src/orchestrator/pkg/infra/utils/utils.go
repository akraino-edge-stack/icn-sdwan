// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package utils

// GetSliceSubtract .. (c1 - c2)
func GetSliceSubtract(c1, c2 []string) (diffSlice []string) {
	sliceMap := make(map[string]bool)

	// Create a map for slice-2
	for _, key := range c2 {
		sliceMap[key] = true
	}

	for _, value := range c1 {
		if _, ok := sliceMap[value]; !ok {
			diffSlice = append(diffSlice, value)
		}
	}
	return
}

// GetSliceIntersect .. intersect slices
func GetSliceIntersect(c1 []string, c2 []string) []string {
	var set = make([]string, 0)
	for _, v1 := range c1 {
		for _, v2 := range c2 {
			if v1 == v2 {
				set = append(set, v1)
			}
		}
	}
	return set
}

// GetSliceContains ... return element index, otherwise return -1 and a bool of false.
func GetSliceContains(slice []string, element string) (int, bool) {
	for index, val := range slice {
		if val == element {
			return index, true
		}
	}
	return -1, false
}

// SliceRemove .. Remove a element from slice
func SliceRemove(slice []string, item string) []string {
	for i, element := range slice {
		if element == item {
			slice = append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
