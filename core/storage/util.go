// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

// ListOptions are options for all list commands
type ListOptions struct {
	Limit  int
	Offset int
}

func selectKeys(keys []string, options *ListOptions) []string {
	var start int
	var end = len(keys)
	if options != nil {
		if options.Offset >= len(keys) {
			return []string{}
		}
		start = options.Offset
		if options.Limit > 0 {
			if options.Offset+options.Limit > len(keys) {
				options.Limit = len(keys) - options.Offset
			}
			end = options.Offset + options.Limit
		}
	}
	return keys[start:end]
}

func stringInSlice(search string, slice []string) bool {
	for _, i := range slice {
		if i == search {
			return true
		}
	}
	return false
}
