// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"errors"
	"testing"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestNewRegistry(t *testing.T) {
	a := New(t)
	r := NewRegistry(GetLogger(t, "TestNewRegistry"))
	a.So(r, ShouldNotBeNil)
}

func TestInitClient(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestInitClient")

	t.Run("RegistersValidClients", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnClient

		r.InitClient("", "")

		a.So(r.monitorClients, ShouldHaveLength, 1)
	})

	t.Run("HandlesInitError", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnError

		r.InitClient("", "")

		a.So(r.monitorClients, ShouldHaveLength, 0)
	})
}

func TestBrokerClients(t *testing.T) {
	a := New(t)
	r := NewRegistry(GetLogger(t, "TestBrokerClients")).(*registry)
	r.newMonitorClient = returnClient

	r.InitClient("a", "")
	r.InitClient("b", "")

	a.So(r.BrokerClients(), ShouldHaveLength, 2)
}

func TestGatewayClients(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestGatewayClients")

	t.Run("Init", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnClient

		r.InitClient("a", "")

		a.So(r.GatewayClients("foo"), ShouldHaveLength, 1)
	})

	t.Run("GatewayClients -> Init", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnClient

		r.GatewayClients("foo")
		r.InitClient("a", "")

		a.So(r.GatewayClients("foo"), ShouldHaveLength, 1)
	})

	t.Run("GatewayClients -> Init -> Token", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnClient

		r.GatewayClients("foo")
		r.InitClient("", "")
		r.SetGatewayToken("foo", "bar")

		a.So(r.GatewayClients("foo")[0].(*gatewayClient).token, ShouldEqual, "bar")
	})

	t.Run("GatewayClients -> Token -> Init", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnClient

		r.GatewayClients("foo")
		r.SetGatewayToken("foo", "bar")
		r.InitClient("", "")

		a.So(r.GatewayClients("foo")[0].(*gatewayClient).token, ShouldEqual, "bar")
	})

	t.Run("Init -> GatewayClients -> Token", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnClient

		r.InitClient("", "")
		r.GatewayClients("foo")
		r.SetGatewayToken("foo", "bar")

		a.So(r.GatewayClients("foo")[0].(*gatewayClient).token, ShouldEqual, "bar")
	})

	t.Run("Init -> Token -> GatewayClients", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnClient

		r.InitClient("", "")
		r.SetGatewayToken("foo", "bar")

		a.So(r.GatewayClients("foo")[0].(*gatewayClient).token, ShouldEqual, "bar")
	})

	t.Run("Token -> GatewayClients -> Init", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnClient

		r.SetGatewayToken("foo", "bar")
		r.GatewayClients("foo")
		r.InitClient("", "")

		a.So(r.GatewayClients("foo")[0].(*gatewayClient).token, ShouldEqual, "bar")
	})

	t.Run("Token -> Init -> GatewayClients", func(t *testing.T) {
		r := NewRegistry(ctx).(*registry)
		r.newMonitorClient = returnClient

		r.SetGatewayToken("foo", "bar")
		r.InitClient("a", "")

		a.So(r.GatewayClients("foo")[0].(*gatewayClient).token, ShouldEqual, "bar")
	})
}

var returnClient = func(ctx ttnlog.Interface, addr string) (*Client, error) {
	return &Client{
		Ctx:          ctx,
		BrokerClient: &brokerClient{},
		gateways:     make(map[string]GatewayClient),
	}, nil
}

var returnError = func(ctx ttnlog.Interface, addr string) (*Client, error) {
	return nil, errors.New("")
}
