// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"net"
	"strings"
	"sync"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Components defines a structure to make the instantiation easier to read
type Components struct {
	DutyManager dutycycle.DutyManager
	Brokers     []core.BrokerClient
	Ctx         log.Interface
	Storage     Storage
}

// Options defines a structure to make the instantiation easier to read
type Options struct {
	NetAddr string
}

// component implements the core.RouterServer interface
type component struct {
	Components
	NetAddr string
}

// Interface defines the Router interface
type Interface interface {
	core.RouterServer
	Start() error
}

// New constructs a new router
func New(c Components, o Options) Interface {
	return component{Components: c, NetAddr: o.NetAddr}
}

// Start actually runs the component and starts the rpc server
func (r component) Start() error {
	conn, err := net.Listen("tcp", r.NetAddr)
	if err != nil {
		return errors.New(errors.Operational, err)
	}

	server := grpc.NewServer()
	core.RegisterRouterServer(server, r)

	if err := server.Serve(conn); err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// HandleStats implements the core.RouterClient interface
func (r component) HandleStats(ctx context.Context, req *core.StatsReq) (*core.StatsRes, error) {
	if req == nil {
		return nil, errors.New(errors.Structural, "Invalid nil stats request")
	}

	if len(req.GatewayID) != 8 {
		return nil, errors.New(errors.Structural, "Invalid gateway identifier")
	}

	if req.Metadata == nil {
		return nil, errors.New(errors.Structural, "Missing mandatory Metadata")
	}

	stats.MarkMeter("router.stat.in")
	return nil, r.Storage.UpdateStats(req.GatewayID, *req.Metadata)
}

// HandleData implements the core.RouterClient interface
func (r component) HandleData(ctx context.Context, req *core.DataRouterReq) (*core.DataRouterRes, error) {
	// Get some logs / analytics
	r.Ctx.Debug("Handling uplink packet")
	stats.MarkMeter("router.uplink.in")

	// Validate coming data
	_, _, fhdr, _, err := core.ValidateLoRaWANData(req.Payload)
	if err != nil {
		r.Ctx.WithError(err).Debug("Invalid request payload")
		return nil, errors.New(errors.Structural, err)
	}
	if req.Metadata == nil {
		r.Ctx.Debug("Invalid request Metadata")
		return nil, errors.New(errors.Structural, "Missing mandatory Metadata")
	}
	if len(req.GatewayID) != 8 {
		r.Ctx.Debug("Invalid request GatewayID")
		return nil, errors.New(errors.Structural, "Invalid gatewayID")
	}

	// Lookup for an existing broker
	entries, err := r.Storage.Lookup(fhdr.DevAddr)
	if err != nil && err.(errors.Failure).Nature != errors.NotFound {
		r.Ctx.Warn("Database lookup failed")
		return nil, errors.New(errors.Operational, err)
	}
	shouldBroadcast := err != nil
	r.Ctx.WithField("Should Broadcast?", shouldBroadcast).Debug("Storage Lookup done")

	// Add Gateway location metadata
	if gmeta, err := r.Storage.LookupStats(req.GatewayID); err == nil {
		r.Ctx.Debug("Adding Gateway Metadata to packet")
		req.Metadata.Latitude = gmeta.Latitude
		req.Metadata.Longitude = gmeta.Longitude
		req.Metadata.Altitude = gmeta.Altitude
	}

	// Add Gateway duty metadata
	cycles, err := r.DutyManager.Lookup(req.GatewayID)
	if err != nil {
		r.Ctx.WithError(err).Debug("Unable to get any metadata about duty-cycles")
		cycles = make(dutycycle.Cycles)
	}

	sb1, err := dutycycle.GetSubBand(float32(req.Metadata.Frequency))
	if err != nil {
		stats.MarkMeter("router.uplink.not_supported")
		return nil, errors.New(errors.Structural, "Unhandled uplink signal frequency")
	}

	rx1, rx2 := dutycycle.StateFromDuty(cycles[sb1]), dutycycle.StateFromDuty(cycles[dutycycle.EuropeG3])
	req.Metadata.DutyRX1, req.Metadata.DutyRX2 = uint32(rx1), uint32(rx2)
	r.Ctx.WithField("DutyRX1", rx1).WithField("DutyRX2", rx2).Debug("Adding Duty values for RX1 & RX2")

	bpacket := &core.DataBrokerReq{Payload: req.Payload, Metadata: req.Metadata}

	// Send packet to broker(s)
	var response *core.DataBrokerRes
	if shouldBroadcast {
		// No Recipient available -> broadcast
		stats.MarkMeter("router.broadcast")
		r.Ctx.Debug("Broadcasting packet to all known brokers")
		response, err = r.send(bpacket, true, r.Brokers...)
	} else {
		// Recipients are available
		stats.MarkMeter("router.send")
		var brokers []core.BrokerClient
		r.Ctx.Debug("Forwarding packet to known broker(s)")
		for _, e := range entries {
			brokers = append(brokers, r.Brokers[e.BrokerIndex])
		}
		response, err = r.send(bpacket, false, brokers...)
		if err != nil && err.(errors.Failure).Nature == errors.NotFound {
			r.Ctx.Debug("No response from known broker(s). Trying again with broadcast")
			// Might be a collision with the dev addr, we better broadcast
			response, err = r.send(bpacket, true, r.Brokers...)
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

	return r.handleDataDown(response, req.GatewayID)
}

func (r component) handleDataDown(req *core.DataBrokerRes, gatewayID []byte) (*core.DataRouterRes, error) {
	if req == nil { // No response
		r.Ctx.Debug("Packet sent. No downlink received.")
		return nil, nil
	}

	r.Ctx.Debug("Handling downlink response")
	// Update downlink metadata for the related gateway
	if req.Metadata == nil {
		stats.MarkMeter("router.uplink.bad_broker_response")
		r.Ctx.Warn("Missing mandatory Metadata in response")
		return nil, errors.New(errors.Structural, "Missing mandatory Metadata in response")
	}
	freq := req.Metadata.Frequency
	datr := req.Metadata.DataRate
	codr := req.Metadata.CodingRate
	size := req.Metadata.PayloadSize
	if err := r.DutyManager.Update(gatewayID, freq, size, datr, codr); err != nil {
		r.Ctx.WithError(err).Debug("Unable to update DutyManager")
		return nil, errors.New(errors.Operational, err)
	}

	// Send Back the response
	return &core.DataRouterRes{Payload: req.Payload, Metadata: req.Metadata}, nil
}

func (r component) send(req *core.DataBrokerReq, isBroadcast bool, brokers ...core.BrokerClient) (*core.DataBrokerRes, error) {
	// Define a more helpful context
	nb := len(brokers)
	ctx := r.Ctx.WithField("devAddr", req.Payload.MACPayload.FHDR.DevAddr)
	ctx.WithField("Nb Brokers", nb).Debug("Sending Packet")
	stats.UpdateHistogram("router.send_recipients", int64(nb))

	// Prepare ground for parrallel requests
	cherr := make(chan error, nb)
	chresp := make(chan struct {
		Response    *core.DataBrokerRes
		BrokerIndex int
	}, nb)
	wg := sync.WaitGroup{}
	wg.Add(nb)

	// Run each request
	for i, broker := range brokers {
		go func(index int, broker core.BrokerClient) {
			defer wg.Done()

			// Send request
			resp, err := broker.HandleData(context.Background(), req)

			// Handle error
			if err != nil {
				if strings.Contains(err.Error(), string(errors.NotFound)) { // FIXME Find a better way to analyze the error
					cherr <- errors.New(errors.NotFound, "Broker not responsible for the node")
					return
				}
				cherr <- errors.New(errors.Operational, err)
				return
			}

			// Transfer the response
			chresp <- struct {
				Response    *core.DataBrokerRes
				BrokerIndex int
			}{resp, index}
		}(i, broker)
	}

	// Wait for each request to be done
	stats.IncCounter("router.waiting_for_send")
	wg.Wait()
	stats.DecCounter("router.waiting_for_send")
	close(cherr)
	close(chresp)

	var errored uint8
	var notFound uint8
	for i := 0; i < len(cherr); i++ {
		err := <-cherr
		if err.(errors.Failure).Nature != errors.NotFound {
			errored++
			ctx.WithError(err).Warn("Failed to contact broker")
		} else {
			notFound++
			ctx.WithError(err).Debug("Packet destination not found")
		}
	}

	// Collect response
	if len(chresp) > 1 {
		r.Ctx.Warn("Received too many positive answers")
		return nil, errors.New(errors.Behavioural, "Received too many positive answers")
	}

	if len(chresp) == 0 && errored > 0 {
		r.Ctx.Debug("No positive response but got errored response(s)")
		return nil, errors.New(errors.Operational, "No positive response from recipients but got unexpected answer")
	}

	if len(chresp) == 0 && notFound > 0 {
		r.Ctx.Debug("No available recipient found")
		return nil, errors.New(errors.NotFound, "No available recipient found")
	}

	if len(chresp) == 0 {
		return nil, nil
	}

	resp := <-chresp
	// Save the broker for later if it was a broadcast
	if isBroadcast {
		if err := r.Storage.Store(req.Payload.MACPayload.FHDR.DevAddr, resp.BrokerIndex); err != nil {
			r.Ctx.WithError(err).Warn("Failed to store accepted broker")
		}
	}
	return resp.Response, nil
}
