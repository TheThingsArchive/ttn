package gateway

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestNewGateway(t *testing.T) {
	a := New(t)
	gtw := NewGateway(GetLogger(t, "TestNewGateway"), types.GatewayEUI{1, 2, 3, 4, 5, 6, 7})
	a.So(gtw, ShouldNotBeNil)
}
