package amqp

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestPublishUplink(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestPublishUplink"), "guest", "guest", host, "test")
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldBeNil)

	err = c.PublishUplink(types.UplinkMessage{
		AppID:      "app",
		DevID:      "test",
		PayloadRaw: []byte{0x01, 0x08},
	})
	a.So(err, ShouldBeNil)
}
