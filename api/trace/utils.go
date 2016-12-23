// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package trace

import (
	"encoding/base64"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/random"
)

var identifierPrefix = ""

// SetIdentifierPrefix sets the prefix that should be included in generated Trace identifiers
func SetIdentifierPrefix(prefix string) {
	identifierPrefix = prefix
}

func getIdentifier() string {
	rnd := base64.RawURLEncoding.EncodeToString(random.Bytes(24))
	return identifierPrefix + rnd
}

// WithEvent returns a new Trace for the event and its metadata, with the original trace as its parent
func (m *Trace) WithEvent(event string, metadata map[string]string) *Trace {
	t := &Trace{
		Identifier: getIdentifier(),
		Time:       time.Now().UnixNano(),
		Event:      event,
		Metadata:   metadata,
	}
	if m != nil {
		t.Parents = append(t.Parents, m)
	}
	return t
}
