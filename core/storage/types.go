// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import "time"

// Time is a wrapper around time.Time that marshals and unmarshals to the time.RFC3339Nano format
type Time struct {
	time.Time
}

// MarshalText implements the encoding.TextMarshaler interface
func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.Time.UTC().Format(time.RFC3339Nano)), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (t *Time) UnmarshalText(in []byte) error {
	parsed, err := time.Parse(time.RFC3339Nano, string(in))
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}
