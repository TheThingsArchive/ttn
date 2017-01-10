// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
	"gopkg.in/redis.v5"
)

func getDevAddr(bytes ...byte) (addr types.DevAddr) {
	copy(addr[:], bytes[:4])
	return
}

func getEUI(bytes ...byte) (eui types.EUI64) {
	copy(eui[:], bytes[:8])
	return
}

func TestNewNetworkServer(t *testing.T) {
	a := New(t)
	var client redis.Client

	// TTN NetID
	ns := NewRedisNetworkServer(&client, 19)
	a.So(ns, ShouldNotBeNil)
	a.So(ns.(*networkServer).netID, ShouldEqual, [3]byte{0, 0, 0x13})

	// Other NetID, same NwkID
	ns = NewRedisNetworkServer(&client, 66067)
	a.So(ns, ShouldNotBeNil)
	a.So(ns.(*networkServer).netID, ShouldEqual, [3]byte{0x01, 0x02, 0x13})
}

func TestUsePrefix(t *testing.T) {
	a := New(t)
	var client redis.Client
	ns := NewRedisNetworkServer(&client, 19)

	a.So(ns.UsePrefix(types.DevAddrPrefix{DevAddr: types.DevAddr([4]byte{0, 0, 0, 0}), Length: 0}, []string{"otaa"}), ShouldNotBeNil)
	a.So(ns.UsePrefix(types.DevAddrPrefix{DevAddr: types.DevAddr([4]byte{0x14, 0, 0, 0}), Length: 7}, []string{"otaa"}), ShouldNotBeNil)
	a.So(ns.UsePrefix(types.DevAddrPrefix{DevAddr: types.DevAddr([4]byte{0x26, 0, 0, 0}), Length: 7}, []string{"otaa"}), ShouldBeNil)
	a.So(ns.(*networkServer).prefixes, ShouldHaveLength, 1)
}
