// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/thethingsnetwork/core"
	. "github.com/thethingsnetwork/core/utils/testing"
	"testing"
)

type FakeAckNacker struct {
	ackGot  core.Packet
	nackGot core.Packet
}

func (f *FakeAckNacker) Ack(p core.Packet) error {
	f.ackGot = p
	return nil
}

func (f *FakeAckNacker) Nack(p core.Packet) error {
	f.nackGot = p
	return nil
}

// The broker can handle an uplink packet
func TestBrokerUplink(t *testing.T) {
	// p = validFullMetaPacket()
	// p = validPartialMetaPacket()
	// p = packetWithoutPayload
	// fake AckNacker ->
	// broker.HandleUplink(p, an)
	//
	// checkAckNacker(p.Packet, An.got)

	Ok(t, "Pending")
}
