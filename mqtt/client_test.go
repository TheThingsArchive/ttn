// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

var host string
var sslHost string

func init() {
	host = os.Getenv("MQTT_ADDRESS")
	if host == "" {
		host = "localhost:1883"
	}
	sslHost = os.Getenv("MQTT_SSL_ADDRESS")
	if sslHost == "" {
		sslHost = "iot.eclipse.org:8883"
	}
}

func waitForOK(token Token, a *Assertion) {
	success := token.WaitTimeout(100 * time.Millisecond)
	a.So(success, ShouldBeTrue)
	a.So(token.Error(), ShouldBeNil)
}

func TestToken(t *testing.T) {
	a := New(t)

	okToken := newToken()
	go func() {
		time.Sleep(1 * time.Millisecond)
		okToken.flowComplete()
	}()
	okToken.Wait()
	a.So(okToken.Error(), ShouldBeNil)

	failToken := newToken()
	go func() {
		time.Sleep(1 * time.Millisecond)
		failToken.err = errors.New("Err")
		failToken.flowComplete()
	}()
	failToken.Wait()
	a.So(failToken.Error(), ShouldNotBeNil)

	timeoutToken := newToken()
	timeoutTokenDone := timeoutToken.WaitTimeout(5 * time.Millisecond)
	a.So(timeoutTokenDone, ShouldBeFalse)
}

func TestSimpleToken(t *testing.T) {
	a := New(t)

	okToken := simpleToken{}
	okToken.Wait()
	a.So(okToken.Error(), ShouldBeNil)

	failToken := simpleToken{fmt.Errorf("Err")}
	failToken.Wait()
	a.So(failToken.Error(), ShouldNotBeNil)
}

func TestNewClient(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))
	a.So(c.(*DefaultClient).mqtt, ShouldNotBeNil)
}

func TestConnect(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldBeNil)

	// Connecting while already connected should not change anything
	err = c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldBeNil)
}

func TestConnectWithTLS(t *testing.T) {
	if sslHost == "SKIP" {
		t.Skip("Skipping MQTT/TLS test")
	}

	a := New(t)

	cert, err := ioutil.ReadFile("../.env/mqtt/ca.pem")
	if err != nil {
		t.Errorf("MQTT CA Cert could not be loaded")
	}

	RootCAs.AppendCertsFromPEM(cert)

	c := NewTLSClient(getLogger(t, "Test"), "test", "", "", nil, fmt.Sprintf("ssl://%s", sslHost))

	err = c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldBeNil)
}

func TestConnectInvalidAddress(t *testing.T) {
	a := New(t)
	ConnectRetries = 2
	ConnectRetryDelay = 50 * time.Millisecond
	c := NewClient(getLogger(t, "Test"), "test", "", "", "tcp://localhost:18830") // No MQTT on 18830
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldNotBeNil)
}

func TestConnectInvalidCredentials(t *testing.T) {
	t.Skipf("Need authenticated MQTT for TestConnectInvalidCredentials - Skipping")
}

func TestIsConnected(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))

	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()

	a.So(c.IsConnected(), ShouldBeTrue)
}

func TestDisconnect(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))

	// Disconnecting when not connected should not change anything
	c.Disconnect()
	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()
	c.Disconnect()

	a.So(c.IsConnected(), ShouldBeFalse)
}

func TestRandomTopicPublish(t *testing.T) {
	a := New(t)
	ctx := getLogger(t, "TestRandomTopicPublish")

	c := NewClient(ctx, "test", "", "", fmt.Sprintf("tcp://%s", host))
	c.Connect()
	defer c.Disconnect()

	subToken := c.(*DefaultClient).mqtt.Subscribe("randomtopic", SubscribeQoS, nil)
	waitForOK(subToken, a)
	pubToken := c.(*DefaultClient).mqtt.Publish("randomtopic", PublishQoS, false, []byte{0x00})
	waitForOK(pubToken, a)

	<-time.After(50 * time.Millisecond)

	ctx.Info("This test should have printed one message.")
}

func ExampleNewClient() {
	ctx := apex.Stdout().WithField("Example", "NewClient")
	exampleClient := NewClient(ctx, "ttnctl", "my-app-id", "my-access-key", "eu.thethings.network:1883")
	err := exampleClient.Connect()
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect")
	}
}

var exampleClient Client

func ExampleDefaultClient_SubscribeDeviceUplink() {
	token := exampleClient.SubscribeDeviceUplink("my-app-id", "my-dev-id", func(client Client, appID string, devID string, req types.UplinkMessage) {
		// Do something with the message
	})
	token.Wait()
	if err := token.Error(); err != nil {
		panic(err)
	}
}

func ExampleDefaultClient_PublishDownlink() {
	token := exampleClient.PublishDownlink(types.DownlinkMessage{
		AppID:      "my-app-id",
		DevID:      "my-dev-id",
		FPort:      1,
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	})
	token.Wait()
	if err := token.Error(); err != nil {
		panic(err)
	}
}
