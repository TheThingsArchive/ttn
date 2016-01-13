// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding/json"
	"fmt"
	"github.com/thethingsnetwork/core/semtech"
	"github.com/thethingsnetwork/core/utils/pointer"
	"time"
)

type metadata Metadata

// MarshalJSON implements the json.Marshal interface
func (m Metadata) MarshalJSON() ([]byte, error) {
	var d *semtech.Datrparser
	var t *semtech.Timeparser

	if m.Datr != nil {
		d = new(semtech.Datrparser)
		if m.Modu != nil && *m.Modu == "FSK" {
			*d = semtech.Datrparser{Kind: "uint", Value: *m.Datr}
		} else {
			*d = semtech.Datrparser{Kind: "string", Value: *m.Datr}
		}
	}

	if m.Time != nil {
		t = new(semtech.Timeparser)
		*t = semtech.Timeparser{Layout: time.RFC3339Nano, Value: m.Time}
	}

	return json.Marshal(metadataProxy{
		metadata: metadata(m),
		Datr:     d,
		Time:     t,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (m *Metadata) UnmarshalJSON(raw []byte) error {
	if m == nil {
		return fmt.Errorf("Cannot unmarshal nil Metadata")
	}

	proxy := metadataProxy{}
	if err := json.Unmarshal(raw, &proxy); err != nil {
		return err
	}
	*m = Metadata(proxy.metadata)
	if proxy.Time != nil {
		m.Time = proxy.Time.Value
	}

	if proxy.Datr != nil {
		m.Datr = &proxy.Datr.Value
	}

	return nil
}

// String implement the io.Stringer interface
func (m Metadata) String() string {
	return pointer.DumpPStruct(m, false)
}

// type metadataProxy is used to conveniently marshal and unmarshal Metadata structure.
//
// Datr field could be either string or uint depending on the Modu field.
// Time field could be parsed in a lot of different way depending of the time format.
// This proxy make sure that everything is marshaled and unmarshaled to the right thing and allow
// the Metadata struct to be user-friendly.
type metadataProxy struct {
	metadata
	Datr *semtech.Datrparser `json:"datr,omitempty"`
	Time *semtech.Timeparser `json:"time,omitempty"`
}
