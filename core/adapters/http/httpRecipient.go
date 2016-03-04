// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"encoding"
	"encoding/json"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

// Recipient represents the recipient used by the http adapter
type Recipient interface {
	json.Marshaler
	encoding.BinaryMarshaler
	URL() string
	Method() string
}

// NewRecipient constructs a new HttpRecipient
func NewRecipient(url string, method string) Recipient {
	return recipient{url: url, method: method}
}

// Recipient materializes recipients manipulated by the http adapter
type recipient struct {
	url    string
	method string
}

// Url implements the Recipient interface
func (h recipient) URL() string {
	return h.url
}

// Method implements the Recipient interface
func (h recipient) Method() string {
	return h.method
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (h recipient) MarshalBinary() ([]byte, error) {
	w := readwriter.New(nil)
	w.Write(h.url)
	w.Write(h.method)
	return w.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (h *recipient) UnmarshalBinary(data []byte) error {
	r := readwriter.New(data)
	r.Read(func(data []byte) { h.url = string(data) })
	r.Read(func(data []byte) { h.method = string(data) })
	return r.Err()
}

// MarshalJSON implements the encoding/json.Marshaler interface
func (h recipient) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Url    string `json:"url"`
		Method string `json:"method"`
	}{
		Url:    h.url,
		Method: h.method,
	})
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	return data, nil
}
