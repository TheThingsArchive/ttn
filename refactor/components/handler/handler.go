// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	. "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

// component implements the core.Component interface
type component struct {
	ctx     log.Interface
	devices devStorage
	set     chan<- bundle
}

type bundle struct {
	adapter Adapter
	chresp  chan interface{}
	entry   devEntry
	id      [16]byte
	packet  HPacket
}

// New construct a new Handler component from ...
func New() (Component, error) {
	return nil, nil
}

// Register implements the core.Component interface
func (h component) Register(reg Registration, an AckNacker) error {
	return nil
}

// HandleUp implements the core.Component interface
func (h component) HandleUp(data []byte, an AckNacker, up Adapter) error {
	itf, err := UnmarshalPacket(data)
	if err != nil {
		return errors.New(errors.Structural, data)
	}

	switch itf.(type) {
	case HPacket:
		// 0. Retrieve the handler packet
		packet := itf.(HPacket)
		appEUI := packet.AppEUI()
		devEUI := packet.DevEUI()

		// 1. Lookup for the associated AppSKey + Recipient
		entry, err := h.devices.Lookup(appEUI, devEUI)
		if err != nil {
			return errors.New(errors.Operational, err)
		}

		// 2. Prepare a channel to receive the response from the consumer
		chresp := make(chan interface{})

		// 3. Create a "bundle" which holds info waiting for other related packets
		var bundleId [16]byte // AppEUI(8) | DevEUI(8)
		copy(bundleId[:8], appEUI[:])
		copy(bundleId[8:], devEUI[:])

		// 4. Send the actual bundle to the consumer
		ctx := h.ctx.WithField("BundleID", bundleId)
		ctx.Debug("Define new bundle")
		h.set <- bundle{
			id:      bundleId,
			packet:  packet,
			entry:   entry,
			adapter: up,
			chresp:  chresp,
		}

		// 5. Wait for the response. Could be an error, a packet or nothing.
		// We'll respond to a maximum of one node. The handler will use the
		// rssi + gateway's duty cycle to select to best fit.
		// All other channels will get a nil response.
		// If there's an error, all channels get the error.
		resp := <-chresp
		switch resp.(type) {
		case Packet:
			ctx.Debug("Received response with packet. Sending Ack")
			an.Ack(resp.(Packet))
		case error:
			ctx.WithError(resp.(error)).Warn("Received errored response. Sending Ack")
			an.Nack()
			return errors.New(errors.Operational, resp.(error))
		default:
			ctx.Debug("Received empty response. Sending empty Ack")
			an.Ack(nil)
		}

		return nil
	case JPacket:
		return errors.New(errors.Implementation, "Join Request not yet implemented")
	default:
		return errors.New(errors.Structural, "Unhandled packet type")
	}
}

// consumeBundles processes list of bundle generated overtime, decrypt the underlying packet,
// deduplicate them, and send a single enhanced packet to the upadapter for further processing.
func (h component) consumeBundles(chbundle <-chan []bundle) {
	ctx := h.ctx.WithField("goroutine", "consumer")
	ctx.Debug("Starting bundle consumer")

browseBundles:
	for bundles := range chbundle {
		var metadata []Metadata
		var devEUI lorawan.EUI64
		var payload []byte
		var adapter Adapter
		var recipient Recipient

		for i, bundle := range bundles {
			// We only decrypt the payload of the first bundle's packet.
			// We assume all the other to be equal and we'll merely collect
			// metadata from other bundle.
			if i == 0 {
				var err error
				payload, err = bundle.packet.Payload(bundle.entry.AppSKey)
				if err != nil {
					go h.abortConsume(err, bundles)
					continue browseBundles
				}
				devEUI = bundle.packet.DevEUI()
				adapter = bundle.adapter
				recipient = bundle.entry.Recipient
			}

			// And append metadata for each of them
			metadata = append(metadata, bundle.packet.Metadata())
		}

		// Then create an application-level packet
		packet, err := NewAPacket(payload, devEUI, metadata)
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		// And send it
		_, err = adapter.Send(packet, recipient)
		if err != nil {
			go h.abortConsume(err, bundles)
			continue browseBundles
		}

		// Then respond to node -> no response for the moment
		for _, bundle := range bundles {
			bundle.chresp <- nil
		}
	}
}

func (h component) abortConsume(fault error, bundles []bundle) {
	err := errors.New(errors.Structural, fault)
	h.ctx.WithError(err).Debug("Unable to consume bundle")
	for _, bundle := range bundles {
		bundle.chresp <- err
	}
}

// HandleDown implements the core.Component interface
func (h component) HandleDown(p []byte, an AckNacker, down Adapter) error {
	return nil
}
