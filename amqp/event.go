// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"encoding/json"
	"fmt"
	"github.com/TheThingsNetwork/ttn/core/types"
	"time"
)

// AppEventHandler is called for events
type AppEventHandler func(sub Subscriber, appID string, eventType types.EventType, payload []byte)

// DeviceEventHandler is called for events
type DeviceEventHandler func(sub Subscriber, appID string, devID string, eventType types.EventType, payload []byte)

// PublishAppEvent publishes an event to the topic for application events of the given type
// it will marshal the payload to json
func (c *DefaultPublisher) PublishAppEvent(appID string, eventType types.EventType, payload interface{}) error {
	key := ApplicationKey{appID, AppEvents, string(eventType)}
	msg, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Unable to marshal the message payload")
	}
	c.publish(key.String(), msg, time.Now())
	return nil
}

// PublishDeviceEvent publishes an event to the topic for device events of the given type
// it will marshal the payload to json
func (c *DefaultPublisher) PublishDeviceEvent(appID string, devID string, eventType types.EventType, payload interface{}) error {
	key := DeviceKey{appID, devID, DeviceEvents, string(eventType)}
	msg, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Unable to marshal the message payload")
	}
	c.publish(key.String(), msg, time.Now())
	return nil
}

// SubscribeAppEvents subscribes to events of the given type for the given application. In order to subscribe to
// application events from all applications the user has access to, pass an empty string as appID.
func (s *DefaultSubscriber) SubscribeAppEvents(appID string, eventType types.EventType, handler AppEventHandler) error {
	key := ApplicationKey{appID, AppEvents, string(eventType)}
	deliveries, err := s.subscribe(key.String())
	if err != nil {
		return err
	}
	go func() {
		for letter := range deliveries {
			handler(s, appID, types.EventType(eventType), letter.Body)
		}
	}()
	return nil
}

// SubscribeDeviceEvents subscribes to events of the given type for the given device. In order to subscribe to
// events from all devices within an application, pass an empty string as devID. In order to subscribe to all
// events from all devices in all applications the user has access to, pass an empty string as appID.
func (s *DefaultSubscriber) SubscribeDeviceEvents(appID string, devID string, eventType types.EventType, handler DeviceEventHandler) error {
	key := DeviceKey{appID, devID, DeviceEvents, string(eventType)}
	deliveries, err := s.subscribe(key.String())
	if err != nil {
		return err
	}
	go func() {
		for letter := range deliveries {
			handler(s, appID, devID, types.EventType(eventType), letter.Body)
		}
	}()
	return nil
}
