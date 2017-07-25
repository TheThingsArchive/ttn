// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cayennelpp

import (
	pb_handler "github.com/TheThingsNetwork/api/handler"
	protocol "github.com/TheThingsNetwork/go-cayenne-lib/cayennelpp"
)

// Encoder is a CayenneLPP PayloadEncoder
type Encoder struct {
}

// Encode encodes the fields to CayenneLPP
func (e *Encoder) Encode(fields map[string]interface{}, fPort uint8) ([]byte, bool, error) {
	encoder := protocol.NewEncoder()
	for name, value := range fields {
		key, channel, err := parseName(name)
		if err != nil {
			continue
		}
		switch key {
		case valueKey:
			if val, ok := value.(float64); ok {
				encoder.AddPort(channel, float32(val))
			}
		}
	}
	return encoder.Bytes(), true, nil
}

// Log returns the log
func (e *Encoder) Log() []*pb_handler.LogEntry {
	return nil
}
