// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import "time"

// JSONTime is a time.Time that marshals to/from RFC3339Nano format
type JSONTime time.Time

// MarshalText implements the encoding.TextMarshaler interface
func (t JSONTime) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() || time.Time(t).Unix() == 0 {
		return []byte{}, nil
	}
	stamp := time.Time(t).UTC().Format(time.RFC3339Nano)
	return []byte(stamp), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (t *JSONTime) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*t = JSONTime{}
		return nil
	}
	time, err := time.Parse(time.RFC3339Nano, string(text))
	if err != nil {
		return err
	}
	*t = JSONTime(time)
	return nil
}

// BuildTime builds a new JSONTime
func BuildTime(unixNano int64) JSONTime {
	if unixNano == 0 {
		return JSONTime{}
	}
	return JSONTime(time.Unix(0, 0).Add(time.Duration(unixNano)).UTC())
}
