// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"testing"

	pb_gateway "github.com/TheThingsNetwork/api/gateway"
	. "github.com/smartystreets/assertions"
)

func TestBestSNR(t *testing.T) {
	a := New(t)
	best := bestSNR([]*pb_gateway.RxMetadata{
		&pb_gateway.RxMetadata{SNR: 1},
		&pb_gateway.RxMetadata{SNR: 2},
		&pb_gateway.RxMetadata{SNR: 0},
		&pb_gateway.RxMetadata{SNR: 10},
		&pb_gateway.RxMetadata{SNR: -10},
	})
	a.So(best, ShouldEqual, 10)
}

func TestLinkMargin(t *testing.T) {
	a := New(t)
	a.So(linkMargin("SF7BW125", 4.3), ShouldEqual, 11.8)
}
