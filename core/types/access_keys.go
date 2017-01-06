// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

// Right is the type that represents a right to do something on TTN
type Right string

// AccessKey is an access key that gives someone the right to do something with
// an application
type AccessKey struct {
	// The friendly name of the access key
	Name string `json:"name"`

	// The hard-to-guess access key
	Key string `json:"key"`

	// The rights associated with the key
	Rights []Right `json:"rights"`
}

// HasRight checks if an AccessKey has a certain right
func (k *AccessKey) HasRight(right Right) bool {
	for _, r := range k.Rights {
		if r == right {
			return true
		}
	}
	return false
}

func (r *Right) String() string {
	return string(*r)
}
