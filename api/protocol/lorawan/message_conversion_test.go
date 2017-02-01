// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package lorawan

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/band"
	. "github.com/smartystreets/assertions"
)

func TestConvertBytes(t *testing.T) {
	a := New(t)
	var err error
	var in, out Message
	mac := in.InitDownlink()
	in.Mic = []byte{1, 2, 3, 4}
	mac.Ack = true
	mac.Adr = true
	mac.AdrAckReq = true
	mac.DevAddr = types.DevAddr([4]byte{1, 2, 3, 4})
	mac.FCnt = 1
	mac.FOpts = []MACCommand{MACCommand{Cid: 0x02, Payload: []byte{0x00, 0x00}}}
	mac.FPending = true
	mac.FPort = 1
	mac.FrmPayload = []byte{1, 2, 3, 4}

	bytes := in.PHYPayloadBytes()

	a.So(bytes, ShouldResemble, []byte{0x60, 0x4, 0x3, 0x2, 0x1, 0xf3, 0x1, 0x0, 0x2, 0x0, 0x0, 0x1, 0x1, 0x2, 0x3, 0x4, 0x1, 0x2, 0x3, 0x4})

	out, err = MessageFromPHYPayloadBytes(bytes)

	a.So(err, ShouldBeNil)
	a.So(out, ShouldResemble, in)
}

func TestConvertPHYPayload(t *testing.T) {
	a := New(t)

	{
		m1 := Message{Mic: []byte{0, 0, 0, 0}}
		m1.MType = MType_UNCONFIRMED_UP
		macPayload := MACPayload{}
		macPayload.FOpts = []MACCommand{
			MACCommand{Cid: 0x02},
		}
		m1.Payload = &Message_MacPayload{MacPayload: &macPayload}
		phy := m1.PHYPayload()
		m2 := MessageFromPHYPayload(phy)
		a.So(m2, ShouldResemble, m1)
	}

	{
		m1 := Message{Mic: []byte{0, 0, 0, 0}}
		m1.MType = MType_JOIN_REQUEST
		joinRequestPayload := JoinRequestPayload{}
		m1.Payload = &Message_JoinRequestPayload{JoinRequestPayload: &joinRequestPayload}
		phy := m1.PHYPayload()
		m2 := MessageFromPHYPayload(phy)
		a.So(m2, ShouldResemble, m1)
	}

	{
		m1 := Message{Mic: []byte{0, 0, 0, 0}}
		m1.MType = MType_JOIN_ACCEPT
		joinAcceptPayload := JoinAcceptPayload{}
		joinAcceptPayload.CfList = &CFList{
			Freq: []uint32{867100000, 867300000, 867500000, 867700000, 867900000},
		}
		m1.Payload = &Message_JoinAcceptPayload{JoinAcceptPayload: &joinAcceptPayload}
		phy := m1.PHYPayload()
		m2 := MessageFromPHYPayload(phy)
		a.So(m2, ShouldResemble, m1)

		phy.MACPayload = &lorawan.DataPayload{Bytes: []byte{0x01, 0x02, 0x03, 0x04}}

		m3 := MessageFromPHYPayload(phy)

		phy = m3.PHYPayload()
	}

}

func TestConvertDataRate(t *testing.T) {
	a := New(t)

	{
		md := &Metadata{
			Modulation: Modulation_LORA,
			DataRate:   "SF7BW125",
		}
		dr, err := md.GetLoRaWANDataRate()
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldResemble, band.DataRate{Modulation: band.LoRaModulation, SpreadFactor: 7, Bandwidth: 125})
	}

	{
		md := &Metadata{
			Modulation: Modulation_FSK,
			BitRate:    50000,
		}
		dr, err := md.GetLoRaWANDataRate()
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldResemble, band.DataRate{Modulation: band.FSKModulation, BitRate: 50000})
	}

	{
		tx := new(TxConfiguration)
		err := tx.SetDataRate(band.DataRate{Modulation: band.LoRaModulation, SpreadFactor: 7, Bandwidth: 125})
		a.So(err, ShouldBeNil)
		a.So(tx.Modulation, ShouldEqual, Modulation_LORA)
		a.So(tx.DataRate, ShouldEqual, "SF7BW125")
	}

	{
		tx := new(TxConfiguration)
		err := tx.SetDataRate(band.DataRate{Modulation: band.FSKModulation, BitRate: 50000})
		a.So(err, ShouldBeNil)
		a.So(tx.Modulation, ShouldEqual, Modulation_FSK)
		a.So(tx.BitRate, ShouldEqual, 50000)
	}

}
