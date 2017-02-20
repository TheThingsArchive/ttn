// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/router"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestClient(t *testing.T) {
	a := New(t)

	ctx := GetLogger(t, "Monitor Client")
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(1000) + 10000
	go startExampleServer(2, port)

	{
		client, err := NewClient(ctx, fmt.Sprintf("localhost:%d", port))
		a.So(err, ShouldBeNil)
		a.So(client.IsConnected(), ShouldBeTrue)
		a.So(client.Reopen(), ShouldBeNil)
		a.So(client.IsConnected(), ShouldBeTrue)
		a.So(client.Close(), ShouldBeNil)
	}

	{
		client, _ := NewClient(ctx, fmt.Sprintf("localhost:%d", port))
		gtw := client.GatewayClient("dev")
		a.So(gtw.IsConfigured(), ShouldBeFalse)

		gtwCl := gtw.(*gatewayClient)

		for _, token := range []string{
			"SOME.AWESOME.JWT", "SOME.COOLER.JWT",
		} {

			time.AfterFunc(100*time.Millisecond, func() { gtw.SetToken(token) })
			select {
			case <-gtw.TokenChanged():
			case <-time.After(200 * time.Millisecond):
				t.Error("Token failed to update")
			}

			a.So(gtw.IsConfigured(), ShouldBeTrue)

			ctx := gtwCl.Context()
			id, _ := api.IDFromContext(ctx)
			a.So(id, ShouldEqual, "dev")
			ctxToken, _ := api.TokenFromContext(ctx)
			a.So(ctxToken, ShouldEqual, token)
		}

		a.So(client.Close(), ShouldBeNil)
	}

	{
		client, _ := NewClient(ctx, fmt.Sprintf("localhost:%d", port))
		defer client.Close()
		gtw := client.GatewayClient("dev")

		err := gtw.SendStatus(&gateway.Status{})
		a.So(err, ShouldBeNil)

		gtw.SetToken("SOME.AWESOME.JWT")

		// The first two statuses are OK
		for i := 0; i < 2; i++ {
			err = gtw.SendStatus(&gateway.Status{})
			a.So(err, ShouldBeNil)
		}

		// The next one will cause an error on the test server
		err = gtw.SendStatus(&gateway.Status{})
		time.Sleep(10 * time.Millisecond)

		// Then, we are going to buffer 10 statuses locally
		for i := 0; i < 10; i++ {
			err = gtw.SendStatus(&gateway.Status{})
			a.So(err, ShouldBeNil)
		}

		// After which statuses will get dropped
		gtw.SendStatus(&gateway.Status{})
		err = gtw.SendStatus(&gateway.Status{})
		a.So(err, ShouldNotBeNil)

		time.Sleep(100 * time.Millisecond)
	}

	{
		client, _ := NewClient(ctx, fmt.Sprintf("localhost:%d", port))
		defer client.Close()
		gtw := client.GatewayClient("dev")

		err := gtw.SendUplink(&router.UplinkMessage{})
		a.So(err, ShouldBeNil)

		gtw.SetToken("SOME.AWESOME.JWT")

		// The first two messages are OK
		for i := 0; i < 2; i++ {
			err = gtw.SendUplink(&router.UplinkMessage{})
			a.So(err, ShouldBeNil)
		}

		// The next one will cause an error on the test server
		err = gtw.SendUplink(&router.UplinkMessage{})
		time.Sleep(10 * time.Millisecond)

		// Then, we are going to buffer 10 messages locally
		for i := 0; i < 10; i++ {
			err = gtw.SendUplink(&router.UplinkMessage{})
			a.So(err, ShouldBeNil)
		}

		// After which messages will get dropped
		gtw.SendUplink(&router.UplinkMessage{})
		err = gtw.SendUplink(&router.UplinkMessage{})
		a.So(err, ShouldNotBeNil)

		time.Sleep(100 * time.Millisecond)
	}

	{
		client, _ := NewClient(ctx, fmt.Sprintf("localhost:%d", port))
		defer client.Close()
		gtw := client.GatewayClient("dev")

		err := gtw.SendDownlink(&router.DownlinkMessage{})
		a.So(err, ShouldBeNil)

		gtw.SetToken("SOME.AWESOME.JWT")

		// The first two messages are OK
		for i := 0; i < 2; i++ {
			err = gtw.SendDownlink(&router.DownlinkMessage{})
			a.So(err, ShouldBeNil)
		}

		// The next one will cause an error on the test server
		err = gtw.SendDownlink(&router.DownlinkMessage{})
		time.Sleep(10 * time.Millisecond)

		// Then, we are going to buffer 10 messages locally
		for i := 0; i < 10; i++ {
			err = gtw.SendDownlink(&router.DownlinkMessage{})
			a.So(err, ShouldBeNil)
		}

		// After which messages will get dropped
		gtw.SendDownlink(&router.DownlinkMessage{})
		err = gtw.SendDownlink(&router.DownlinkMessage{})
		a.So(err, ShouldNotBeNil)

		time.Sleep(100 * time.Millisecond)
	}

	{
		client, _ := NewClient(ctx, fmt.Sprintf("localhost:%d", port))
		defer client.Close()

		err := client.BrokerClient.SendUplink(&broker.DeduplicatedUplinkMessage{})
		a.So(err, ShouldBeNil)

		// The first two messages are OK
		for i := 0; i < 2; i++ {
			err = client.BrokerClient.SendUplink(&broker.DeduplicatedUplinkMessage{})
			a.So(err, ShouldBeNil)
		}

		// The next one will cause an error on the test server
		err = client.BrokerClient.SendUplink(&broker.DeduplicatedUplinkMessage{})
		time.Sleep(10 * time.Millisecond)

		// Then, we are going to buffer 10 messages locally
		for i := 0; i < 10; i++ {
			err = client.BrokerClient.SendUplink(&broker.DeduplicatedUplinkMessage{})
			a.So(err, ShouldBeNil)
		}

		// After which messages will get dropped
		client.BrokerClient.SendUplink(&broker.DeduplicatedUplinkMessage{})
		err = client.BrokerClient.SendUplink(&broker.DeduplicatedUplinkMessage{})
		a.So(err, ShouldNotBeNil)

		time.Sleep(100 * time.Millisecond)
	}

	{
		client, _ := NewClient(ctx, fmt.Sprintf("localhost:%d", port))
		defer client.Close()

		err := client.BrokerClient.SendDownlink(&broker.DownlinkMessage{})
		a.So(err, ShouldBeNil)

		// The first two messages are OK
		for i := 0; i < 2; i++ {
			err = client.BrokerClient.SendDownlink(&broker.DownlinkMessage{})
			a.So(err, ShouldBeNil)
		}

		// The next one will cause an error on the test server
		err = client.BrokerClient.SendDownlink(&broker.DownlinkMessage{})
		time.Sleep(10 * time.Millisecond)

		// Then, we are going to buffer 10 messages locally
		for i := 0; i < 10; i++ {
			err = client.BrokerClient.SendDownlink(&broker.DownlinkMessage{})
			a.So(err, ShouldBeNil)
		}

		// After which messages will get dropped
		client.BrokerClient.SendDownlink(&broker.DownlinkMessage{})
		err = client.BrokerClient.SendDownlink(&broker.DownlinkMessage{})
		a.So(err, ShouldNotBeNil)

		time.Sleep(100 * time.Millisecond)
	}
}
