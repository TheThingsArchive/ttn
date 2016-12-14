// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"fmt"

	AMQP "github.com/streadway/amqp"
)

var (
	// PrefetchCount represents the number of messages to prefetch before the AMQP server requires acknowledgment
	PrefetchCount = 3
	// PrefetchSize represents the number of bytes to prefetch before the AMQP server requires acknowledgment
	PrefetchSize = 0
)

// Subscriber represents a subscriber for uplink messages
type Subscriber interface {
	ChannelClient

	SubscribeDeviceUplink(appID, devID string, handler UplinkHandler) error
	SubscribeAppUplink(appID string, handler UplinkHandler) error
	SubscribeUplink(handler UplinkHandler) error

	SubscribeDeviceDownlink(appID, devID string, handler DownlinkHandler) error
	SubscribeAppDownlink(appID string, handler DownlinkHandler) error
	SubscribeDownlink(handler DownlinkHandler) error
}

// DefaultSubscriber represents the default AMQP subscriber
type DefaultSubscriber struct {
	DefaultChannelClient

	name       string
	durable    bool
	autoDelete bool
}

// NewSubscriber returns a new topic subscriber on the specified exchange
func (c *DefaultClient) NewSubscriber(exchange, name string, durable, autoDelete bool) Subscriber {
	return &DefaultSubscriber{
		DefaultChannelClient: DefaultChannelClient{
			ctx:      c.ctx,
			client:   c,
			exchange: exchange,
			name:     "Subscriber",
		},
		name:       name,
		durable:    durable,
		autoDelete: autoDelete,
	}
}

func (s *DefaultSubscriber) subscribe(key string) (<-chan AMQP.Delivery, error) {
	queue, err := s.channel.QueueDeclare(s.name, s.durable, s.autoDelete, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare queue '%s' (%s)", s.name, err)
	}

	err = s.channel.QueueBind(queue.Name, key, s.exchange, false, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to bind queue %s with key %s on exchange '%s' (%s)", queue.Name, key, s.exchange, err)
	}

	err = s.channel.Qos(PrefetchCount, PrefetchSize, false)
	if err != nil {
		return nil, fmt.Errorf("Failed to set channel QoS (%s)", err)
	}

	return s.channel.Consume(queue.Name, "", false, false, false, false, nil)
}
