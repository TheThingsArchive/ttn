// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"sync"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/amqp"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// AMQPBufferSize indicates the size for uplink channel buffers
var AMQPBufferSize = 10

func (h *handler) assertAMQPExchange() error {
	ch, err := h.amqpClient.(*amqp.DefaultClient).GetChannel()
	if err != nil {
		return err
	}
	err = ch.ExchangeDeclarePassive(h.amqpExchange, "topic", true, false, false, false, nil)
	if err != nil {
		h.Ctx.Warnf("Could not assert presence of AMQP Exchange %s, trying to create...", h.amqpExchange)
		ch, err := h.amqpClient.(*amqp.DefaultClient).GetChannel()
		if err != nil {
			return err
		}
		err = ch.ExchangeDeclare(h.amqpExchange, "topic", true, false, false, false, nil)
		if err != nil {
			h.Ctx.Errorf("Could not create AMQP Exchange %s.", h.amqpExchange)
			return err
		}
		h.Ctx.Infof("Created AMQP Exchange %s", h.amqpExchange)
	}
	return nil
}

func (h *handler) HandleAMQP(username, password, host, exchange, downlinkQueue string) error {
	h.amqpClient = amqp.NewClient(h.Ctx, username, password, host)

	err := h.amqpClient.Connect()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			h.amqpClient.Disconnect()
		}
	}()

	if err := h.assertAMQPExchange(); err != nil {
		return err
	}

	h.amqpUp = make(chan *types.UplinkMessage, AMQPBufferSize)
	h.amqpEvent = make(chan *types.DeviceEvent, AMQPBufferSize)

	subscriber := h.amqpClient.NewSubscriber(h.amqpExchange, downlinkQueue, downlinkQueue != "", downlinkQueue == "")
	err = subscriber.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			subscriber.Close()
		}
	}()
	err = subscriber.SubscribeDownlink(func(_ amqp.Subscriber, _, _ string, req types.DownlinkMessage) {
		h.EnqueueDownlink(&req)
	})
	if err != nil {
		return err
	}

	ctx := h.Ctx.WithField("Protocol", "AMQP")
	publisher := h.amqpClient.NewPublisher(h.amqpExchange)
	err = publisher.Open()
	if err != nil {
		ctx.WithError(err).Error("Could not open publisher channel")
		return err
	}

	var pubWait sync.WaitGroup
	pubWait.Add(2)
	defer func() {
		go func() {
			pubWait.Wait()
			publisher.Close()
		}()
	}()

	go func() {
		defer pubWait.Done()
		for up := range h.amqpUp {
			ctx.WithFields(ttnlog.Fields{
				"DevID": up.DevID,
				"AppID": up.AppID,
			}).Debug("Publish Uplink")
			err := publisher.PublishUplink(*up)
			if err != nil {
				ctx.WithError(err).Warn("Could not publish Uplink")
			}
		}
	}()

	go func() {
		defer pubWait.Done()
		for event := range h.amqpEvent {
			ctx.WithFields(ttnlog.Fields{
				"DevID": event.DevID,
				"AppID": event.AppID,
				"Event": event.Event,
			}).Debug("Publish Event")
			if event.DevID == "" {
				publisher.PublishAppEvent(event.AppID, event.Event, event.Data)
			} else {
				publisher.PublishDeviceEvent(event.AppID, event.DevID, event.Event, event.Data)
			}
		}
	}()

	return nil
}
