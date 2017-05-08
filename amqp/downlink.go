// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// DownlinkHandler is called for downlink messages
type DownlinkHandler func(subscriber Subscriber, appID string, devID string, req types.DownlinkMessage)

// PublishDownlink publishes a downlink message to the AMQP broker
func (c *DefaultPublisher) PublishDownlink(dataDown types.DownlinkMessage) error {
	key := DeviceKey{dataDown.AppID, dataDown.DevID, DeviceDownlink, ""}
	msg, err := json.Marshal(dataDown)
	if err != nil {
		return fmt.Errorf("Unable to marshal the message payload: %s", err)
	}
	return c.publish(key.String(), msg, time.Now())
}

// SubscribeDeviceDownlink subscribes to all downlink messages for the given application and device
func (s *DefaultSubscriber) SubscribeDeviceDownlink(appID, devID string, handler DownlinkHandler) error {
	key := DeviceKey{appID, devID, DeviceDownlink, ""}
	messages, err := s.subscribe(key.String())
	if err != nil {
		return err
	}

	go func() {
		for delivery := range messages {
			dataDown := &types.DownlinkMessage{}
			err := json.Unmarshal(delivery.Body, dataDown)
			if err != nil {
				s.ctx.Warnf("Could not unmarshal downlink %v (%s)", delivery, err)
				continue
			}
			handler(s, dataDown.AppID, dataDown.DevID, *dataDown)
			delivery.Ack(false)
		}
	}()

	return nil
}

// SubscribeAppDownlink subscribes to all downlink messages for the given application
func (s *DefaultSubscriber) SubscribeAppDownlink(appID string, handler DownlinkHandler) error {
	return s.SubscribeDeviceDownlink(appID, "", handler)
}

// SubscribeDownlink subscribes to all downlink messages that the current user has access to
func (s *DefaultSubscriber) SubscribeDownlink(handler DownlinkHandler) error {
	return s.SubscribeDeviceDownlink("", "", handler)
}
