package networkserver

import (
	"testing"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	. "github.com/smartystreets/assertions"
)

func TestBestSNR(t *testing.T) {
	a := New(t)
	best := bestSNR([]*pb_gateway.RxMetadata{
		&pb_gateway.RxMetadata{Snr: 1},
		&pb_gateway.RxMetadata{Snr: 2},
		&pb_gateway.RxMetadata{Snr: 0},
		&pb_gateway.RxMetadata{Snr: 10},
		&pb_gateway.RxMetadata{Snr: -10},
	})
	a.So(best, ShouldEqual, 10)
}

func TestLinkMargin(t *testing.T) {
	a := New(t)
	a.So(linkMargin("SF7BW125", 4.3), ShouldEqual, 11.8)
}
