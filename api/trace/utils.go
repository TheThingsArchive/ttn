// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package trace

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
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

type byTime []*Trace

func (a byTime) Len() int           { return len(a) }
func (a byTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTime) Less(i, j int) bool { return a[i].Time < a[j].Time }

// Flatten the trace events and sort by timestamp
func (m *Trace) Flatten() []*Trace {
	flattened := []*Trace{m}
	for _, parent := range m.Parents {
		flattened = append(flattened, parent.Flatten()...)
	}
	sort.Sort(byTime(flattened))
	return flattened
}

func (m *Trace) GoString() (out string) {
	flattened := m.Flatten()
	for _, trace := range flattened {
		out += fmt.Sprintf("%d | %s %s | %s", trace.Time, trace.ServiceName, trace.ServiceId, trace.Event)
		for k, v := range trace.Metadata {
			out += fmt.Sprintf(" (%s=%s)", k, v)
		}
		out += "\n"
	}
	return strings.TrimSpace(out)
}
