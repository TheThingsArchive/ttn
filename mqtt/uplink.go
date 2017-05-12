// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/json"
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/types"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// UplinkHandler is called for uplink messages
type UplinkHandler func(client Client, appID string, devID string, req types.UplinkMessage)

// PublishUplink publishes an uplink message to the MQTT broker
func (c *DefaultClient) PublishUplink(dataUp types.UplinkMessage) Token {
	topic := DeviceTopic{dataUp.AppID, dataUp.DevID, DeviceUplink, ""}
	msg, err := json.Marshal(dataUp)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload: %s", err)}
	}
	return c.publish(topic.String(), msg)
}

// PublishUplinkFields publishes uplink fields to MQTT
func (c *DefaultClient) PublishUplinkFields(appID string, devID string, fields map[string]interface{}) Token {
	flattenedFields := make(map[string]interface{})
	flatten("", "/", fields, flattenedFields)
	tokens := make([]Token, 0, len(flattenedFields))
	for field, value := range flattenedFields {
		topic := DeviceTopic{appID, devID, DeviceUplink, field}
		pld, _ := json.Marshal(value)
		token := c.publish(topic.String(), pld)
		tokens = append(tokens, token)
	}
	t := newToken()
	go func() {
		for _, token := range tokens {
			token.Wait()
			if token.Error() != nil {
				c.ctx.Warnf("mqtt: error publishing uplink fields: %s", token.Error())
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
	topic := DeviceTopic{appID, devID, DeviceUplink, ""}
	return c.subscribe(topic.String(), func(mqtt MQTT.Client, msg MQTT.Message) {
		// Determine the actual topic
		topic, err := ParseDeviceTopic(msg.Topic())
		if err != nil {
			c.ctx.Warnf("mqtt: received message on invalid uplink topic: %s", msg.Topic())
			return
		}

		// Unmarshal the payload
		dataUp := &types.UplinkMessage{}
		err = json.Unmarshal(msg.Payload(), dataUp)
		dataUp.AppID = topic.AppID
		dataUp.DevID = topic.DevID

		if err != nil {
			c.ctx.Warnf("mqtt: could not unmarshal uplink: %s", err)
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
	topic := DeviceTopic{appID, devID, DeviceUplink, ""}
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
