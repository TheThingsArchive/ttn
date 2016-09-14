// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/apex/log"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const QoS = 0x00

// Client connects to the MQTT server and can publish/subscribe on uplink, downlink and activations from devices
type Client interface {
	Connect() error
	Disconnect()

	IsConnected() bool

	// Uplink pub/sub
	PublishUplink(payload UplinkMessage) Token
	PublishUplinkFields(appID string, devID string, fields map[string]interface{}) Token
	SubscribeDeviceUplink(appID string, devID string, handler UplinkHandler) Token
	SubscribeAppUplink(appID string, handler UplinkHandler) Token
	SubscribeUplink(handler UplinkHandler) Token
	UnsubscribeDeviceUplink(appID string, devID string) Token
	UnsubscribeAppUplink(appID string) Token
	UnsubscribeUplink() Token

	// Downlink pub/sub
	PublishDownlink(payload DownlinkMessage) Token
	SubscribeDeviceDownlink(appID string, devID string, handler DownlinkHandler) Token
	SubscribeAppDownlink(appID string, handler DownlinkHandler) Token
	SubscribeDownlink(handler DownlinkHandler) Token
	UnsubscribeDeviceDownlink(appID string, devID string) Token
	UnsubscribeAppDownlink(appID string) Token
	UnsubscribeDownlink() Token

	// Activation pub/sub
	PublishActivation(payload Activation) Token
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

// UplinkHandler is called for uplink messages
type UplinkHandler func(client Client, appID string, devID string, req UplinkMessage)

// DownlinkHandler is called for downlink messages
type DownlinkHandler func(client Client, appID string, devID string, req DownlinkMessage)

// ActivationHandler is called for activations
type ActivationHandler func(client Client, appID string, devID string, req Activation)

// DefaultClient is the default MQTT client for The Things Network
type DefaultClient struct {
	mqtt          MQTT.Client
	ctx           log.Interface
	subscriptions map[string]MQTT.MessageHandler
}

// NewClient creates a new DefaultClient
func NewClient(ctx log.Interface, id, username, password string, brokers ...string) Client {
	mqttOpts := MQTT.NewClientOptions()

	for _, broker := range brokers {
		mqttOpts.AddBroker(broker)
	}

	mqttOpts.SetClientID(fmt.Sprintf("%s-%s", id, random.String(16)))
	mqttOpts.SetUsername(username)
	mqttOpts.SetPassword(password)

	// TODO: Some tuning of these values probably won't hurt:
	mqttOpts.SetKeepAlive(30 * time.Second)
	mqttOpts.SetPingTimeout(10 * time.Second)

	mqttOpts.SetCleanSession(true)

	mqttOpts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		ctx.WithField("message", msg).Warn("Received unhandled message")
	})

	var reconnecting bool

	mqttOpts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		ctx.WithError(err).Warn("Disconnected, reconnecting...")
		reconnecting = true
	})

	ttnClient := &DefaultClient{
		ctx:           ctx,
		subscriptions: make(map[string]MQTT.MessageHandler),
	}

	mqttOpts.SetOnConnectHandler(func(client MQTT.Client) {
		ctx.Info("Connected to MQTT")
		if reconnecting {
			for topic, handler := range ttnClient.subscriptions {
				ctx.Infof("Re-subscribing to %s", topic)
				ttnClient.subscribe(topic, handler)
			}
			reconnecting = false
		}
	})

	ttnClient.mqtt = MQTT.NewClient(mqttOpts)

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
		c.ctx.WithError(err).Warn("Could not connect to MQTT Broker. Retrying...")
		<-time.After(ConnectRetryDelay)
	}
	if err != nil {
		return fmt.Errorf("Could not connect to MQTT Broker: %s", err)
	}
	return nil
}

func (c *DefaultClient) subscribe(topic string, handler MQTT.MessageHandler) Token {
	c.subscriptions[topic] = handler
	return c.mqtt.Subscribe(topic, QoS, handler)
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

// PublishUplink publishes an uplink message to the MQTT broker
func (c *DefaultClient) PublishUplink(dataUp UplinkMessage) Token {
	topic := DeviceTopic{dataUp.AppID, dataUp.DevID, Uplink, ""}
	dataUp.AppID = ""
	dataUp.DevID = ""
	msg, err := json.Marshal(dataUp)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.mqtt.Publish(topic.String(), QoS, false, msg)
}

// PublishUplinkFields publishes uplink fields to MQTT
func (c *DefaultClient) PublishUplinkFields(appID string, devID string, fields map[string]interface{}) Token {
	flattenedFields := make(map[string]interface{})
	flatten("", "/", fields, flattenedFields)
	tokens := make([]Token, 0, len(flattenedFields))
	for field, value := range flattenedFields {
		topic := DeviceTopic{appID, devID, Uplink, field}
		pld, _ := json.Marshal(value)
		token := c.mqtt.Publish(topic.String(), QoS, false, pld)
		tokens = append(tokens, token)
	}
	t := newToken()
	go func() {
		for _, token := range tokens {
			token.Wait()
			if token.Error() != nil {
				fmt.Println(token.Error())
				t.err = token.Error()
			}
		}
		t.flowComplete()
	}()
	return t
}

func flatten(prefix, sep string, in, out map[string]interface{}) {
	for k, v := range in {
		key := prefix + sep + k
		if prefix == "" {
			key = k
		}
		out[key] = v
		if next, ok := v.(map[string]interface{}); ok {
			flatten(key, sep, next, out)
		}
	}
}

// SubscribeDeviceUplink subscribes to all uplink messages for the given application and device
func (c *DefaultClient) SubscribeDeviceUplink(appID string, devID string, handler UplinkHandler) Token {
	topic := DeviceTopic{appID, devID, Uplink, ""}
	return c.subscribe(topic.String(), func(mqtt MQTT.Client, msg MQTT.Message) {
		// Determine the actual topic
		topic, err := ParseDeviceTopic(msg.Topic())
		if err != nil {
			c.ctx.WithField("topic", msg.Topic()).WithError(err).Warn("Received message on invalid uplink topic")
			return
		}

		// Unmarshal the payload
		dataUp := &UplinkMessage{}
		err = json.Unmarshal(msg.Payload(), dataUp)
		dataUp.AppID = topic.AppID
		dataUp.DevID = topic.DevID

		if err != nil {
			c.ctx.WithError(err).Warn("Could not unmarshal uplink")
			return
		}

		// Call the uplink handler
		handler(c, topic.AppID, topic.DevID, *dataUp)
	})
}

// SubscribeAppUplink subscribes to all uplink messages for the given application
func (c *DefaultClient) SubscribeAppUplink(appID string, handler UplinkHandler) Token {
	return c.SubscribeDeviceUplink(appID, "", handler)
}

// SubscribeUplink subscribes to all uplink messages that the current user has access to
func (c *DefaultClient) SubscribeUplink(handler UplinkHandler) Token {
	return c.SubscribeDeviceUplink("", "", handler)
}

// UnsubscribeDeviceUplink unsubscribes from the uplink messages for the given application and device
func (c *DefaultClient) UnsubscribeDeviceUplink(appID string, devID string) Token {
	topic := DeviceTopic{appID, devID, Uplink, ""}
	return c.unsubscribe(topic.String())
}

// UnsubscribeAppUplink unsubscribes from the uplink messages for the given application
func (c *DefaultClient) UnsubscribeAppUplink(appID string) Token {
	return c.UnsubscribeDeviceUplink(appID, "")
}

// UnsubscribeUplink unsubscribes from the uplink messages that the current user has access to
func (c *DefaultClient) UnsubscribeUplink() Token {
	return c.UnsubscribeDeviceUplink("", "")
}

// PublishDownlink publishes a downlink message
func (c *DefaultClient) PublishDownlink(dataDown DownlinkMessage) Token {
	topic := DeviceTopic{dataDown.AppID, dataDown.DevID, Downlink, ""}
	dataDown.AppID = ""
	dataDown.DevID = ""
	msg, err := json.Marshal(dataDown)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.mqtt.Publish(topic.String(), QoS, false, msg)
}

// SubscribeDeviceDownlink subscribes to all downlink messages for the given application and device
func (c *DefaultClient) SubscribeDeviceDownlink(appID string, devID string, handler DownlinkHandler) Token {
	topic := DeviceTopic{appID, devID, Downlink, ""}
	return c.subscribe(topic.String(), func(mqtt MQTT.Client, msg MQTT.Message) {
		// Determine the actual topic
		topic, err := ParseDeviceTopic(msg.Topic())
		if err != nil {
			c.ctx.WithField("topic", msg.Topic()).WithError(err).Warn("Received message on invalid Downlink topic")
			return
		}

		// Unmarshal the payload
		dataDown := &DownlinkMessage{}
		err = json.Unmarshal(msg.Payload(), dataDown)
		if err != nil {
			c.ctx.WithError(err).Warn("Could not unmarshal Downlink")
			return
		}
		dataDown.AppID = topic.AppID
		dataDown.DevID = topic.DevID

		// Call the Downlink handler
		handler(c, topic.AppID, topic.DevID, *dataDown)
	})
}

// SubscribeAppDownlink subscribes to all downlink messages for the given application
func (c *DefaultClient) SubscribeAppDownlink(appID string, handler DownlinkHandler) Token {
	return c.SubscribeDeviceDownlink(appID, "", handler)
}

// SubscribeDownlink subscribes to all downlink messages that the current user has access to
func (c *DefaultClient) SubscribeDownlink(handler DownlinkHandler) Token {
	return c.SubscribeDeviceDownlink("", "", handler)
}

// UnsubscribeDeviceDownlink unsubscribes from the downlink messages for the given application and device
func (c *DefaultClient) UnsubscribeDeviceDownlink(appID string, devID string) Token {
	topic := DeviceTopic{appID, devID, Downlink, ""}
	return c.unsubscribe(topic.String())
}

// UnsubscribeAppDownlink unsubscribes from the downlink messages for the given application
func (c *DefaultClient) UnsubscribeAppDownlink(appID string) Token {
	return c.UnsubscribeDeviceDownlink(appID, "")
}

// UnsubscribeDownlink unsubscribes from the downlink messages that the current user has access to
func (c *DefaultClient) UnsubscribeDownlink() Token {
	return c.UnsubscribeDeviceDownlink("", "")
}

// PublishActivation publishes an activation
func (c *DefaultClient) PublishActivation(activation Activation) Token {
	topic := DeviceTopic{activation.AppID, activation.DevID, Activations, ""}
	activation.AppID = ""
	activation.DevID = ""
	msg, err := json.Marshal(activation)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.mqtt.Publish(topic.String(), QoS, false, msg)
}

// SubscribeDeviceActivations subscribes to all activations for the given application and device
func (c *DefaultClient) SubscribeDeviceActivations(appID string, devID string, handler ActivationHandler) Token {
	topic := DeviceTopic{appID, devID, Activations, ""}
	return c.subscribe(topic.String(), func(mqtt MQTT.Client, msg MQTT.Message) {
		// Determine the actual topic
		topic, err := ParseDeviceTopic(msg.Topic())
		if err != nil {
			c.ctx.WithField("topic", msg.Topic()).WithError(err).Warn("Received message on invalid Activations topic")
			return
		}

		// Unmarshal the payload
		activation := &Activation{}
		err = json.Unmarshal(msg.Payload(), activation)
		if err != nil {
			c.ctx.WithError(err).Warn("Could not unmarshal Activation")
			return
		}
		activation.AppID = topic.AppID
		activation.DevID = topic.DevID

		// Call the Activation handler
		handler(c, topic.AppID, topic.DevID, *activation)
	})
}

// SubscribeAppActivations subscribes to all activations for the given application
func (c *DefaultClient) SubscribeAppActivations(appID string, handler ActivationHandler) Token {
	return c.SubscribeDeviceActivations(appID, "", handler)
}

// SubscribeActivations subscribes to all activations that the current user has access to
func (c *DefaultClient) SubscribeActivations(handler ActivationHandler) Token {
	return c.SubscribeDeviceActivations("", "", handler)
}

// UnsubscribeDeviceActivations unsubscribes from the activations for the given application and device
func (c *DefaultClient) UnsubscribeDeviceActivations(appID string, devID string) Token {
	topic := DeviceTopic{appID, devID, Activations, ""}
	return c.unsubscribe(topic.String())
}

// UnsubscribeAppActivations unsubscribes from the activations for the given application
func (c *DefaultClient) UnsubscribeAppActivations(appID string) Token {
	return c.UnsubscribeDeviceActivations(appID, "")
}

// UnsubscribeActivations unsubscribes from the activations that the current user has access to
func (c *DefaultClient) UnsubscribeActivations() Token {
	return c.UnsubscribeDeviceActivations("", "")
}
