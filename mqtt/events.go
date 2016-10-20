// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"encoding/json"
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// AppEventHandler is called for events
type AppEventHandler func(client Client, appID string, eventType string, payload []byte)

// DeviceEventHandler is called for events
type DeviceEventHandler func(client Client, appID string, devID string, eventType string, payload []byte)

// PublishAppEvent publishes an event to the topic for application events of the given type
// it will marshal the payload to json
func (c *DefaultClient) PublishAppEvent(appID string, eventType string, payload interface{}) Token {
	topic := ApplicationTopic{appID, AppEvents, eventType}
	msg, err := json.Marshal(payload)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.publish(topic.String(), msg)
}

// PublishDeviceEvent publishes an event to the topic for device events of the given type
// it will marshal the payload to json
func (c *DefaultClient) PublishDeviceEvent(appID string, devID string, eventType string, payload interface{}) Token {
	topic := DeviceTopic{appID, devID, DeviceEvents, eventType}
	msg, err := json.Marshal(payload)
	if err != nil {
		return &simpleToken{fmt.Errorf("Unable to marshal the message payload")}
	}
	return c.publish(topic.String(), msg)
}

// SubscribeAppEvents subscribes to events of the given type for the given application. In order to subscribe to
// application events from all applications the user has access to, pass an empty string as appID.
func (c *DefaultClient) SubscribeAppEvents(appID string, eventType string, handler AppEventHandler) Token {
	topic := ApplicationTopic{appID, AppEvents, eventType}
	return c.subscribe(topic.String(), func(mqtt MQTT.Client, msg MQTT.Message) {
		topic, err := ParseApplicationTopic(msg.Topic())
		if err != nil {
			c.ctx.Warnf("Received message on invalid events topic: %s", msg.Topic())
			return
		}
		handler(c, topic.AppID, topic.Field, msg.Payload())
	})
}

// SubscribeDeviceEvents subscribes to events of the given type for the given device. In order to subscribe to
// events from all devices within an application, pass an empty string as devID. In order to subscribe to all
// events from all devices in all applications the user has access to, pass an empty string as appID.
func (c *DefaultClient) SubscribeDeviceEvents(appID string, devID string, eventType string, handler DeviceEventHandler) Token {
	topic := DeviceTopic{appID, devID, DeviceEvents, eventType}
	return c.subscribe(topic.String(), func(mqtt MQTT.Client, msg MQTT.Message) {
		topic, err := ParseDeviceTopic(msg.Topic())
		if err != nil {
			c.ctx.Warnf("Received message on invalid events topic: %s", msg.Topic())
			return
		}
		handler(c, topic.AppID, topic.DevID, topic.Field, msg.Payload())
	})
}

// UnsubscribeAppEvents unsubscribes from the events that were subscribed to by SubscribeAppEvents
func (c *DefaultClient) UnsubscribeAppEvents(appID string, eventType string) Token {
	topic := ApplicationTopic{appID, AppEvents, eventType}
	return c.unsubscribe(topic.String())
}

// UnsubscribeDeviceEvents unsubscribes from the events that were subscribed to by SubscribeDeviceEvents
func (c *DefaultClient) UnsubscribeDeviceEvents(appID string, devID string, eventType string) Token {
	topic := DeviceTopic{appID, devID, DeviceEvents, eventType}
	return c.unsubscribe(topic.String())
}
