// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	AMQP "github.com/streadway/amqp"
)

// Publisher represents a publisher for uplink messages
type Publisher interface {
	ChannelClient

	PublishUplink(dataUp types.UplinkMessage) error
	PublishDownlink(dataDown types.DownlinkMessage) error
	PublishDeviceEvent(appID string, devID string, eventType types.EventType, payload interface{}) error
	PublishAppEvent(appID string, eventType types.EventType, payload interface{}) error
}

// DefaultPublisher represents the default AMQP publisher
type DefaultPublisher struct {
	DefaultChannelClient
}

// NewPublisher returns a new topic publisher on the specified exchange
func (c *DefaultClient) NewPublisher(exchange string) Publisher {
	return &DefaultPublisher{
		DefaultChannelClient: DefaultChannelClient{
			ctx:      c.ctx,
			client:   c,
			exchange: exchange,
			name:     "Publisher",
		},
	}
}

func (p *DefaultPublisher) publish(key string, msg []byte, timestamp time.Time) error {
	return p.channel.Publish(p.exchange, key, false, false, AMQP.Publishing{
		ContentType:  "application/json",
		DeliveryMode: AMQP.Persistent,
		Timestamp:    timestamp,
		Body:         msg,
	})
}
