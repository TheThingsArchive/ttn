// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"sync"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/random"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// QoS indicates the MQTT Quality of Service level.
// 0: The broker/client will deliver the message once, with no confirmation.
// 1: The broker/client will deliver the message at least once, with confirmation required.
// 2: The broker/client will deliver the message exactly once by using a four step handshake.
var (
	PublishQoS   byte = 0x00
	SubscribeQoS byte = 0x00
)

// Client connects to the MQTT server and can publish/subscribe on uplink, downlink and activations from devices
type Client interface {
	Connect() error
	Disconnect()

	IsConnected() bool

	// Uplink pub/sub
	PublishUplink(payload types.UplinkMessage) Token
	PublishUplinkFields(appID string, devID string, fields map[string]interface{}) Token
	SubscribeDeviceUplink(appID string, devID string, handler UplinkHandler) Token
	SubscribeAppUplink(appID string, handler UplinkHandler) Token
	SubscribeUplink(handler UplinkHandler) Token
	UnsubscribeDeviceUplink(appID string, devID string) Token
	UnsubscribeAppUplink(appID string) Token
	UnsubscribeUplink() Token

	// Downlink pub/sub
	PublishDownlink(payload types.DownlinkMessage) Token
	SubscribeDeviceDownlink(appID string, devID string, handler DownlinkHandler) Token
	SubscribeAppDownlink(appID string, handler DownlinkHandler) Token
	SubscribeDownlink(handler DownlinkHandler) Token
	UnsubscribeDeviceDownlink(appID string, devID string) Token
	UnsubscribeAppDownlink(appID string) Token
	UnsubscribeDownlink() Token

	// Event pub/sub
	PublishAppEvent(appID string, eventType types.EventType, payload interface{}) Token
	PublishDeviceEvent(appID string, devID string, eventType types.EventType, payload interface{}) Token
	SubscribeAppEvents(appID string, eventType types.EventType, handler AppEventHandler) Token
	SubscribeDeviceEvents(appID string, devID string, eventType types.EventType, handler DeviceEventHandler) Token
	UnsubscribeAppEvents(appID string, eventType types.EventType) Token
	UnsubscribeDeviceEvents(appID string, devID string, eventType types.EventType) Token

	// Activation pub/sub
	PublishActivation(payload types.Activation) Token
	SubscribeDeviceActivations(appID string, devID string, handler ActivationHandler) Token
	SubscribeAppActivations(appID string, handler ActivationHandler) Token
	SubscribeActivations(handler ActivationHandler) Token
	UnsubscribeDeviceActivations(appID string, devID string) Token
	UnsubscribeAppActivations(appID string) Token
	UnsubscribeActivations() Token
}

// Token is returned on asyncronous functions
type Token interface {
	// Wait for the function to finish
	Wait() bool
	// Wait for the function to finish or return false after a certain time
	WaitTimeout(time.Duration) bool
	// The error associated with the result of the function (nil if everything okay)
	Error() error
}

type simpleToken struct {
	err error
}

// Wait always returns true
func (t *simpleToken) Wait() bool {
	return true
}

// WaitTimeout always returns true
func (t *simpleToken) WaitTimeout(_ time.Duration) bool {
	return true
}

// Error contains the error if present
func (t *simpleToken) Error() error {
	return t.err
}

type token struct {
	sync.RWMutex
	complete chan bool
	ready    bool
	err      error
}

func newToken() *token {
	return &token{
		complete: make(chan bool),
	}
}

func (t *token) Wait() bool {
	t.Lock()
	defer t.Unlock()
	if !t.ready {
		<-t.complete
		t.ready = true
	}
	return t.ready
}

func (t *token) WaitTimeout(d time.Duration) bool {
	t.Lock()
	defer t.Unlock()
	if !t.ready {
		select {
		case <-t.complete:
			t.ready = true
		case <-time.After(d):
		}
	}
	return t.ready
}

func (t *token) flowComplete() {
	close(t.complete)
}

func (t *token) Error() error {
	t.RLock()
	defer t.RUnlock()
	return t.err
}

// DefaultClient is the default MQTT client for The Things Network
type DefaultClient struct {
	opts          *MQTT.ClientOptions
	mqtt          MQTT.Client
	ctx           log.Interface
	subscriptions map[string]MQTT.MessageHandler
}

// NewClient creates a new DefaultClient
func NewClient(ctx log.Interface, id, username, password string, brokers ...string) Client {
	if ctx == nil {
		ctx = log.Get()
	}

	ttnClient := &DefaultClient{
		opts:          MQTT.NewClientOptions(),
		ctx:           ctx,
		subscriptions: make(map[string]MQTT.MessageHandler),
	}

	for _, broker := range brokers {
		ttnClient.opts.AddBroker(broker)
	}

	ttnClient.opts.SetClientID(fmt.Sprintf("%s-%s", id, random.String(16)))
	ttnClient.opts.SetUsername(username)
	ttnClient.opts.SetPassword(password)

	// TODO: Some tuning of these values probably won't hurt:
	ttnClient.opts.SetKeepAlive(30 * time.Second)
	ttnClient.opts.SetPingTimeout(10 * time.Second)

	ttnClient.opts.SetCleanSession(true)

	ttnClient.opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		ctx.Warnf("Received unhandled message: %v", msg)
	})

	var reconnecting bool

	ttnClient.opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		ctx.Warnf("Disconnected (%s). Reconnecting...", err.Error())
		reconnecting = true
	})

	ttnClient.opts.SetOnConnectHandler(func(client MQTT.Client) {
		ctx.Info("Connected to MQTT")
		if reconnecting {
			for topic, handler := range ttnClient.subscriptions {
				ctx.Infof("Re-subscribing to topic: %s", topic)
				ttnClient.subscribe(topic, handler)
			}
			reconnecting = false
		}
	})

	ttnClient.mqtt = MQTT.NewClient(ttnClient.opts)

	return ttnClient
}

var (
	// ConnectRetries says how many times the client should retry a failed connection
	ConnectRetries = 10
	// ConnectRetryDelay says how long the client should wait between retries
	ConnectRetryDelay = time.Second
)

// Connect to the MQTT broker. It will retry for ConnectRetries times with a delay of ConnectRetryDelay between retries
func (c *DefaultClient) Connect() error {
	if c.mqtt.IsConnected() {
		return nil
	}
	var err error
	for retries := 0; retries < ConnectRetries; retries++ {
		token := c.mqtt.Connect()
		finished := token.WaitTimeout(1 * time.Second)
		if !finished {
			c.ctx.Warn("MQTT connection took longer than expected...")
			token.Wait()
		}
		err = token.Error()
		if err == nil {
			break
		}
		c.ctx.Warnf("Could not connect to MQTT Broker (%s). Retrying...", err.Error())
		<-time.After(ConnectRetryDelay)
	}
	if err != nil {
		return fmt.Errorf("Could not connect to MQTT Broker (%s)", err)
	}
	return nil
}

func (c *DefaultClient) publish(topic string, msg []byte) Token {
	return c.mqtt.Publish(topic, PublishQoS, false, msg)
}

func (c *DefaultClient) subscribe(topic string, handler MQTT.MessageHandler) Token {
	c.subscriptions[topic] = handler
	return c.mqtt.Subscribe(topic, SubscribeQoS, handler)
}

func (c *DefaultClient) unsubscribe(topic string) Token {
	delete(c.subscriptions, topic)
	return c.mqtt.Unsubscribe(topic)
}

// Disconnect from the MQTT broker
func (c *DefaultClient) Disconnect() {
	if !c.mqtt.IsConnected() {
		return
	}
	c.ctx.Debug("Disconnecting from MQTT")
	c.mqtt.Disconnect(25)
}

// IsConnected returns true if there is a connection to the MQTT broker
func (c *DefaultClient) IsConnected() bool {
	return c.mqtt.IsConnected()
}
