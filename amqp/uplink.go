// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"encoding/json"
	"fmt"
	"time"

	AMQP "github.com/streadway/amqp"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// UplinkHandler is called for uplink messages
type UplinkHandler func(subscriber Subscriber, appID string, devID string, req types.UplinkMessage)

// PublishUplink publishes an uplink message to the AMQP broker
func (c *DefaultPublisher) PublishUplink(dataUp types.UplinkMessage) error {
	key := DeviceKey{dataUp.AppID, dataUp.DevID, DeviceUplink, ""}
	msg, err := json.Marshal(dataUp)
	if err != nil {
		return fmt.Errorf("Unable to marshal the message payload: %s", err)
	}
	return c.publish(key.String(), msg, time.Time(dataUp.Metadata.Time))
}

func (s *DefaultSubscriber) handleUplink(messages <-chan AMQP.Delivery, handler UplinkHandler) {
	for delivery := range messages {
		dataUp := &types.UplinkMessage{}
		if err := json.Unmarshal(delivery.Body, dataUp); err != nil {
			s.ctx.Warnf("Could not unmarshal uplink (%s)", err)
			continue
		}
		handler(s, dataUp.AppID, dataUp.DevID, *dataUp)
		if err := delivery.Ack(false); err != nil {
			s.ctx.Warnf("Could not acknowledge message (%s)", err)
		}
	}
}

// SubscribeDeviceUplink subscribes to all uplink messages for the given application and device
func (s *DefaultSubscriber) SubscribeDeviceUplink(appID, devID string, handler UplinkHandler) error {
	key := DeviceKey{appID, devID, DeviceUplink, ""}
	messages, err := s.subscribe(key.String())
	if err != nil {
		return err
	}

	go s.handleUplink(messages, handler)
	return nil
}

// ConsumeUplink consumes uplink messages in a specific queue
func (s *DefaultSubscriber) ConsumeUplink(queue string, handler UplinkHandler) error {
	messages, err := s.consume(s.name)
	if err != nil {
		return err
	}

	go s.handleUplink(messages, handler)
	return nil
}

// SubscribeAppUplink subscribes to all uplink messages for the given application
func (s *DefaultSubscriber) SubscribeAppUplink(appID string, handler UplinkHandler) error {
	return s.SubscribeDeviceUplink(appID, "", handler)
}

// SubscribeUplink subscribes to all uplink messages that the current user has access to
func (s *DefaultSubscriber) SubscribeUplink(handler UplinkHandler) error {
	return s.SubscribeDeviceUplink("", "", handler)
}
