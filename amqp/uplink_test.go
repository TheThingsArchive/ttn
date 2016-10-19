package amqp

import (
	"fmt"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestPublishUplink(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestPublishUplink"), fmt.Sprintf("amqp://guest:guest@%s:5672/", host), "test")
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldBeNil)

	err = c.PublishUplink(UplinkMessage{
		AppID:      "app",
		DevID:      "test",
		PayloadRaw: []byte{0x01, 0x08},
	})
	a.So(err, ShouldBeNil)
}
