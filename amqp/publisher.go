// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"fmt"
	"io"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	AMQP "github.com/streadway/amqp"
)

// Publisher holds an AMQP channel and can publish on uplink from devices
type Publisher interface {
	Open() error
	io.Closer
	PublishUplink(payload types.UplinkMessage) error
}

// DefaultPublisher is the default AMQP publisher
type DefaultPublisher struct {
	ctx      Logger
	conn     *AMQP.Connection
	channel  *AMQP.Channel
	exchange string
}

// Open opens an AMQP channel for publishing
func (p *DefaultPublisher) Open() error {
	channel, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("Could not open AMQP channel (%s).", err)
	}

	if err = channel.ExchangeDeclare(p.exchange, "topic", true, false, false, false, nil); err != nil {
		channel.Close()
		return fmt.Errorf("Could not declare AMQP exchange (%s).", err)
	}

	p.channel = channel
	return nil
}

// Close closes the AMQP channel
func (p *DefaultPublisher) Close() error {
	if p.channel == nil {
		return nil
	}
	return p.channel.Close()
}

func (p *DefaultPublisher) do(action func() error) error {
	err := action()
	if err == nil {
		return nil
	}
	if err == AMQP.ErrClosed {
		err = p.Open()
		if err != nil {
			return err
		}
		err = action()
	}
	return err
}

func (p *DefaultPublisher) publish(key string, msg []byte, timestamp time.Time) error {
	return p.do(func() error {
		return p.channel.Publish(p.exchange, key, false, false, AMQP.Publishing{
			ContentType:  "application/json",
			DeliveryMode: AMQP.Persistent,
			Timestamp:    timestamp,
			Body:         msg,
		})
	})
}
