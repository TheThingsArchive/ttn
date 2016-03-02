// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

// Recipient describes recipient manipulated by the mqtt adapter
type Recipient interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	TopicUp() string
	TopicDown() string
}

// recipient implements the MqttRecipient interface
type recipient struct {
	up   string
	down string
}

// NewRecipient creates a new MQTT recipient from two topics
func NewRecipient(up string, down string) Recipient {
	return &recipient{up: up, down: down}
}

// TopicUp implements the Recipient interface
func (r recipient) TopicUp() string {
	return r.up
}

// TopicDown implements the Recipient interface
func (r recipient) TopicDown() string {
	return r.down
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r recipient) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(r.up)
	rw.Write(r.down)

	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *recipient) UnmarshalBinary(data []byte) error {
	if r == nil {
		return errors.New(errors.Structural, "Cannot unmarshal nil structure")
	}

	rw := readwriter.New(data)
	rw.Read(func(data []byte) { r.up = string(data) })
	rw.Read(func(data []byte) { r.down = string(data) })

	if err := rw.Err(); err != nil {
		return errors.New(errors.Structural, err)
	}
	return nil
}
