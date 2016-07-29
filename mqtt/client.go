// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/json"
	"fmt"
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

type Token interface {
	Wait() bool
	WaitTimeout(time.Duration) bool
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

type UplinkHandler func(client Client, appID string, devID string, req UplinkMessage)
type DownlinkHandler func(client Client, appID string, devID string, req DownlinkMessage)
type ActivationHandler func(client Client, appID string, devID string, req Activation)

type defaultClient struct {
	mqtt          MQTT.Client
	ctx           log.Interface
	subscriptions map[string]MQTT.MessageHandler
}

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

	ttnClient := &defaultClient{
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

func (c *defaultClient) Connect() error {
	if c.mqtt.IsConnected() {
		return nil
	}
	var err error
	for retries := 0; retries < ConnectRetries; retries++ {
		c.ctx.Debug("Connecting to MQTT...")
		token := c.mqtt.Connect()
		token.Wait()
		err = token.Error()
		if err == nil {
			break
		}
		<-time.After(ConnectRetryDelay)
	}
	if err != nil {
		return fmt.Errorf("Could not connect: %s", err)
	}
	return nil
}

func (c *defaultClient) subscribe(topic string, handler MQTT.MessageHandler) Token {
	c.subscriptions[topic] = handler
	return c.mqtt.Subscribe(topic, QoS, handler)
}

func (c *defaultClient) unsubscribe(topic string) Token {
	delete(c.subscriptions, topic)
	return c.mqtt.Unsubscribe(topic)
}

func (c *defaultClient) Disconnect() {
	if !c.mqtt.IsConnected() {
		return
	}
	c.ctx.Debug("Disconnecting from MQTT")
	c.mqtt.Disconnect(25)
}

func (c *defaultClient) IsConnected() bool {
	return c.mqtt.IsConnected()
}

func (c *defaultClient) PublishUplink(dataUp UplinkMessage) Token {
	topic := DeviceTopic{dataUp.AppID, dataUp.DevID, Uplink}
	msg, err := json.Marshal(dataUp)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.mqtt.Publish(topic.String(), QoS, false, msg)
}

func (c *defaultClient) SubscribeDeviceUplink(appID string, devID string, handler UplinkHandler) Token {
	topic := DeviceTopic{appID, devID, Uplink}
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

		if err != nil {
			c.ctx.WithError(err).Warn("Could not unmarshal uplink")
			return
		}

		// Call the uplink handler
		handler(c, topic.AppID, topic.DevID, *dataUp)
	})
}

func (c *defaultClient) SubscribeAppUplink(appID string, handler UplinkHandler) Token {
	return c.SubscribeDeviceUplink(appID, "", handler)
}

func (c *defaultClient) SubscribeUplink(handler UplinkHandler) Token {
	return c.SubscribeDeviceUplink("", "", handler)
}

func (c *defaultClient) UnsubscribeDeviceUplink(appID string, devID string) Token {
	topic := DeviceTopic{appID, devID, Uplink}
	return c.unsubscribe(topic.String())
}
func (c *defaultClient) UnsubscribeAppUplink(appID string) Token {
	return c.UnsubscribeDeviceUplink(appID, "")
}
func (c *defaultClient) UnsubscribeUplink() Token {
	return c.UnsubscribeDeviceUplink("", "")
}

func (c *defaultClient) PublishDownlink(dataDown DownlinkMessage) Token {
	topic := DeviceTopic{dataDown.AppID, dataDown.DevID, Downlink}
	msg, err := json.Marshal(dataDown)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.mqtt.Publish(topic.String(), QoS, false, msg)
}

func (c *defaultClient) SubscribeDeviceDownlink(appID string, devID string, handler DownlinkHandler) Token {
	topic := DeviceTopic{appID, devID, Downlink}
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

		// Call the Downlink handler
		handler(c, topic.AppID, topic.DevID, *dataDown)
	})
}

func (c *defaultClient) SubscribeAppDownlink(appID string, handler DownlinkHandler) Token {
	return c.SubscribeDeviceDownlink(appID, "", handler)
}

func (c *defaultClient) SubscribeDownlink(handler DownlinkHandler) Token {
	return c.SubscribeDeviceDownlink("", "", handler)
}

func (c *defaultClient) UnsubscribeDeviceDownlink(appID string, devID string) Token {
	topic := DeviceTopic{appID, devID, Downlink}
	return c.unsubscribe(topic.String())
}
func (c *defaultClient) UnsubscribeAppDownlink(appID string) Token {
	return c.UnsubscribeDeviceDownlink(appID, "")
}
func (c *defaultClient) UnsubscribeDownlink() Token {
	return c.UnsubscribeDeviceDownlink("", "")
}

func (c *defaultClient) PublishActivation(activation Activation) Token {
	topic := DeviceTopic{activation.AppID, activation.DevID, Activations}
	msg, err := json.Marshal(activation)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.mqtt.Publish(topic.String(), QoS, false, msg)
}

func (c *defaultClient) SubscribeDeviceActivations(appID string, devID string, handler ActivationHandler) Token {
	topic := DeviceTopic{appID, devID, Activations}
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

		// Call the Activation handler
		handler(c, topic.AppID, topic.DevID, *activation)
	})
}

func (c *defaultClient) SubscribeAppActivations(appID string, handler ActivationHandler) Token {
	return c.SubscribeDeviceActivations(appID, "", handler)
}

func (c *defaultClient) SubscribeActivations(handler ActivationHandler) Token {
	return c.SubscribeDeviceActivations("", "", handler)
}

func (c *defaultClient) UnsubscribeDeviceActivations(appID string, devID string) Token {
	topic := DeviceTopic{appID, devID, Activations}
	return c.unsubscribe(topic.String())
}

func (c *defaultClient) UnsubscribeAppActivations(appID string) Token {
	return c.UnsubscribeDeviceActivations(appID, "")
}

func (c *defaultClient) UnsubscribeActivations() Token {
	return c.UnsubscribeDeviceActivations("", "")
}
