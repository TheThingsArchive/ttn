// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

// ListOptions are options for all list commands
type ListOptions struct {
	Limit  int
	Offset int

	total    int
	selected int
}

// GetTotalAndSelected returns the total number of items, along with the number of selected items
func (o ListOptions) GetTotalAndSelected() (total, selected int) {
	return o.total, o.selected
}

func selectKeys(keys []string, options *ListOptions) []string {
	var start int
	var end = len(keys)
	if options != nil {
		options.total = end
		if options.Offset >= options.total {
			return []string{}
		}
		start = options.Offset
		if options.Limit > 0 {
			if options.Offset+options.Limit > options.total {
				options.Limit = options.total - options.Offset
			}
			end = options.Offset + options.Limit
		}
		options.selected = end - start
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
