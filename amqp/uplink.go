// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// UplinkHandler is called for uplink messages
type UplinkHandler func(subscriber Subscriber, appID string, devID string, req types.UplinkMessage)

// PublishUplink publishes an uplink message to the AMQP broker
func (c *DefaultPublisher) PublishUplink(dataUp types.UplinkMessage) error {
	key := DeviceKey{dataUp.AppID, dataUp.DevID, DeviceUplink, ""}
	msg, err := json.Marshal(dataUp)
	if err != nil {
		return fmt.Errorf("Unable to marshal the message payload")
	}
	return c.publish(key.String(), msg, time.Time(dataUp.Metadata.Time))
}

// SubscribeDeviceUplink subscribes to all uplink messages for the given application and device
func (s *DefaultSubscriber) SubscribeDeviceUplink(appID, devID string, handler UplinkHandler) error {
	key := DeviceKey{appID, devID, DeviceUplink, ""}
	messages, err := s.subscribe(key.String())
	if err != nil {
		return err
	}

	go func() {
		for delivery := range messages {
			dataUp := &types.UplinkMessage{}
			err := json.Unmarshal(delivery.Body, dataUp)
			if err != nil {
				s.ctx.Warnf("Could not unmarshal uplink %v (%s)", delivery, err)
				continue
			}
			handler(s, dataUp.AppID, dataUp.DevID, *dataUp)
			delivery.Ack(false)
			break
		}
	}()

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
