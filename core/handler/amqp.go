// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"github.com/apex/log"

	"github.com/TheThingsNetwork/ttn/amqp"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// AMQPBufferSize indicates the size for uplink channel buffers
var AMQPBufferSize = 10

func (h *handler) HandleAMQP(username, password, host, exchange string) error {
	h.amqpPublisher = amqp.NewPublisher(h.Ctx, username, password, host, exchange)

	err := h.amqpPublisher.Connect()
	if err != nil {
		return err
	}

	h.amqpUp = make(chan *types.UplinkMessage, AMQPBufferSize)

	ctx := h.Ctx.WithField("Protocol", "AMQP")

	go func() {
		for up := range h.amqpUp {
			ctx.WithFields(log.Fields{
				"DevID": up.DevID,
				"AppID": up.AppID,
			}).Debug("Publish Uplink")
			msg := *up
			go func() {
				err := h.amqpPublisher.PublishUplink(msg)
				if err != nil {
					ctx.WithError(err).Warn("Could not publish Uplink")
				}
			}()
		}
	}()

	return nil
}
