// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package trace

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/TheThingsNetwork/go-utils/pseudorandom"
)

var rand = pseudorandom.New(time.Now().UnixNano())

var _serviceID = ""
var _serviceName = ""

// SetComponent sets the component information
func SetComponent(serviceName, serviceID string) {
	_serviceName = serviceName
	_serviceID = serviceID
}

// GetIDs gets the IDs of this Trace
func (m *Trace) GetIDs() (uids []string) {
	if m == nil {
		return
	}
	for _, p := range m.Parents {
		uids = append(uids, p.GetIDs()...)
	}
	if m.Id != "" {
		uids = append(uids, m.Id)
	}
	return
}

// GenID generates a random ID
func (m *Trace) GenID() {
	id := make([]byte, 16)
	rand.FillBytes(id)
	m.Id = base64.RawURLEncoding.EncodeToString(id)
}

// WithEvent returns a new Trace for the event and its metadata, with the original trace as its parent
func (m *Trace) WithEvent(event string, keyvalue ...interface{}) *Trace {
	t := &Trace{
		ServiceName: _serviceName,
		ServiceId:   _serviceID,
		Time:        time.Now().UnixNano(),
		Event:       event,
	}
	if m != nil {
		t.Parents = append(t.Parents, m)
	}
	if len(t.GetIDs()) == 0 {
		t.GenID()
	}
	if len(keyvalue) > 0 {
		if len(keyvalue)%2 == 1 {
			panic("Got an odd number of key-value pairs")
		}
		t.Metadata = make(map[string]string)
		var k string
		for i, s := range keyvalue {
			if i%2 == 0 {
				k = fmt.Sprint(s) // it's a key
				continue
			}
			t.Metadata[k] = fmt.Sprint(s) // it's a value
		}
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
