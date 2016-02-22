// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

// HttpRecipient materializes recipients manipulated by the http adapter
type httpRecipient struct {
	Url    string
	Method string
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (h *httpRecipient) MarshalBinary() ([]byte, error) {
	w := readwriter.New(nil)
	w.Write(h.Url)
	w.Write(h.Method)
	return w.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (h httpRecipient) UnmarshalBinary(data []byte) error {
	r := readwriter.New(data)
	r.Read(func(data []byte) { h.Url = string(data) })
	r.Read(func(data []byte) { h.Method = string(data) })
	return r.Err()
}
