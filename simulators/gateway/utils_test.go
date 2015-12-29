// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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

}

func TestGenerateRssi(t *testing.T) {
	Convey("The generateRSSI should generate random RSSI values -120 < val < 0", t, func() {
		values := make(map[int]bool)
		for i := 0; i < 10; i += 1 {
			rssi := generateRssi()
			So(rssi, ShouldBeGreaterThanOrEqualTo, -120)
			So(rssi, ShouldBeLessThanOrEqualTo, 0)
			values[rssi] = true
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
		}
		So(len(values), ShouldBeGreaterThan, 5)
	})
}
