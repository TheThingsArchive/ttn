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
	r.ctx.Debug("Handling uplink packet")

	// Extract the given packet
	itf, _ := UnmarshalPacket(data)
	switch itf.(type) {
	case RPacket:
		stats.MarkMeter("router.uplink.in")
		ack, err = r.handleDataUp(itf.(RPacket), up)
	case SPacket:
		stats.MarkMeter("router.stat.in")
		err = r.UpdateStats(itf.(SPacket))
	default:
		stats.MarkMeter("router.uplink.invalid")
		err = errors.New(errors.Structural, "Unreckognized packet type")
	}

	if err != nil {
		r.ctx.WithError(err).Debug("Unable to process uplink")
	}
	return err
}

// handleDataUp handle an upcoming message which carries a data frame payload
func (r component) handleDataUp(packet RPacket, up Adapter) (Packet, error) {
	// Lookup for an existing broker
	entries, err := r.Lookup(packet.DevEUI())
	if err != nil && err.(errors.Failure).Nature != errors.NotFound {
		r.ctx.Warn("Database lookup failed")
		return nil, errors.New(errors.Operational, err)
	}
	shouldBroadcast := err != nil

	metadata := packet.Metadata()
	if metadata.Freq == nil {
		stats.MarkMeter("router.uplink.not_supported")
		return nil, errors.New(errors.Structural, "Missing mandatory frequency in metadata")
	}

	// Add Gateway location metadata
	gmeta, _ := r.LookupStats(packet.GatewayID())
	metadata.Lati = gmeta.Lati
	metadata.Long = gmeta.Long
	metadata.Alti = gmeta.Alti

	// Add Gateway duty metadata
	cycles, err := r.Manager.Lookup(packet.GatewayID())
	if err != nil {
		r.ctx.WithError(err).Debug("Unable to get any metadata about duty-cycles")
		cycles = make(dutycycle.Cycles)
	}

	sb1, err := dutycycle.GetSubBand(*metadata.Freq)
	if err != nil {
		stats.MarkMeter("router.uplink.not_supported")
		return nil, errors.New(errors.Structural, "Unhandled uplink signal frequency")
	}

	rx1, rx2 := uint(dutycycle.StateFromDuty(cycles[sb1])), uint(dutycycle.StateFromDuty(cycles[dutycycle.EuropeG3]))
	metadata.DutyRX1, metadata.DutyRX2 = &rx1, &rx2

	bpacket, err := NewBPacket(packet.Payload(), metadata)
	if err != nil {
		stats.MarkMeter("router.uplink.not_supported")
		r.ctx.WithError(err).Warn("Unable to create router packet")
		return nil, errors.New(errors.Structural, err)
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
				return nil, errors.New(errors.Structural, err)
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
		default:
			stats.MarkMeter("router.uplink.bad_broker_response")
		}
		return nil, err
	}

	return r.handleDataDown(response, packet.GatewayID())
}

// handleDataDown controls that data received from an uplink are okay.
// It also updates metadata about the related gateway
func (r component) handleDataDown(data []byte, gatewayID []byte) (Packet, error) {
	if data == nil {
		return nil, nil
	}

	itf, err := UnmarshalPacket(data)
	if err != nil {
		stats.MarkMeter("router.uplink.bad_broker_response")
		return nil, errors.New(errors.Operational, err)
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
			stats.MarkMeter("router.uplink.bad_broker_response")
			return nil, errors.New(errors.Operational, "Missing mandatory metadata in response")
		}

		if err := r.Manager.Update(gatewayID, *freq, *size, *datr, *codr); err != nil {
			return nil, errors.New(errors.Operational, err)
		}

		// Finally, define the ack to be sent
		return itf.(RPacket), nil
	default:
		stats.MarkMeter("router.uplink.bad_broker_response")
		return nil, errors.New(errors.Implementation, "Unexpected packet type")
	}
}

// ensureAckNack is used to make sure we Ack / Nack correctly in the HandleUp method.
// The method will probably change or be moved outside the router itself.
func ensureAckNack(an AckNacker, ack *Packet, err *error) {
	if err != nil && *err != nil {
		an.Nack(*err)
	} else {
		stats.MarkMeter("router.uplink.ok")
		var p Packet
		if ack != nil {
			p = *ack
		}
		an.Ack(p)
	}
}
