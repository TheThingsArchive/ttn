// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package trace

import (
	"encoding/base64"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/random"
)

var _serviceID = ""
var _serviceName = ""

// SetComponent sets the component information
func SetComponent(serviceName, serviceID string) {
	_serviceName = serviceName
	_serviceID = serviceID
}

// WithEvent returns a new Trace for the event and its metadata, with the original trace as its parent
func (m *Trace) WithEvent(event string, metadata map[string]string) *Trace {
	t := &Trace{
		Id:          base64.RawURLEncoding.EncodeToString(random.Bytes(24)), // Generate a random ID by default
		ServiceName: _serviceName,
		ServiceId:   _serviceID,
		Time:        time.Now().UnixNano(),
		Event:       event,
		Metadata:    metadata,
	}
	if m != nil {
		t.Parents = append(t.Parents, m)
	}
	return t
}
