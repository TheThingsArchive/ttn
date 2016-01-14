// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/semtech"
	. "github.com/smartystreets/goconvey/convey"
)

type fakeAdapter struct {
	id       string
	written  []byte
	Downlink chan []byte
}

func newFakeAdapter(id string) *fakeAdapter {
	return &fakeAdapter{
		id:       id,
		written:  []byte{},
		Downlink: make(chan []byte),
	}
}

// Write implement io.Writer interface
func (a *fakeAdapter) Write(p []byte) (int, error) {
	a.written = p
	return len(p), nil
}

// Read implement io.Reader interface
func (a *fakeAdapter) Read(buf []byte) (int, error) {
	raw, ok := <-a.Downlink
	if !ok {
		return 0, fmt.Errorf("Connection has been closed")
	}
	return copy(buf, raw), nil
}

// Close implement io.Closer interface
func (a *fakeAdapter) Close() error {
	close(a.Downlink)
	return nil
}

// generatePacket provides quick Packet generation for test purpose
func generatePacket(identifier byte, id [8]byte) semtech.Packet {
	switch identifier {
	case semtech.PUSH_DATA, semtech.PULL_DATA:
		return semtech.Packet{
			Version:    semtech.VERSION,
			Token:      genToken(),
			Identifier: identifier,
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
	if err != nil {
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
			So(fwd, ShouldBeNil)
		})

		Convey("Invalid: too many adapters", func() {
			var adapters []io.ReadWriteCloser
			for i := 0; i < 300; i += 1 {
				adapters = append(adapters, newFakeAdapter(fmt.Sprintf("%d", i)))
			}
			fwd, err := NewForwarder(id, adapters...)
			So(fwd, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Forward", t, func() {
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
				So(a1.written, ShouldResemble, raw)
				So(a2.written, ShouldResemble, raw)
			}
		}

		checkInvalid := func(identifier byte) func() {
			return func() {
				err := fwd.Forward(generatePacket(identifier, fwd.Id))
				So(err, ShouldNotBeNil)
			}
		}

		Convey("Valid: PUSH_DATA", checkValid(semtech.PUSH_DATA))
		Convey("Invalid: PULL_DATA", checkInvalid(semtech.PULL_DATA))
		Convey("Invalid: PUSH_ACK", checkInvalid(semtech.PUSH_ACK))
		Convey("Invalid: PULL_ACK", checkInvalid(semtech.PULL_ACK))
		Convey("Invalid: PULL_RESP", checkInvalid(semtech.PULL_RESP))
		Convey("Invalid: wrong PUSH_DATA", func() {
			pkt := generatePacket(semtech.PUSH_DATA, fwd.Id)
			pkt.Token = []byte{0x14}
			err := fwd.Forward(pkt)
			So(err, ShouldNotBeNil)
			So(len(a1.written), ShouldEqual, 0)
			So(len(a2.written), ShouldEqual, 0)
		})
	})

	Convey("Flush", t, func() {
		// Make sure we use a complete new forwarder each time
		fwd, a1, a2 := initForwarder(id)
		defer fwd.Stop()

		Convey("Init flush", func() {
			So(fwd.Flush(), ShouldResemble, make([]semtech.Packet, 0))
		})

		Convey("Store incoming valid packet", func() {
			// Make sure the connection is established
			pkt := generatePacket(semtech.PUSH_DATA, id)
			if err := fwd.Forward(pkt); err != nil {
				panic(err)
			}

			// Simulate an ack and a valid response
			ack := generatePacket(semtech.PUSH_ACK, id)
			ack.Token = pkt.Token
			raw, err := semtech.Marshal(ack)
			if err != nil {
				panic(err)
			}
			a1.Downlink <- raw

			// Simulate a resp
			resp := generatePacket(semtech.PULL_RESP, id)
			resp.Token = []byte{0x0, 0x0}
			raw, err = semtech.Marshal(resp)
			if err != nil {
				panic(err)
			}
			a1.Downlink <- raw

			// Flush and check if the response is there
			time.Sleep(time.Millisecond * 50)
			packets := fwd.Flush()
			So(len(packets), ShouldEqual, 1)
			So(packets[0], ShouldResemble, resp)
		})

		Convey("Ignore invalid datagrams", func() {
			packets := fwd.Flush()
			a2.Downlink <- []byte{0x6, 0x8, 0x14}
			time.Sleep(time.Millisecond * 50)
			So(fwd.Flush(), ShouldResemble, packets)
		})

		Convey("Ignore non relevant packets", func() {
			// Simulate a resp
			resp := generatePacket(semtech.PULL_DATA, id)
			resp.Token = []byte{0x0, 0x0}
			raw, err := semtech.Marshal(resp)
			if err != nil {
				panic(err)
			}
			a1.Downlink <- raw

			// Flush and check wether or not the response has been stored
			time.Sleep(time.Millisecond * 50)
			packets := fwd.Flush()
			So(len(packets), ShouldEqual, 0)
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

			sendAndAck := func(a1Ack, a2Ack uint) {
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
				for i := uint(0); i < a1Ack; i += 1 {
					a1.Downlink <- raw
				}

				for i := uint(0); i < a2Ack; i += 1 {
					a2.Downlink <- raw
				}
				time.Sleep(50 * time.Millisecond)
			}

			Convey("ackr: valid packet acknowledged", func() {
				// Send packet + ack
				sendAndAck(1, 1)

				// Check stats
				stats := fwd.Stats()
				So(*stats.Ackr, ShouldEqual, 1)
			})

			Convey("ackr: valid packet partially acknowledged", func() {
				// Send packet + ack
				sendAndAck(1, 0)

				// Check stats
				stats := fwd.Stats()
				So(*stats.Ackr, ShouldEqual, float64(1.0)/float64(2.0))
			})

			Convey("ackr: valid packet several acks from same", func() {
				// Send packet + ack
				sendAndAck(2, 0)

				// Check stats
				stats := fwd.Stats()
				So(*stats.Ackr, ShouldEqual, float64(1.0)/float64(2.0))
			})

			Convey("ackr: valid packet  not ackowledged", func() {
				// Send packet + ack
				sendAndAck(0, 0)

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
