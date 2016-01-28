// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"reflect"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// This illustrates the mechanism used by the handler to bufferize connections
// for a while before processing them.
// This is an analogy to what's done in the handler where
// bundleId -> _bundleId
// uplinkBundle -> _uplinkBundle

func TestHandlerBuffering(t *testing.T) {
	// Describe
	Desc(t, "Generate fake bundle traffic")

	// Build
	bundles := make(chan []_uplinkBundle)
	set := make(chan _uplinkBundle)
	received := new([][]_uplinkBundle) // There's a datarace here, but that's okay, chill.

	go _manageBuffers(bundles, set)
	go _consumeBundles(bundles, received)

	b1_1 := _uplinkBundle{_bundleId(1), "bundle1_1"}
	b1_2 := _uplinkBundle{_bundleId(1), "bundle1_2"}
	b1_3 := _uplinkBundle{_bundleId(1), "bundle1_3"}
	b1_4 := _uplinkBundle{_bundleId(1), "bundle1_4"}

	b2_1 := _uplinkBundle{_bundleId(2), "bundle2_1"}
	b2_2 := _uplinkBundle{_bundleId(2), "bundle2_2"}
	b2_3 := _uplinkBundle{_bundleId(2), "bundle2_3"}

	// Operate
	// Expecting [ b1_1, b1_2, b1_3 ]
	go func() {
		set <- b1_1
		<-time.After(BUFFER_DELAY / 3)
		set <- b1_2
		<-time.After(BUFFER_DELAY / 3)
		set <- b1_3
	}()

	// Expecting [ b2_1, b2_2, b2_3 ]
	go func() {
		set <- b2_1
		<-time.After(BUFFER_DELAY / 4)
		set <- b2_2
		<-time.After(BUFFER_DELAY / 4)
		set <- b2_3
	}()

	// Expecting [ b1_4 ]
	go func() {
		<-time.After(BUFFER_DELAY * 2)
		set <- b1_4
	}()

	// Check
	<-time.After(BUFFER_DELAY * 4)
	if len(*received) != 3 {
		Ko(t, "Expected 3 bundles to have been received but got %d", len(*received))
		return
	}
	Ok(t, "Check bundles number")

	for _, bundles := range *received {
		if reflect.DeepEqual(bundles, []_uplinkBundle{b1_1, b1_2, b1_3}) {
			continue
		}
		if reflect.DeepEqual(bundles, []_uplinkBundle{b1_4}) {
			continue
		}
		if reflect.DeepEqual(bundles, []_uplinkBundle{b2_1, b2_2, b2_3}) {
			continue
		}
		Ko(t, "Collected bundles in a non-expected way: %v", bundles)
		return
	}
	Ok(t, "Check bundles shapes")
}

type _bundleId int
type _uplinkBundle struct {
	id   _bundleId
	data string
}

// _setAlarm ~> setAlarm
func _setAlarm(alarm chan<- _bundleId, id _bundleId, delay time.Duration) {
	<-time.After(delay)
	alarm <- id
}

// _manageBuffers ~> manageBuffers
func _manageBuffers(bundles chan<- []_uplinkBundle, set <-chan _uplinkBundle) {
	buffers := make(map[_bundleId][]_uplinkBundle)
	alarm := make(chan _bundleId)

	for {
		select {
		case id := <-alarm:
			b := buffers[id]
			delete(buffers, id)
			go func(b []_uplinkBundle) { bundles <- b }(b)
		case bundle := <-set:
			b := append(buffers[bundle.id], bundle)
			if len(b) == 1 {
				go _setAlarm(alarm, bundle.id, BUFFER_DELAY)
			}
			buffers[bundle.id] = b
		}
	}
}

// _consumeBundles, just for the test
func _consumeBundles(bundles <-chan []_uplinkBundle, received *[][]_uplinkBundle) {
	for b := range bundles {
		*received = append(*received, b)
	}
}
