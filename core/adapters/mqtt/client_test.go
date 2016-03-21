// Copyright © 2016 T//e Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

const BrokerAddr = "0.0.0.0:1883"

// MockClient implements the Client interface
type MockClient struct {
	Failures  map[string]error
	InPublish struct {
		Options *client.PublishOptions
	}
	InTerminate struct {
		Called bool
	}
}

// NewMockClient constructs a new MockClient
func NewMockClient() *MockClient {
	return &MockClient{
		Failures: make(map[string]error),
	}
}

// Publish implements the Client interface
func (m *MockClient) Publish(o *client.PublishOptions) error {
	m.InPublish.Options = o
	return m.Failures["Publish"]
}

// Terminate implements the Client interface
func (m *MockClient) Terminate() {
	m.InTerminate.Called = true
}

// MockTryConnect simulates a tryConnect function
type MockTryConnect struct {
	Func    connecter
	Attempt int
}

// NewMockTryConnect instantiates a MockTryConnect structure
func NewMockTryConnect(maxFailures int) *MockTryConnect {
	m := new(MockTryConnect)

	m.Func = func() error {
		m.Attempt++
		if m.Attempt > maxFailures {
			return nil
		}
		return fmt.Errorf("MockTryConnect: Nope")
	}

	return m
}

func TestCreateErrorHandler(t *testing.T) {
	{
		Desc(t, "Try reconnect once and fail")

		// Build
		tryConnect := NewMockTryConnect(3)
		maxDelay := InitialReconnectDelay
		handler := createErrorHandler(tryConnect.Func, maxDelay, GetLogger(t, "Test Logger"))

		// Operate
		go handler(fmt.Errorf("Mock Failure"))
		<-time.After(maxDelay + 50*time.Millisecond)

		// Check
		Check(t, 1, tryConnect.Attempt, "Reconnection Attempts")
	}

	// --------------------

	{
		Desc(t, "Try reconnect more than once and fail")

		// Build
		tryConnect := NewMockTryConnect(3)
		maxDelay := InitialReconnectDelay * 10
		handler := createErrorHandler(tryConnect.Func, maxDelay, GetLogger(t, "Test Logger"))

		// Operate
		go handler(fmt.Errorf("Mock Failure"))
		<-time.After(maxDelay + 50*time.Millisecond)

		// Check
		Check(t, 2, tryConnect.Attempt, "Reconnection Attempts")
	}

	// --------------------

	{
		Desc(t, "Try reconnect once and succeed")

		// Build
		tryConnect := NewMockTryConnect(0)
		maxDelay := InitialReconnectDelay
		handler := createErrorHandler(tryConnect.Func, maxDelay, GetLogger(t, "Test Logger"))

		// Operate
		go handler(fmt.Errorf("Mock Failure"))
		<-time.After(maxDelay + 50*time.Millisecond)

		// Check
		Check(t, 1, tryConnect.Attempt, "Reconnection Attempts")
	}
}

func newID() string {
	return fmt.Sprintf("(%d)Client", time.Now().Nanosecond())
}

func TestNewClient(t *testing.T) {
	testCli := client.New(nil)
	if err := testCli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  BrokerAddr,
		ClientID: []byte(newID()),
	}); err != nil {
		panic(err)
	}

	// --------------------

	{
		Desc(t, "Create client with invalid address")

		// Build
		_, _, err := NewClient(newID(), "invalidAddress", GetLogger(t, "Test Logger"))

		// Check
		CheckErrors(t, ErrOperational, err)
	}

	// --------------------

	{
		Desc(t, "Connect a client and receive a down msg")

		// Build
		cli, chmsg, err := NewClient(newID(), BrokerAddr, GetLogger(t, "Test Logger"))
		FatalUnless(t, err)
		msg := Msg{
			Type:    Down,
			Topic:   "01020304/devices/01020304/down",
			Payload: []byte("message"),
		}

		// Operate
		testCli.Publish(&client.PublishOptions{
			QoS:       mqtt.QoS2,
			Retain:    false,
			TopicName: []byte(msg.Topic),
			Message:   msg.Payload,
		})

		// Check
		var got Msg
		select {
		case got = <-chmsg:
		case <-time.After(75 * time.Millisecond):
		}
		Check(t, msg, got, "MQTT Messages")

		// Clean
		cli.Terminate()
		<-time.After(time.Millisecond * 50)
	}

	// --------------------

	{
		Desc(t, "Connect a client and send on a random topic")

		// Build
		cli, chmsg, err := NewClient(newID(), BrokerAddr, GetLogger(t, "Test Logger"))
		FatalUnless(t, err)
		var msg Msg

		// Operate
		testCli.Publish(&client.PublishOptions{
			QoS:       mqtt.QoS2,
			Retain:    false,
			TopicName: []byte("topic"),
			Message:   []byte{14, 42},
		})

		// Check
		var got Msg
		select {
		case got = <-chmsg:
		case <-time.After(75 * time.Millisecond):
		}
		Check(t, msg, got, "MQTT Messages")

		// Clean
		cli.Terminate()
		<-time.After(time.Millisecond * 50)
	}

	// --------------------

	{
		Desc(t, "Connect the client and simulate a disconnection")

		// Build
		id := newID()
		cli, chmsg, err := NewClient(id, BrokerAddr, GetLogger(t, "Test Logger"))
		FatalUnless(t, err)
		msg := Msg{
			Type:    Down,
			Topic:   "0102030405060708/devices/0102030401020304/down",
			Payload: []byte("message"),
		}

		// Operate
		usurp := client.New(nil)
		err = usurp.Connect(&client.ConnectOptions{
			Network:  "tcp",
			Address:  BrokerAddr,
			ClientID: []byte(id),
		})
		FatalUnless(t, err)
		<-time.After(InitialReconnectDelay * 2)
		testCli.Publish(&client.PublishOptions{
			QoS:       mqtt.QoS2,
			Retain:    false,
			TopicName: []byte(msg.Topic),
			Message:   msg.Payload,
		})

		// Check
		var got Msg
		select {
		case got = <-chmsg:
		case <-time.After(75 * time.Millisecond):
		}
		Check(t, msg, got, "MQTT Messages")

		// Clean
		cli.Terminate()
		_ = usurp.Disconnect()
		usurp.Terminate()
		<-time.After(time.Millisecond * 50)
	}

	// --------------------

	{
		Desc(t, "Connect a client and publish on a topic")

		// Build
		chmsg := make(chan Msg)
		msg := Msg{
			Type:    Down,
			Topic:   "topic",
			Payload: []byte{14, 42},
		}
		cli, _, err := NewClient(newID(), BrokerAddr, GetLogger(t, "Test Logger"))
		FatalUnless(t, err)
		err = testCli.Subscribe(&client.SubscribeOptions{
			SubReqs: []*client.SubReq{
				&client.SubReq{
					TopicFilter: []byte("topic"),
					QoS:         mqtt.QoS2,
					Handler: func(topic, msg []byte) {
						chmsg <- Msg{
							Topic:   string(topic),
							Payload: msg,
							Type:    Down,
						}
					},
				},
			},
		})
		FatalUnless(t, err)

		// Operate
		err = cli.Publish(&client.PublishOptions{
			QoS:       mqtt.QoS2,
			Retain:    false,
			TopicName: []byte(msg.Topic),
			Message:   msg.Payload,
		})

		// Check
		var got Msg
		select {
		case got = <-chmsg:
		case <-time.After(75 * time.Millisecond):
		}

		CheckErrors(t, nil, err)
		Check(t, msg, got, "MQTT Messages")

		// Clean
		err = testCli.Unsubscribe(&client.UnsubscribeOptions{
			TopicFilters: [][]byte{[]byte(msg.Topic)},
		})
		FatalUnless(t, err)
		cli.Terminate()
		<-time.After(time.Millisecond * 50)
	}

	// --------------------

	_ = testCli.Disconnect()
	testCli.Terminate()
}

func TestMonitorClient(t *testing.T) {
	{
		Desc(t, "Ensure monitor stops when cmd is closed, without any client")

		// Build
		chcmd := make(chan interface{})
		chdone := make(chan bool)

		// Operate
		go func() {
			monitorClient(newID(), BrokerAddr, chcmd, GetLogger(t, "Test Client"))
			chdone <- true
		}()
		close(chcmd)

		// Check
		var done bool
		select {
		case done = <-chdone:
		case <-time.After(time.Millisecond * 50):
		}
		Check(t, true, done, "Done signals")

	}

	// --------------------

	{
		Desc(t, "Ensure monitor stops when cmd is closed, with a client")

		// Build
		cli := client.New(nil)
		chcmd := make(chan interface{})
		chdone := make(chan bool)
		cherr := make(chan error)

		// Operate
		go func() {
			monitorClient(newID(), BrokerAddr, chcmd, GetLogger(t, "Test Client"))
			chdone <- true
		}()
		chcmd <- cmdClient{cherr: cherr, options: cli}
		<-cherr
		close(chcmd)

		// Check
		var done bool
		select {
		case done = <-chdone:
		case <-time.After(time.Millisecond * 50):
		}
		Check(t, true, done, "Done signals")

	}

	// --------------------

	{
		Desc(t, "Send an invalid command to monitor, no client")

		// Build
		chcmd := make(chan interface{})
		chdone := make(chan bool)

		// Operate
		go func() {
			monitorClient(newID(), BrokerAddr, chcmd, GetLogger(t, "Test Client"))
			chdone <- true
		}()
		chcmd <- "Patate"

		// Check
		var done bool
		select {
		case done = <-chdone:
		case <-time.After(time.Millisecond * 50):
		}
		Check(t, false, done, "Done signals")
	}

	// --------------------

	{
		Desc(t, "Send an invalid command to monitor, no client")

		// Build
		cli := client.New(nil)
		chcmd := make(chan interface{})
		chdone := make(chan bool)
		cherr := make(chan error)

		// Operate
		go func() {
			monitorClient(newID(), BrokerAddr, chcmd, GetLogger(t, "Test Client"))
			chdone <- true
		}()
		chcmd <- cmdClient{cherr: cherr, options: cli}
		<-cherr
		chcmd <- "Patate"

		// Check
		var done bool
		select {
		case done = <-chdone:
		case <-time.After(time.Millisecond * 50):
		}
		Check(t, false, done, "Done signals")

		// Clean
		_ = cli.Disconnect()
		cli.Terminate()
	}
}
