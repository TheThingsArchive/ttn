// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import "bytes"

// AddMetadata adds metadata to the announcement if it doesn't exist
func (announcement *Announcement) AddMetadata(key Metadata_Key, value []byte) {
	for _, meta := range announcement.Metadata {
		if meta.Key == key && bytes.Equal(value, meta.Value) {
			return
		}
	}
	announcement.Metadata = append(announcement.Metadata, &Metadata{
		Key:   key,
		Value: value,
	})
}

// DeleteMetadata deletes metadata from the announcement if it exists
func (announcement *Announcement) DeleteMetadata(key Metadata_Key, value []byte) {
	newMeta := make([]*Metadata, 0, len(announcement.Metadata))
	for _, meta := range announcement.Metadata {
		if !(meta.Key == key && bytes.Equal(value, meta.Value)) {
			newMeta = append(newMeta, meta)
		}
	}
	announcement.Metadata = newMeta
}
