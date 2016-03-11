// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"sync"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
)

var timeout = time.Second

// Adapter type materializes an mqtt adapter which implements the basic mqtt protocol
type Adapter struct {
	Client
	ctx           log.Interface
	packets       chan PktReq // Channel used to "transforms" incoming request to something we can handle concurrently
	registrations chan RegReq // Incoming registrations
}

// Handler defines topic-specific handler.
type Handler interface {
	Topic() string
	Handle(client Client, chpkt chan<- PktReq, chreg chan<- RegReq, msg MQTT.Message) error
}

// MsgRes are sent through the response channel of a pktReq or regReq
type MsgRes []byte // The response content.

// PktReq are sent through the packets channel when an incoming request arrives
type PktReq struct {
	Packet []byte      // The actual packet that has been parsed
	Chresp chan MsgRes // A response channel waiting for an success or reject confirmation
}

// RegReq are sent through the registration channel when an incoming registration arrives
type RegReq struct {
	Registration core.Registration
	Chresp       chan MsgRes
}

// Scheme defines all MQTT communication schemes available
type Scheme string

// The following constants are used as scheme identifers
const (
	TCP       Scheme = "tcp"
	TLS       Scheme = "tls"
	WebSocket Scheme = "ws"
)

// NewAdapter constructs and allocates a new mqtt adapter
//
// The client is expected to be already connected to the right broker and ready to be used.
func NewAdapter(client Client, ctx log.Interface) *Adapter {
	adapter := &Adapter{
		Client:        client,
		ctx:           ctx,
		packets:       make(chan PktReq),
		registrations: make(chan RegReq),
	}

	return adapter
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, recipients ...core.Recipient) ([]byte, error) {
	stats.MarkMeter("mqtt_adapter.send")
	stats.UpdateHistogram("mqtt_adapter.send_recipients", int64(len(recipients)))

	// Marshal the packet to raw binary data
	data, err := p.MarshalBinary()
	if err != nil {
		a.ctx.WithError(err).Warn("Invalid Packet")
		return nil, errors.New(errors.Structural, err)
	}

	a.ctx.Debug("Sending Packet")

	// Prepare ground for parrallel mqtt publications
	nb := len(recipients)
	cherr := make(chan error, nb)
	chresp := make(chan []byte, nb)
	wg := sync.WaitGroup{}
	wg.Add(2 * nb)

	for _, r := range recipients {
		// Get the actual recipient
		recipient, ok := r.(Recipient)
		if !ok {
			err := errors.New(errors.Structural, "Unable to interpret recipient as mqttRecipient")
			a.ctx.WithField("recipient", r).Warn(err.Error())
			wg.Done()
			wg.Done()
			cherr <- err
			continue
		}

		// Subscribe to down channel (before publishing anything)
		chdown := make(chan []byte)
		if recipient.TopicDown() != "" {
			token := a.Subscribe(recipient.TopicDown(), 2, func(client Client, msg MQTT.Message) {
				chdown <- msg.Payload()
			})
			if token.Wait() && token.Error() != nil {
				err := errors.New(errors.Operational, "Unable to subscribe to down topic")
				a.ctx.WithField("recipient", recipient).Warn(err.Error())
				wg.Done()
				wg.Done()
				cherr <- err
				close(chdown)
				continue
			}
		}

		// Publish on each topic
		go func(recipient Recipient) {
			defer wg.Done()

			ctx := a.ctx.WithField("topic", recipient.TopicUp())

			// Publish packet
			ctx.WithField("data", data).Debug("Publish data to mqtt")
			token := a.Publish(recipient.TopicUp(), 2, false, data)
			if token.Wait() && token.Error() != nil {
				ctx.WithError(token.Error()).Error("Unable to publish")
				cherr <- errors.New(errors.Operational, token.Error())
				return
			}
		}(recipient)

		// Avoid waiting for response when there's no topic down
		if recipient.TopicDown() == "" {
			a.ctx.WithField("recipient", recipient).Debug("No response expected from mqtt recipient")
			wg.Done()
			continue
		}

		// Pull responses from each down topic, expecting only one
		go func(recipient Recipient, chdown <-chan []byte) {
			defer wg.Done()

			ctx := a.ctx.WithField("topic", recipient.TopicDown())

			ctx.Debug("Wait for mqtt response")
			defer func(ctx log.Interface) {
				if token := a.Unsubscribe(recipient.TopicDown()); token.Wait() && token.Error() != nil {
					ctx.Warn("Unable to unsubscribe topic")
				}
			}(ctx)

			// Forward the downlink response received if any
			select {
			case data, ok := <-chdown:
				if ok {
					chresp <- data
				}
			case <-time.After(timeout): // Timeout
			}
		}(recipient, chdown)
	}

	// Wait for each request to be done
	stats.IncCounter("mqtt_adapter.waiting_for_send")
	wg.Wait()
	stats.DecCounter("mqtt_adapter.waiting_for_send")
	close(cherr)
	close(chresp)

	// Collect errors
	errored := len(cherr)

	// Collect response
	if len(chresp) > 1 {
		return nil, errors.New(errors.Behavioural, "Received too many positive answers")
	}

	if len(chresp) == 0 && errored > 0 {
		return nil, errors.New(errors.Operational, "No positive response from recipients but got unexpected answers")
	}

	if len(chresp) == 0 {
		return nil, nil
	}
	return <-chresp, nil
}

// GetRecipient implements the core.Adapter interface
func (a *Adapter) GetRecipient(raw []byte) (core.Recipient, error) {
	recipient := new(recipient)
	if err := recipient.UnmarshalBinary(raw); err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	return recipient, nil
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() ([]byte, core.AckNacker, error) {
	p := <-a.packets
	return p.Packet, mqttAckNacker{Chresp: p.Chresp}, nil
}

// NextRegistration implements the core.Adapter interface. Not implemented for this adapters.
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	r := <-a.registrations
	return r.Registration, mqttAckNacker{Chresp: r.Chresp}, nil
}

// Bind registers a handler to a specific endpoint
func (a *Adapter) Bind(h Handler) error {
	ctx := a.ctx.WithField("topic", h.Topic())
	ctx.Info("Subscribe new handler")
	token := a.Subscribe(h.Topic(), 2, func(client Client, msg MQTT.Message) {
		ctx.Debug("Handle new mqtt message")
		if err := h.Handle(client, a.packets, a.registrations, msg); err != nil {
			ctx.WithError(err).Warn("Unable to handle mqtt message")
		}
	})
	if token.Wait() && token.Error() != nil {
		ctx.WithError(token.Error()).Error("Unable to Subscribe")
		return errors.New(errors.Operational, token.Error())
	}
	return nil
}
