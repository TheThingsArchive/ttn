// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/core/types"
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

	QueueDeclare() (string, error)
	QueueBind(name, key string) error
	QueueUnbind(name, key string) error

	SubscribeDeviceUplink(appID, devID string, handler UplinkHandler) error
	SubscribeAppUplink(appID string, handler UplinkHandler) error
	SubscribeUplink(handler UplinkHandler) error
	ConsumeMessages(queue string, uplinkHandler UplinkHandler, eventHandler DeviceEventHandler) error

	SubscribeDeviceDownlink(appID, devID string, handler DownlinkHandler) error
	SubscribeAppDownlink(appID string, handler DownlinkHandler) error
	SubscribeDownlink(handler DownlinkHandler) error

	SubscribeDeviceEvents(appID string, devID string, eventType types.EventType, handler DeviceEventHandler) error
	SubscribeAppEvents(appID string, eventType types.EventType, handler AppEventHandler) error
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

// QueueDeclare declares the queue on the AMQP broker
func (s *DefaultSubscriber) QueueDeclare() (string, error) {
	queue, err := s.channel.QueueDeclare(s.name, s.durable, s.autoDelete, false, false, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to declare queue '%s' (%s)", s.name, err)
	}
	return queue.Name, nil
}

// QueueBind binds the routing key to the specified queue
func (s *DefaultSubscriber) QueueBind(name, key string) error {
	err := s.channel.QueueBind(name, key, s.exchange, false, nil)
	if err != nil {
		return fmt.Errorf("Failed to bind queue %s with key %s on exchange '%s' (%s)", name, key, s.exchange, err)
	}
	return nil
}

// QueueUnbind unbinds the routing key from the specified queue
func (s *DefaultSubscriber) QueueUnbind(name, key string) error {
	err := s.channel.QueueUnbind(name, key, s.exchange, nil)
	if err != nil {
		return fmt.Errorf("Failed to unbind queue %s with key %s on exchange '%s' (%s)", name, key, s.exchange, err)
	}
	return nil
}

type consumer struct {
	queue      string
	deliveries chan AMQP.Delivery
}

func (c *consumer) use(channel *AMQP.Channel) error {
	err := channel.Qos(PrefetchCount, PrefetchSize, false)
	if err != nil {
		return fmt.Errorf("Failed to set channel QoS (%s)", err)
	}
	deliveries, err := channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for delivery := range deliveries {
			c.deliveries <- delivery
		}
	}()
	return nil
}

func (c *consumer) close() {
	if c.deliveries != nil {
		close(c.deliveries)
		c.deliveries = nil
	}
}

func (s *DefaultSubscriber) consume(queue string) (<-chan AMQP.Delivery, error) {
	deliveries := make(chan AMQP.Delivery)
	c := &consumer{
		queue:      queue,
		deliveries: deliveries,
	}
	if err := c.use(s.channel); err != nil {
		return nil, err
	}
	s.addUser(c)
	return deliveries, nil
}

func (s *DefaultSubscriber) subscribe(key string) (<-chan AMQP.Delivery, error) {
	queue, err := s.QueueDeclare()
	if err != nil {
		return nil, err
	}
	err = s.QueueBind(queue, key)
	if err != nil {
		return nil, err
	}
	return s.consume(queue)
}

// ConsumeMessages consumes all messages in a specific queue and try to handle them as uplinks or events depending on the
// routing key.
func (s *DefaultSubscriber) ConsumeMessages(queue string, uplinkHandler UplinkHandler, eventHandler DeviceEventHandler) error {
	messages, err := s.consume(s.name)
	if err != nil {
		return err
	}
	go s.handleMessages(messages, uplinkHandler, eventHandler)
	return nil
}

// handleMessages consume AMQP.Delivery on a channel and call the uplinkHandler and eventHandler respectively for
// device uplinks and device events. Deliveries won't be acknowledged if the routing key contain unknown fields or their
// body cannot be parsed.
func (s *DefaultSubscriber) handleMessages(messages <-chan AMQP.Delivery, uplinkHandler UplinkHandler, eventHandler DeviceEventHandler) {
	for delivery := range messages {
		pins := strings.Split(delivery.RoutingKey, ".")
		if len(pins) < 4 {
			s.ctx.Warnf("Routing key is missing fields, cannot handle delivery (%s)", delivery.RoutingKey)
			continue
		}
		if pins[3] == string(DeviceEvents) {
			event := &types.DeviceEvent{}
			if err := json.Unmarshal(delivery.Body, event); err != nil {
				s.ctx.Warnf("Could not unmarshal device event (%s)", err)
				continue
			}
			eventHandler(s, event.AppID, event.DevID, *event)
		} else if pins[3] == string(DeviceUplink) {
			dataUp := &types.UplinkMessage{}
			if err := json.Unmarshal(delivery.Body, dataUp); err != nil {
				s.ctx.Warnf("Could not unmarshal device uplink (%s)", err)
				continue
			}
			uplinkHandler(s, dataUp.AppID, dataUp.DevID, *dataUp)
		} else {
			s.ctx.Warnf("Unknown routing key field, cannot handle delivery (%s)", delivery.RoutingKey)
			continue
		}
		if err := delivery.Ack(false); err != nil {
			s.ctx.Warnf("Could not acknowledge message (%s)", err)
		}
	}
}
