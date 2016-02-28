// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
)

// component implements the core.Component interface
type component struct {
	Storage
	ctx        log.Interface
	Controller NetworkController
}

// New construct a new Broker component
func New(db Storage, ctx log.Interface) Component {
	return component{Storage: db, ctx: ctx}
}

// Register implements the core.Component interface
func (b component) Register(reg Registration, an AckNacker) (err error) {
	defer ensureAckNack(an, nil, &err)
	stats.MarkMeter("broker.registration.in")
	b.ctx.Debug("Handling registration")

	breg, ok := reg.(BRegistration)
	if !ok {
		return errors.New(errors.Structural, "Not a Broker registration")
	}

	if err := b.Store(breg); err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// HandleUp implements the core.Component interface
func (b component) HandleUp(data []byte, an AckNacker, up Adapter) (err error) {
	// Make sure we don't forget the AckNacker
	var ack Packet
	defer ensureAckNack(an, &ack, &err)

	// Get some logs / analytics
	stats.MarkMeter("broker.uplink.in")
	b.ctx.Debug("Handling uplink packet")

	// Extract the given packet
	itf, err := UnmarshalPacket(data)
	if err != nil {
		stats.MarkMeter("broker.uplink.invalid")
		b.ctx.Warn("Uplink Invalid")
		return errors.New(errors.Structural, err)
	}

	switch itf.(type) {
	case BPacket:
		// NOTE So far, we're not using the Frame Counter. This has to be done to get a correct
		// behavior. The frame counter needs to be used to ensure we're not processing a wrong or
		// late packet.

		// 0. Retrieve the packet
		packet := itf.(BPacket)
		ctx := b.ctx.WithField("DevEUI", packet.DevEUI())

		// 1. Check whether we should handle it
		entries, err := b.Lookup(packet.DevEUI())
		if err != nil {
			switch err.(errors.Failure).Nature {
			case errors.Behavioural:
				stats.MarkMeter("broker.uplink.handler_lookup.device_not_found")
				ctx.Warn("Uplink device not found")
			default:
				b.ctx.Warn("Database lookup failed")
			}
			return err
		}
		stats.UpdateHistogram("broker.uplink.handler_lookup.entries", int64(len(entries)))

		// 2. Several handlers might be associated to the same device, we distinguish them using
		// MIC check. Only one should verify the MIC check

		var mEntry *entry
		for _, entry := range entries {
			ok, err := packet.ValidateMIC(entry.NwkSKey)
			if err != nil {
				continue
			}
			if ok {
				mEntry = &entry
				stats.MarkMeter("broker.uplink.handler_lookup.match")
				ctx.WithField("handler", entry.Recipient).Debug("Associated device with handler")
				break
			}
		}
		if mEntry == nil {
			stats.MarkMeter("broker.uplink.handler_lookup.no_match")
			err := errors.New(errors.Behavioural, "Could not find handler for device")
			ctx.Warn(err.Error())
			return err
		}

		// It does matter here to use the DevEUI from the entry and not from the packet.
		// The packet actually holds a DevAddr and the real DevEUI has been determined thanks
		// to the MIC check

		// 3. If one was found, we notify the network controller
		if err := b.Controller.HandleCommands(packet); err != nil {
			ctx.WithError(err).Error("Failed to handle mac commands")
			// Shall we return ? Sounds quite safe to keep going
		}
		b.Controller.UpdateFCntUp(mEntry.AppEUI, mEntry.DevEUI, packet.FCnt())

		// 4. Then we forward the packet to the handler and wait for the response
		hpacket, err := NewHPacket(mEntry.AppEUI, mEntry.DevEUI, packet.Payload(), packet.Metadata())
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		recipient, err := up.GetRecipient(mEntry.Recipient)
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		resp, err := up.Send(hpacket, recipient)
		if err != nil {
			stats.MarkMeter("broker.uplink.bad_handler_response")
			return errors.New(errors.Operational, err)
		}
		stats.MarkMeter("broker.uplink.ok")

		// 5. If a response was sent, i.e. a downlink data, we notify the network controller
		var bpacket BPacket
		if resp != nil {
			itf, err := UnmarshalPacket(resp)
			if err != nil {
				return errors.New(errors.Operational, err)
			}
			var ok bool
			bpacket, ok = itf.(BPacket)
			if !ok {
				return errors.New(errors.Operational, "Received unexpected response")
			}
			b.Controller.UpdateFCntDown(mEntry.AppEUI, mEntry.DevEUI, bpacket.FCnt())
		}

		// 6. And finally, we acknowledge the answer
		ack = b.Controller.MergeCommands(mEntry.AppEUI, mEntry.DevEUI, bpacket)
	case JPacket:
		// TODO
		return errors.New(errors.Implementation, "Join Request not yet implemented")
	default:
		return errors.New(errors.Implementation, "Unreckognized packet type")
	}

	return nil
}

// HandleDown implements the core.Component interface
func (b component) HandleDown(data []byte, an AckNacker, down Adapter) error {
	return errors.New(errors.Implementation, "Handle Down not implemented on broker")
}

func ensureAckNack(an AckNacker, ack *Packet, err *error) {
	if err != nil && *err != nil {
		an.Nack()
	} else {
		var p Packet
		if ack != nil {
			p = *ack
		}
		an.Ack(p)
	}
}
