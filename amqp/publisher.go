// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	AMQP "github.com/streadway/amqp"
)

// Publisher represents a publisher for uplink messages
type Publisher interface {
	ChannelUser
	PublishUplink(dataUp types.UplinkMessage) error
}

// DefaultPublisher represents the default AMQP publisher
type DefaultPublisher struct {
	ctx          Logger
	client       *DefaultClient
	channel      *AMQP.Channel
	exchange     string
	exchangeType string
}

// NewTopicPublisher returns a new topic publisher on the specified exchange
func (c *DefaultClient) NewTopicPublisher(exchange string) Publisher {
	return &DefaultPublisher{
		ctx:          c.ctx,
		client:       c,
		exchange:     exchange,
		exchangeType: "topic",
	}
}

// Open opens a new channel and declares the exchange
func (p *DefaultPublisher) Open() error {
	channel, err := p.client.openChannel(p)
	if err != nil {
		return fmt.Errorf("Could not open AMQP channel (%s)", err)
	}

	if err := channel.ExchangeDeclare(p.exchange, p.exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("Could not declare AMQP exchange (%s)", err)
	}

	p.ctx.Debugf("Opened channel for exchange %s", p.exchange)

	p.channel = channel
	return nil
}

// Close closes the channel
func (p *DefaultPublisher) Close() error {
	return p.client.closeChannel(p)
}

func (p *DefaultPublisher) publish(key string, msg []byte, timestamp time.Time) error {
	return p.channel.Publish(p.exchange, key, false, false, AMQP.Publishing{
		ContentType:  "application/json",
		DeliveryMode: AMQP.Persistent,
		Timestamp:    timestamp,
		Body:         msg,
	})
}
