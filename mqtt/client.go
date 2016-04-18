// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/apex/log"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const QoS = 0x02

// Client connects to the MQTT server and can publish/subscribe on uplink, downlink and activations from devices
type Client interface {
	Connect() error
	Disconnect()

	IsConnected() bool

	// Uplink pub/sub
	PublishUplink(appEUI []byte, devEUI []byte, payload core.DataUpAppReq) Token
	SubscribeDeviceUplink(appEUI []byte, devEUI []byte, handler UplinkHandler) Token
	SubscribeAppUplink(appEUI []byte, handler UplinkHandler) Token
	SubscribeUplink(handler UplinkHandler) Token

	// Downlink pub/sub
	PublishDownlink(appEUI []byte, devEUI []byte, payload core.DataDownAppReq) Token
	SubscribeDeviceDownlink(appEUI []byte, devEUI []byte, handler DownlinkHandler) Token
	SubscribeAppDownlink(appEUI []byte, handler DownlinkHandler) Token
	SubscribeDownlink(handler DownlinkHandler) Token

	// Activation pub/sub
	PublishActivation(appEUI []byte, devEUI []byte, payload core.OTAAAppReq) Token
	SubscribeDeviceActivations(appEUI []byte, devEUI []byte, handler ActivationHandler) Token
	SubscribeAppActivations(appEUI []byte, handler ActivationHandler) Token
	SubscribeActivations(handler ActivationHandler) Token
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

type UplinkHandler func(client Client, appEUI []byte, devEUI []byte, req core.DataUpAppReq)
type DownlinkHandler func(client Client, appEUI []byte, devEUI []byte, req core.DataDownAppReq)
type ActivationHandler func(client Client, appEUI []byte, devEUI []byte, req core.OTAAAppReq)

type defaultClient struct {
	mqtt MQTT.Client
	ctx  log.Interface
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

	mqttOpts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		if ctx != nil {
			ctx.WithField("message", msg).Debug("Received unhandled message")
		}
	})

	mqttOpts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		if ctx != nil {
			ctx.WithError(err).Debug("Disconnected, reconnecting...")
		}
	})

	mqttOpts.SetOnConnectHandler(func(client MQTT.Client) {
		if ctx != nil {
			ctx.Debug("Connected")
		}
	})

	return &defaultClient{
		mqtt: MQTT.NewClient(mqttOpts),
		ctx:  ctx,
	}
}

func (c *defaultClient) Connect() error {
	if c.mqtt.IsConnected() {
		return nil
	}
	if token := c.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Could not connect: %s", token.Error())
	}
	return nil
}

func (c *defaultClient) Disconnect() {
	if !c.mqtt.IsConnected() {
		return
	}
	c.mqtt.Disconnect(25)
}

func (c *defaultClient) IsConnected() bool {
	return c.mqtt.IsConnected()
}

func (c *defaultClient) PublishUplink(appEUI []byte, devEUI []byte, dataUp core.DataUpAppReq) Token {
	topic := DeviceTopic{appEUI, devEUI, Uplink}
	msg, err := json.Marshal(dataUp)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.mqtt.Publish(topic.String(), QoS, false, msg)
}

func (c *defaultClient) SubscribeDeviceUplink(appEUI []byte, devEUI []byte, handler UplinkHandler) Token {
	topic := DeviceTopic{appEUI, devEUI, Uplink}
	return c.mqtt.Subscribe(topic.String(), QoS, func(mqtt MQTT.Client, msg MQTT.Message) {
		// Determine the actual topic
		topic, err := ParseDeviceTopic(msg.Topic())
		if err != nil {
			if c.ctx != nil {
				c.ctx.WithField("topic", msg.Topic()).WithError(err).Warn("Received message on invalid uplink topic")
			}
			return
		}

		// Unmarshal the payload
		dataUp := &core.DataUpAppReq{}
		err = json.Unmarshal(msg.Payload(), dataUp)

		if err != nil {
			if c.ctx != nil {
				c.ctx.WithError(err).Warn("Could not unmarshal uplink")
			}
			return
		}

		// Call the uplink handler
		handler(c, topic.AppEUI, topic.DevEUI, *dataUp)
	})
}

func (c *defaultClient) SubscribeAppUplink(appEUI []byte, handler UplinkHandler) Token {
	return c.SubscribeDeviceUplink(appEUI, []byte{}, handler)
}

func (c *defaultClient) SubscribeUplink(handler UplinkHandler) Token {
	return c.SubscribeDeviceUplink([]byte{}, []byte{}, handler)
}

func (c *defaultClient) PublishDownlink(appEUI []byte, devEUI []byte, dataDown core.DataDownAppReq) Token {
	topic := DeviceTopic{appEUI, devEUI, Downlink}
	msg, err := json.Marshal(dataDown)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.mqtt.Publish(topic.String(), QoS, false, msg)
}

func (c *defaultClient) SubscribeDeviceDownlink(appEUI []byte, devEUI []byte, handler DownlinkHandler) Token {
	topic := DeviceTopic{appEUI, devEUI, Downlink}
	return c.mqtt.Subscribe(topic.String(), QoS, func(mqtt MQTT.Client, msg MQTT.Message) {
		// Determine the actual topic
		topic, err := ParseDeviceTopic(msg.Topic())
		if err != nil {
			if c.ctx != nil {
				c.ctx.WithField("topic", msg.Topic()).WithError(err).Warn("Received message on invalid Downlink topic")
			}
			return
		}

		// Unmarshal the payload
		dataDown := &core.DataDownAppReq{}
		err = json.Unmarshal(msg.Payload(), dataDown)
		if err != nil {
			if c.ctx != nil {
				c.ctx.WithError(err).Warn("Could not unmarshal Downlink")
			}
			return
		}

		// Call the Downlink handler
		handler(c, topic.AppEUI, topic.DevEUI, *dataDown)
	})
}

func (c *defaultClient) SubscribeAppDownlink(appEUI []byte, handler DownlinkHandler) Token {
	return c.SubscribeDeviceDownlink(appEUI, []byte{}, handler)
}

func (c *defaultClient) SubscribeDownlink(handler DownlinkHandler) Token {
	return c.SubscribeDeviceDownlink([]byte{}, []byte{}, handler)
}

func (c *defaultClient) PublishActivation(appEUI []byte, devEUI []byte, activation core.OTAAAppReq) Token {
	topic := DeviceTopic{appEUI, devEUI, Activations}
	msg, err := json.Marshal(activation)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.mqtt.Publish(topic.String(), QoS, false, msg)
}

func (c *defaultClient) SubscribeDeviceActivations(appEUI []byte, devEUI []byte, handler ActivationHandler) Token {
	topic := DeviceTopic{appEUI, devEUI, Activations}
	return c.mqtt.Subscribe(topic.String(), QoS, func(mqtt MQTT.Client, msg MQTT.Message) {
		// Determine the actual topic
		topic, err := ParseDeviceTopic(msg.Topic())
		if err != nil {
			if c.ctx != nil {
				c.ctx.WithField("topic", msg.Topic()).WithError(err).Warn("Received message on invalid Activations topic")
			}
			return
		}

		// Unmarshal the payload
		activation := &core.OTAAAppReq{}
		err = json.Unmarshal(msg.Payload(), activation)
		if err != nil {
			if c.ctx != nil {
				c.ctx.WithError(err).Warn("Could not unmarshal Activation")
			}
			return
		}

		// Call the Activation handler
		handler(c, topic.AppEUI, topic.DevEUI, *activation)
	})
}

func (c *defaultClient) SubscribeAppActivations(appEUI []byte, handler ActivationHandler) Token {
	return c.SubscribeDeviceActivations(appEUI, []byte{}, handler)
}

func (c *defaultClient) SubscribeActivations(handler ActivationHandler) Token {
	return c.SubscribeDeviceActivations([]byte{}, []byte{}, handler)
}
