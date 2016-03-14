// Copyright © 2016 The Things Network
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
	NetworkController
	ctx log.Interface
}

// New construct a new Broker component
func New(controller NetworkController, ctx log.Interface) Broker {
	return component{NetworkController: controller, ctx: ctx}
}

// Register implements the core.Broker interface
func (b component) Register(reg Registration, an AckNacker) (err error) {
	defer ensureAckNack(an, nil, &err)
	stats.MarkMeter("broker.registration.in")
	b.ctx.Debug("Handling registration")

	switch reg.(type) {
	case BRegistration:
		err = b.StoreDevice(reg.(BRegistration))
	case ARegistration:
		err = b.StoreApplication(reg.(ARegistration))
	default:
		err = errors.New(errors.Structural, "Not a Broker registration")
	}

	return err
}

// HandleUp implements the core.Broker interface
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
		// 0. Retrieve the packet
		packet := itf.(BPacket)
		ctx := b.ctx.WithField("DevEUI", packet.DevEUI())

		// 1. Check whether we should handle it
		entries, err := b.LookupDevices(packet.DevEUI())
		if err != nil {
			switch err.(errors.Failure).Nature {
			case errors.NotFound:
				stats.MarkMeter("broker.uplink.handler_lookup.device_not_found")
				ctx.Debug("Uplink device not found")
			default:
				b.ctx.Warn("Database lookup failed")
			}
			return err
		}
		stats.UpdateHistogram("broker.uplink.handler_lookup.entries", int64(len(entries)))

		// 2. Several handlers might be associated to the same device, we distinguish them using
		// MIC check. Only one should verify the MIC check
		var mEntry *devEntry
		for _, entry := range entries {
			// The device only stores a 16-bits counter but could reflect a 32-bits one.
			// We keep track of the real counter in the network controller.
			if err := packet.ComputeFCnt(entry.FCntUp); err != nil {
				continue
			}
			ok, err := packet.ValidateMIC(entry.NwkSKey)
			if err != nil {
				continue
			}
			if ok {
				mEntry = &entry
				stats.MarkMeter("broker.uplink.handler_lookup.mic_match")
				ctx.WithField("handler", entry.Recipient).Debug("MIC check associated device with handler")
				break
			}
		}
		if mEntry == nil {
			stats.MarkMeter("broker.uplink.handler_lookup.no_mic_match")
			err := errors.New(errors.NotFound, "MIC check returned no matches")
			ctx.Debug(err.Error())
			return err
		}

		// It does matter here to use the DevEUI from the entry and not from the packet.
		// The packet actually holds a DevAddr and the real DevEUI has been determined thanks
		// to the MIC check
		b.UpdateFCnt(mEntry.AppEUI, mEntry.DevEUI, packet.FCnt())

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
		if err != nil && err.(errors.Failure).Nature != errors.NotFound {
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

			// TODO Compute mic check
			stats.MarkMeter("broker.downlink.in")
			if err := bpacket.SetMIC(mEntry.NwkSKey); err != nil {
				return errors.New(errors.Structural, "Unable to set response MIC")
			}
		}

		// 6. And finally, we acknowledge the answer
		if bpacket != nil {
			rpacket, err := NewRPacket(bpacket.Payload(), []byte{}, bpacket.Metadata())
			if err != nil {
				return errors.New(errors.Structural, "Invalid downlink packet from the handler")
			}
			stats.MarkMeter("broker.downlink.out")
			ack = rpacket
		}
	case JPacket:
		// TODO
		return errors.New(errors.Implementation, "Join Request not yet implemented")
	default:
		return errors.New(errors.Implementation, "Unreckognized packet type")
	}

	return nil
}

func ensureAckNack(an AckNacker, ack *Packet, err *error) {
	if err != nil && *err != nil {
		an.Nack(*err)
	} else {
		var p Packet
		if ack != nil {
			p = *ack
		}
		an.Ack(p)
	}
}
