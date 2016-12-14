// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

// LocationMetadata contains GPS coordinates
type LocationMetadata struct {
	Latitude  float32 `json:"latitude,omitempty"`
	Longitude float32 `json:"longitude,omitempty"`
	Altitude  int32   `json:"altitude,omitempty"`
}
