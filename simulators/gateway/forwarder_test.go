// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"testing"
	"time"
)

type fakeAdapter struct {
	id       string
	wrote    []byte
	Downlink chan []byte
	closed   bool
}

func newFakeAdapter(id string) *fakeAdapter {
	return &fakeAdapter{
		id:       id,
		wrote:    []byte{},
		Downlink: make(chan []byte),
		closed:   false,
	}
}

// Write implement io.Writer interface
func (a *fakeAdapter) Write(p []byte) (int, error) {
	fmt.Printf("%v wrote %+x\n", a.id, p)
	a.wrote = p
	return len(p), nil
}

// Read implement io.Reader interface
func (a *fakeAdapter) Read(buf []byte) (int, error) {
	return copy(buf, <-a.Downlink), nil
}

// Close implement io.Closer interface
func (a *fakeAdapter) Close() error {
	fmt.Printf("Connection %v closed\n", a.id)
	a.closed = true
	return nil
}

// generatePacket provides quick Packet generation for test purpose
func generatePacket(identifier byte, id [8]byte) semtech.Packet {
	switch identifier {
	case semtech.PUSH_DATA:
		return semtech.Packet{
			Version:    semtech.VERSION,
			Token:      genToken(),
			Identifier: semtech.PULL_DATA,
			GatewayId:  id[:],
			Payload:    nil,
		}
	case semtech.PULL_DATA:
		return semtech.Packet{
			Version:    semtech.VERSION,
			Token:      genToken(),
			Identifier: semtech.PULL_DATA,
			GatewayId:  id[:],
		}
	default:
		return semtech.Packet{
			Version:    semtech.VERSION,
			Identifier: identifier,
		}
	}
}

// initForwarder is a little helper used to instance adapters and forwarder for test purpose
func initForwarder(id [8]byte) (*Forwarder, *fakeAdapter, *fakeAdapter) {
	a1, a2 := newFakeAdapter("adapter1"), newFakeAdapter("adapter2")
	fwd, err := NewForwarder(id, a1, a2)
	if err == nil {
		panic(err)
	}
	return fwd, a1, a2
}

func TestForwarder(t *testing.T) {
	id := [8]byte{0x1, 0x3, 0x3, 0x7, 0x5, 0xA, 0xB, 0x1}
	Convey("NewForwarder", t, func() {
		Convey("Valid: one adapter", func() {
			fwd, err := NewForwarder(id, newFakeAdapter("1"))
			So(err, ShouldBeNil)
			defer fwd.Stop()
			So(fwd, ShouldNotBeNil)
		})

		Convey("Valid: two adapters", func() {
			fwd, err := NewForwarder(id, newFakeAdapter("1"), newFakeAdapter("2"))
			So(err, ShouldBeNil)
			defer fwd.Stop()
			So(fwd, ShouldNotBeNil)
		})

		Convey("Invalid: no adapter", func() {
			fwd, err := NewForwarder(id)
			So(err, ShouldNotBeNil)
			defer fwd.Stop()
			So(fwd, ShouldBeNil)
		})
	})

	Convey("Forwarder", t, func() {
		fwd, a1, a2 := initForwarder(id)
		defer fwd.Stop()

		checkValid := func(identifier byte) func() {
			return func() {
				pkt := generatePacket(identifier, fwd.Id)
				raw, err := semtech.Marshal(pkt)
				if err != nil {
					t.Errorf("Unexpected error %+v\n", err)
					return
				}
				err = fwd.Forward(pkt)
				So(err, ShouldBeNil)
				So(a1.wrote, ShouldResemble, raw)
				So(a2.wrote, ShouldResemble, raw)
			}
		}

		checkInvalid := func(identifier byte) func() {
			return func() {
				err := fwd.Forward(generatePacket(identifier, fwd.Id))
				So(err, ShouldNotBeNil)
			}
		}

		Convey("Valid: PUSH_DATA", checkValid(semtech.PUSH_DATA))
		Convey("Valid: PULL_DATA", checkInvalid(semtech.PULL_DATA))
		Convey("Invalid: PUSH_ACK", checkInvalid(semtech.PUSH_ACK))
		Convey("Invalid: PULL_ACK", checkInvalid(semtech.PULL_ACK))
		Convey("Invalid: PULL_RESP", checkInvalid(semtech.PULL_RESP))
	})

	Convey("Flush", t, func() {
		// Make sure we use a complete new forwarder each time
		fwd, a1, a2 := initForwarder(id)
		defer fwd.Stop()
		packets := fwd.Flush()
		token := []byte{0x0, 0x0}

		checkBasic := func(upIdentifier byte, downIdentifier byte, nbAdapter uint) func() {
			return func() {
				// First forward a packet
				pktUp := generatePacket(upIdentifier, id)
				if err := fwd.Forward(pktUp); err != nil {
					panic(err)
				}
				// Then simulate a downlink ack with the same token
				pktDown := generatePacket(downIdentifier, id)
				pktDown.Token = pktUp.Token
				raw, err := semtech.Marshal(pktDown)
				if err != nil {
					panic(err)
				}
				a1.Downlink <- raw
				if nbAdapter > 1 {
					a2.Downlink <- raw
				}

				// Check that the above packet has been received, handled and stored
				time.Sleep(50 * time.Millisecond)

				packets := fwd.Flush()
				So(len(packets), ShouldEqual, nbAdapter)
				So(packets[0], ShouldResemble, pktDown)
				if nbAdapter > 1 {
					So(packets[1], ShouldResemble, pktDown)
				}
				So(len(fwd.Flush()), ShouldEqual, 0)
			}
		}

		checkInapropriate := func(identifier byte, token []byte) func() {
			return func() {
				pkt := generatePacket(identifier, id)
				pkt.Token = []byte{token[0] + 0x1, token[1]} // Make sure token are different
				raw, err := semtech.Marshal(pkt)
				if err != nil {
					panic(err)
				}
				a1.Downlink <- raw
				So(fwd.Flush(), ShouldResemble, packets)
			}
		}

		checkNonPacket := func() func() {
			return func() {
				a1.Downlink <- []byte{0x1, 0x2, 0x3, 0x4}
				So(fwd.Flush(), ShouldResemble, packets)
			}
		}

		Convey("Store valid packet: PUSH_ACK", checkBasic(semtech.PUSH_DATA, semtech.PUSH_ACK, 1))
		Convey("Store valid packet: PULL_ACK", checkBasic(semtech.PULL_DATA, semtech.PULL_ACK, 1))
		Convey("Store valid packet: PULL_RESP", checkBasic(semtech.PULL_DATA, semtech.PULL_RESP, 1))
		Convey("Store several valid packet: PUSH_ACK", checkBasic(semtech.PUSH_DATA, semtech.PUSH_ACK, 2))
		Convey("Store several valid packet: PULL_ACK", checkBasic(semtech.PULL_DATA, semtech.PULL_ACK, 2))
		Convey("Store several valid packet: PULL_RESP", checkBasic(semtech.PULL_DATA, semtech.PULL_RESP, 2))

		Convey("Ignore non packet", checkNonPacket())
		Convey("Ignore inapropriate downlink: PUSH_ACK", checkInapropriate(semtech.PUSH_ACK, token))
		Convey("Ignore inapropriate downlink: PULL_DATA", checkInapropriate(semtech.PULL_DATA, token))
		Convey("Ignore inapropriate downlink: PULL_ACK", checkInapropriate(semtech.PULL_ACK, token))
		Convey("Ignore inapropriate downlink: PUSH_DATA", checkInapropriate(semtech.PUSH_DATA, token))
		Convey("Ignore inapropriate downlink: PULL_RESP", checkInapropriate(semtech.PULL_RESP, token))

		Convey("When waiting for ack", func() {
			pktUp := generatePacket(semtech.PUSH_ACK, id)
			if err := fwd.Forward(pktUp); err != nil {
				panic(err)
			}
			Convey("Ignore non packet", checkNonPacket())
			Convey("Ignore inapropriate downlink: PUSH_ACK", checkInapropriate(semtech.PUSH_ACK, pktUp.Token))
			Convey("Ignore inapropriate downlink: PULL_DATA", checkInapropriate(semtech.PULL_DATA, pktUp.Token))
			Convey("Ignore inapropriate downlink: PULL_ACK", checkInapropriate(semtech.PULL_ACK, pktUp.Token))
			Convey("Ignore inapropriate downlink: PUSH_DATA", checkInapropriate(semtech.PUSH_DATA, pktUp.Token))
			Convey("Ignore inapropriate downlink: PULL_RESP", checkInapropriate(semtech.PULL_RESP, pktUp.Token))
		})
	})

	Convey("Stats", t, func() {
		fwd, a1, a2 := initForwarder(id)
		defer fwd.Stop()
		refStats := fwd.Stats()

		Convey("lati, long, alti, time", func() {
			So(refStats.Lati, ShouldNotBeNil)
			So(refStats.Long, ShouldNotBeNil)
			So(refStats.Alti, ShouldNotBeNil)
			So(refStats.Time, ShouldNotBeNil)
			So(refStats.Rxnb, ShouldNotBeNil)
			So(refStats.Rxok, ShouldNotBeNil)
			So(refStats.Ackr, ShouldNotBeNil)
			So(refStats.Rxfw, ShouldNotBeNil)
			So(refStats.Dwnb, ShouldNotBeNil)
			So(refStats.Txnb, ShouldNotBeNil)

		})

		Convey("rxnb / rxok", func() {
			fwd.Forward(generatePacket(semtech.PUSH_DATA, fwd.Id))
			stats := fwd.Stats()
			So(stats.Rxnb, ShouldNotBeNil)
			So(stats.Rxok, ShouldNotBeNil)
			So(*stats.Rxnb, ShouldEqual, *refStats.Rxnb+1)
			So(*stats.Rxok, ShouldEqual, *refStats.Rxok+1)
		})

		Convey("rxfw", func() {
			fwd.Forward(generatePacket(semtech.PUSH_DATA, fwd.Id))
			stats := fwd.Stats()
			So(stats.Rxfw, ShouldNotBeNil)
			So(*stats.Rxfw, ShouldEqual, *refStats.Rxfw+1)
		})

		Convey("ackr", func() {
			Convey("ackr: initial", func() {
				So(*refStats.Ackr, ShouldEqual, 0)
			})

			sendAndAck := func(a1Ack, a2Ack bool) {
				// Send packet + ack
				pkt := generatePacket(semtech.PUSH_DATA, id)
				ack := generatePacket(semtech.PUSH_ACK, id)
				ack.Token = pkt.Token
				raw, err := semtech.Marshal(ack)
				if err != nil {
					panic(err)
				}
				fwd.Forward(pkt)
				time.Sleep(50 * time.Millisecond)
				if a1Ack {
					a1.Downlink <- raw
				}

				if a2Ack {
					a2.Downlink <- raw
				}
			}

			Convey("ackr: valid packet acknowledged", func() {
				// Send packet + ack
				sendAndAck(true, true)

				// Check stats
				stats := fwd.Stats()
				So(*stats.Ackr, ShouldEqual, 1)
			})

			Convey("ackr: valid packet partially acknowledged", func() {
				// Send packet + ack
				sendAndAck(true, false)

				// Check stats
				stats := fwd.Stats()
				So(*stats.Ackr, ShouldEqual, float64(1.0)/float64(2.0))
			})

			Convey("ackr: valid packet  not ackowledged", func() {
				// Send packet + ack
				sendAndAck(false, false)

				// Check stats
				stats := fwd.Stats()
				So(*stats.Ackr, ShouldEqual, *refStats.Ackr)
			})
		})

		// TODO dwnb
		// TODO txnb
	})

	Convey("Stop", t, func() {
		//TODO
	})
}
