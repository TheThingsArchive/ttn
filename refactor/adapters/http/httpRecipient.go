// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"encoding"

	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

type HttpRecipient interface {
	encoding.BinaryMarshaler
	Url() string
	Method() string
}

// HttpRecipient materializes recipients manipulated by the http adapter
type httpRecipient struct {
	url    string
	method string
}

// Url implements the HttpRecipient interface
func (h httpRecipient) Url() string {
	return h.url
}

// Method implements the HttpRecipient interface
func (h httpRecipient) Method() string {
	return h.method
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (h httpRecipient) MarshalBinary() ([]byte, error) {
	w := readwriter.New(nil)
	w.Write(h.url)
	w.Write(h.method)
	return w.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (h *httpRecipient) UnmarshalBinary(data []byte) error {
	r := readwriter.New(data)
	r.Read(func(data []byte) { h.url = string(data) })
	r.Read(func(data []byte) { h.method = string(data) })
	return r.Err()
}
