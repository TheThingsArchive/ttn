// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

// MqttRecipient describes recipient manipulated by the mqtt adapter
type MqttRecipient interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	TopicUp() string
	TopicDown() string
}

// mqttRecipient implements the MqttRecipient interface
type mqttRecipient struct {
	up   string
	down string
}

// NewRecipient creates a new MQTT recipient from two topics
func NewRecipient(up string, down string) MqttRecipient {
	return &mqttRecipient{up: up, down: down}
}

// TopicUp implements the MqttRecipient interface
func (r mqttRecipient) TopicUp() string {
	return r.up
}

// TopicDown implements the MqttRecipient interface
func (r mqttRecipient) TopicDown() string {
	return r.down
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (r mqttRecipient) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(r.up)
	rw.Write(r.down)

	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (r *mqttRecipient) UnmarshalBinary(data []byte) error {
	if r == nil {
		return errors.New(errors.Structural, "Cannot unmarshal nil structure")
	}

	rw := readwriter.New(data)
	rw.Read(func(data []byte) { r.up = string(data) })
	rw.Read(func(data []byte) { r.down = string(data) })

	return rw.Err()
}
