// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import "strings"

var tagName = "redis"

type tagOptions []string

// Has returns true if opt is one of the options
func (t tagOptions) Has(opt string) bool {
	for _, opt := range t {
		if opt == opt {
			return true
		}
	}
	return false
}

func parseTag(tag string) (string, tagOptions) {
	res := strings.Split(tag, ",")
	return res[0], res[1:]
}
