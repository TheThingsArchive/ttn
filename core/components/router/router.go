// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
)

type component struct {
	Storage
	Manager dutycycle.DutyManager
	ctx     log.Interface
}

// New constructs a new router
func New(db Storage, dm dutycycle.DutyManager, ctx log.Interface) Router {
	return component{Storage: db, Manager: dm, ctx: ctx}
}

// Register implements the core.Router interface
func (r component) Register(reg Registration, an AckNacker) (err error) {
	defer ensureAckNack(an, nil, &err)
	stats.MarkMeter("router.registration.in")
	r.ctx.Debug("Handling registration")

	rreg, ok := reg.(RRegistration)
	if !ok {
		err = errors.New(errors.Structural, "Unexpected registration type")
		r.ctx.WithError(err).Warn("Unable to register")
		return err
	}

	return r.Store(rreg)
}

// HandleUp implements the core.Router interface
func (r component) HandleUp(data []byte, an AckNacker, up Adapter) (err error) {
	// Make sure we don't forget the AckNacker
	var ack Packet
	defer ensureAckNack(an, &ack, &err)

	// Get some logs / analytics
	stats.MarkMeter("router.uplink.in")
	r.ctx.Debug("Handling uplink packet")

	// Extract the given packet
	itf, err := UnmarshalPacket(data)
	if err != nil {
		stats.MarkMeter("router.uplink.invalid")
		r.ctx.Warn("Uplink Invalid")
		return errors.New(errors.Structural, err)
	}

	switch itf.(type) {
	case RPacket:
		packet := itf.(RPacket)

		// Lookup for an existing broker
		entries, err := r.Lookup(packet.DevEUI())
		if err != nil && err.(errors.Failure).Nature != errors.NotFound {
			r.ctx.Warn("Database lookup failed")
			return errors.New(errors.Operational, err)
		}
		shouldBroadcast := err != nil

		// Add Gateway duty metadata
		// TODO add gateway location
		metadata := packet.Metadata()
		if metadata.Freq == nil {
			stats.MarkMeter("router.uplink.not_supported")
			return errors.New(errors.Structural, "Missing mandatory frequency in metadata")
		}

		cycles, err := r.Manager.Lookup(packet.GatewayID())
		if err != nil {
			r.ctx.WithError(err).Debug("Unable to get any metadata about duty-cycles")
			cycles = make(dutycycle.Cycles)
		}

		sb1, err := dutycycle.GetSubBand(*metadata.Freq)
		if err != nil {
			stats.MarkMeter("router.uplink.not_supported")
			return errors.New(errors.Structural, "Unhandled uplink signal frequency")
		}

		rx1, rx2 := uint(dutycycle.StateFromDuty(cycles[sb1])), uint(dutycycle.StateFromDuty(cycles[dutycycle.EuropeG3]))
		metadata.DutyRX1, metadata.DutyRX2 = &rx1, &rx2

		bpacket, err := NewBPacket(packet.Payload(), metadata)
		if err != nil {
			stats.MarkMeter("router.uplink.not_supported")
			r.ctx.WithError(err).Warn("Unable to create router packet")
			return errors.New(errors.Structural, err)
		}

		// Send packet to broker(s)
		var response []byte
		if shouldBroadcast {
			// No Recipient available -> broadcast
			response, err = up.Send(bpacket)
		} else {
			// Recipients are available
			var recipients []Recipient
			for _, e := range entries {
				// Get the actual broker
				recipient, err := up.GetRecipient(e.Recipient)
				if err != nil {
					r.ctx.Warn("Unable to retrieve Recipient")
					return errors.New(errors.Structural, err)
				}
				recipients = append(recipients, recipient)
			}

			// Send the packet
			response, err = up.Send(bpacket, recipients...)
			if err != nil && err.(errors.Failure).Nature == errors.NotFound {
				// Might be a collision with the dev addr, we better broadcast
				response, err = up.Send(bpacket)
			}
			stats.MarkMeter("router.uplink.out")
		}

		if err != nil {
			switch err.(errors.Failure).Nature {
			case errors.NotFound:
				stats.MarkMeter("router.uplink.negative_broker_response")
				r.ctx.WithError(err).Debug("Negative response from Broker")
			default:
				stats.MarkMeter("router.uplink.bad_broker_response")
				r.ctx.WithError(err).Warn("Invalid response from Broker")
			}
			return err
		}

		// No response, stop there
		if response == nil {
			return nil
		}

		itf, err := UnmarshalPacket(response)
		if err != nil {
			stats.MarkMeter("router.uplink.bad_broker_response")
			r.ctx.WithError(err).Warn("Invalid response from Broker")
			return errors.New(errors.Operational, err)
		}

		switch itf.(type) {
		case RPacket:
			// Update downlink metadata for the related gateway
			metadata := itf.(RPacket).Metadata()
			freq := metadata.Freq
			datr := metadata.Datr
			codr := metadata.Codr
			size := metadata.Size

			if freq == nil || datr == nil || codr == nil || size == nil {
				err := errors.New(errors.Operational, "Missing mandatory metadata in response")
				stats.MarkMeter("router.uplink.bad_broker_response")
				r.ctx.WithError(err).Warn("Invalid response from Broker")
				return err
			}

			if err := r.Manager.Update(packet.GatewayID(), *freq, *size, *datr, *codr); err != nil {
				r.ctx.WithError(err).Warn("Unable to update duty cycle")
				return errors.New(errors.Operational, err)
			}

			// Finally, define the ack to be sent
			ack = itf.(RPacket)
		default:
			return errors.New(errors.Implementation, "Unexpected packet type")
		}
		stats.MarkMeter("router.uplink.ok")

	case SPacket:
		return errors.New(errors.Implementation, "Stats packet not yet implemented")
	case JPacket:
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
