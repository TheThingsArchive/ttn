// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/semtech"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGenToken(t *testing.T) {
	Convey("The genToken() method should return randommly generated 2-byte long tokens", t, func() {
		Convey("Given 5 generated tokens", func() {
			randTokens := [5][]byte{
				genToken(),
				genToken(),
				genToken(),
				genToken(),
				genToken(),
			}

			Convey("They shouldn't be all identical", func() {
				sameTokens := [5][]byte{
					randTokens[0],
					randTokens[0],
					randTokens[0],
					randTokens[0],
					randTokens[0],
				}

				So(randTokens, ShouldNotResemble, sameTokens)
			})

			Convey("They should all be 2-byte long", func() {
				for _, t := range randTokens {
					So(len(t), ShouldEqual, 2)
				}
			})
		})
	})
}

func TestAckToken(t *testing.T) {
	token := []byte{0x1, 0x4}
	generatePacket := func(id byte) semtech.Packet {
		return semtech.Packet{
			Token:      token,
			Identifier: id,
			Version:    semtech.VERSION,
		}
	}

	Convey("The ackToken() method should generate appropriate ACK token", t, func() {
		Convey("Valid identifier, PULL", func() {
			token_data := ackToken(14, generatePacket(semtech.PULL_DATA))
			token_ack := ackToken(14, generatePacket(semtech.PULL_ACK))
			token_resp := ackToken(14, generatePacket(semtech.PULL_RESP))
			So(token_data, ShouldResemble, token_ack)
			So(token_ack, ShouldResemble, token_resp)
		})

		Convey("Valid identifier, PUSH", func() {
			token_data := ackToken(14, generatePacket(semtech.PUSH_DATA))
			token_ack := ackToken(14, generatePacket(semtech.PUSH_ACK))
			So(token_data, ShouldResemble, token_ack)
		})

		Convey("Valid but different ids", func() {
			token_data := ackToken(14, generatePacket(semtech.PULL_DATA))
			token_ack := ackToken(42, generatePacket(semtech.PULL_ACK))
			So(token_data, ShouldNotResemble, token_ack)
		})

	})
}

func TestGenerateRssi(t *testing.T) {
	Convey("The generateRSSI should generate random RSSI values -120 < val < 0", t, func() {
		values := make(map[int]bool)
		for i := 0; i < 10; i += 1 {
			rssi := generateRssi()
			So(rssi, ShouldBeGreaterThanOrEqualTo, -120)
			So(rssi, ShouldBeLessThanOrEqualTo, 0)
			values[rssi] = true
			t.Log(rssi)
		}
		So(len(values), ShouldBeGreaterThan, 5)
	})
}

func TestGenerateFreq(t *testing.T) {
	Convey("The generateFreq() method should generate random frequence between 863-870MHz", t, func() {
		values := make(map[float64]bool)
		for i := 0; i < 10; i += 1 {
			freq := generateFreq()
			So(freq, ShouldBeGreaterThanOrEqualTo, 863.0)
			So(freq, ShouldBeLessThanOrEqualTo, 870.0)
			values[freq] = true
			t.Log(freq)
		}
		So(len(values), ShouldBeGreaterThan, 5)
	})
}

func TestGenerateLsnr(t *testing.T) {
	Convey("The generateLsnr() function should generate random snr ratio between 5.5 and -2", t, func() {
		values := make(map[float64]bool)
		for i := 0; i < 10; i += 1 {
			lsnr := generateLsnr()
			So(lsnr, ShouldBeGreaterThanOrEqualTo, -2)
			So(lsnr, ShouldBeLessThanOrEqualTo, 5.5)
			values[lsnr] = true
			t.Log(lsnr)
		}
		So(len(values), ShouldBeGreaterThan, 5)
	})
}
