// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"testing"

	pb_gateway "github.com/TheThingsNetwork/api/gateway"
	. "github.com/smartystreets/assertions"
)

func TestNewGatewayStatusStore(t *testing.T) {
	a := New(t)
	store := NewStatusStore()
	a.So(store, ShouldNotBeNil)
}

func TestStatusGetUpsert(t *testing.T) {
	a := New(t)
	store := NewStatusStore()

	// Get non-existing gateway status -> expect empty
	status, err := store.Get()
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, pb_gateway.Status{})

	// Update -> expect no error
	statusMessage := &pb_gateway.Status{Description: "Fake Gateway"}
	err = store.Update(statusMessage)
	a.So(err, ShouldBeNil)

	// Get existing gateway status -> expect status
	status, err = store.Get()
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, *statusMessage)
}
