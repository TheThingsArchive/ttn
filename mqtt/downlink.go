// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/json"
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/types"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// DownlinkHandler is called for downlink messages
type DownlinkHandler func(client Client, appID string, devID string, req types.DownlinkMessage)

// PublishDownlink publishes a downlink message
func (c *DefaultClient) PublishDownlink(dataDown types.DownlinkMessage) Token {
	topic := DeviceTopic{dataDown.AppID, dataDown.DevID, DeviceDownlink, ""}
	dataDown.AppID = ""
	dataDown.DevID = ""
	msg, err := json.Marshal(dataDown)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload: %s", err)}
	}
	return c.publish(topic.String(), msg)
}

// SubscribeDeviceDownlink subscribes to all downlink messages for the given application and device
func (c *DefaultClient) SubscribeDeviceDownlink(appID string, devID string, handler DownlinkHandler) Token {
	topic := DeviceTopic{appID, devID, DeviceDownlink, ""}
	return c.subscribe(topic.String(), func(mqtt MQTT.Client, msg MQTT.Message) {
		// Determine the actual topic
		topic, err := ParseDeviceTopic(msg.Topic())
		if err != nil {
			c.ctx.Warnf("mqtt: received message on invalid downlink topic: %s", msg.Topic())
			return
		}

		// Unmarshal the payload
		dataDown := &types.DownlinkMessage{}
		err = json.Unmarshal(msg.Payload(), dataDown)
		if err != nil {
			c.ctx.Warnf("mqtt: could not unmarshal downlink: %s", err)
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
	topic := DeviceTopic{appID, devID, DeviceDownlink, ""}
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
