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
		return new(core.StatsRes), errors.New(errors.Structural, "Invalid nil stats request")
	}

	if len(req.GatewayID) != 8 {
		return new(core.StatsRes), errors.New(errors.Structural, "Invalid gateway identifier")
	}

	if req.Metadata == nil {
		return new(core.StatsRes), errors.New(errors.Structural, "Missing mandatory Metadata")
	}

	stats.MarkMeter("router.stat.in")
	return new(core.StatsRes), r.Storage.UpdateStats(req.GatewayID, *req.Metadata)
}

// HandleJoin implements the core.RouterClient interface
func (r component) HandleJoin(ctx context.Context, req *core.JoinRouterReq) (*core.JoinRouterRes, error) {
	if len(req.GatewayID) != 8 || len(req.AppEUI) != 8 || len(req.DevEUI) != 8 || len(req.DevNonce) != 2 || req.Metadata == nil {
		r.Ctx.Debug("Invalid request. Parameters are incorrects")
		return new(core.JoinRouterRes), errors.New(errors.Structural, "Invalid Request")
	}

	// Add Gateway location metadata
	if gmeta, err := r.Storage.LookupStats(req.GatewayID); err == nil {
		r.Ctx.Debug("Adding Gateway Metadata to packet")
		req.Metadata.Latitude = gmeta.Latitude
		req.Metadata.Longitude = gmeta.Longitude
		req.Metadata.Altitude = gmeta.Altitude
	}

	// Broadcast the join request
	bpacket := &core.JoinBrokerReq{
		AppEUI:   req.AppEUI,
		DevEUI:   req.DevEUI,
		DevNonce: req.DevNonce,
		Metadata: req.Metadata,
	}
	response, err := r.send(bpacket, true, r.Brokers...)
	if err != nil {
		return new(core.JoinRouterRes), err
	}
	return response.(*core.JoinRouterRes), err
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
		return new(core.DataRouterRes), errors.New(errors.Structural, err)
	}
	if req.Metadata == nil {
		r.Ctx.Debug("Invalid request Metadata")
		return new(core.DataRouterRes), errors.New(errors.Structural, "Missing mandatory Metadata")
	}
	if len(req.GatewayID) != 8 {
		r.Ctx.Debug("Invalid request GatewayID")
		return new(core.DataRouterRes), errors.New(errors.Structural, "Invalid gatewayID")
	}

	// Lookup for an existing broker
	entries, err := r.Storage.Lookup(fhdr.DevAddr)
	if err != nil && err.(errors.Failure).Nature != errors.NotFound {
		r.Ctx.Warn("Database lookup failed")
		return new(core.DataRouterRes), errors.New(errors.Operational, err)
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
		return new(core.DataRouterRes), errors.New(errors.Structural, "Unhandled uplink signal frequency")
	}

	rx1, rx2 := dutycycle.StateFromDuty(cycles[sb1]), dutycycle.StateFromDuty(cycles[dutycycle.EuropeG3])
	req.Metadata.DutyRX1, req.Metadata.DutyRX2 = uint32(rx1), uint32(rx2)
	r.Ctx.WithField("DutyRX1", rx1).WithField("DutyRX2", rx2).Debug("Adding Duty values for RX1 & RX2")

	bpacket := &core.DataBrokerReq{Payload: req.Payload, Metadata: req.Metadata}

	// Send packet to broker(s)
	var response interface{}
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
		return new(core.DataRouterRes), err
	}

	return r.handleDataDown(response.(*core.DataBrokerRes), req.GatewayID)
}

func (r component) handleDataDown(req *core.DataBrokerRes, gatewayID []byte) (*core.DataRouterRes, error) {
	if req == nil || req.Payload == nil { // No response
		r.Ctx.Debug("Packet sent. No downlink received.")
		return new(core.DataRouterRes), nil
	}

	r.Ctx.Debug("Handling downlink response")
	// Update downlink metadata for the related gateway
	if req.Metadata == nil {
		stats.MarkMeter("router.uplink.bad_broker_response")
		r.Ctx.Warn("Missing mandatory Metadata in response")
		return new(core.DataRouterRes), errors.New(errors.Structural, "Missing mandatory Metadata in response")
	}
	freq := req.Metadata.Frequency
	datr := req.Metadata.DataRate
	codr := req.Metadata.CodingRate
	size := req.Metadata.PayloadSize
	if err := r.DutyManager.Update(gatewayID, freq, size, datr, codr); err != nil {
		r.Ctx.WithError(err).Debug("Unable to update DutyManager")
		return new(core.DataRouterRes), errors.New(errors.Operational, err)
	}

	// Send Back the response
	return &core.DataRouterRes{Payload: req.Payload, Metadata: req.Metadata}, nil
}

func (r component) send(req interface{}, isBroadcast bool, brokers ...core.BrokerClient) (interface{}, error) {
	// Define a more helpful context
	nb := len(brokers)
	ctx := r.Ctx.WithField("Nb Brokers", nb)
	ctx.Debug("Sending Packet")
	stats.UpdateHistogram("router.send_recipients", int64(nb))

	// Prepare ground for parrallel requests
	cherr := make(chan error, nb)
	chresp := make(chan struct {
		Response    interface{}
		BrokerIndex int
	}, nb)
	wg := sync.WaitGroup{}
	wg.Add(nb)

	// Run each request
	for i, broker := range brokers {
		go func(index int, broker core.BrokerClient) {
			defer wg.Done()

			// Send request
			var resp interface{}
			var err error
			switch req.(type) {
			case *core.DataBrokerReq:
				resp, err = broker.HandleData(context.Background(), req.(*core.DataBrokerReq))
			case *core.JoinBrokerReq:
				resp, err = broker.HandleJoin(context.Background(), req.(*core.JoinBrokerReq))
			default:
				cherr <- errors.New(errors.Structural, "Unknown request type")
				return
			}

			// Handle error
			if err != nil {
				ctx.WithField("index", index).WithError(err).Debug("Error while contacting broker")
				if strings.Contains(err.Error(), string(errors.NotFound)) { // FIXME Find a better way to analyze the error
					cherr <- errors.New(errors.NotFound, "Broker not responsible for the node")
					return
				}
				cherr <- errors.New(errors.Operational, err)
				return
			}

			// Transfer the response
			chresp <- struct {
				Response    interface{}
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
	for err := range cherr {
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
		var devAddr []byte
		switch req.(type) {
		case *core.DataBrokerReq:
			devAddr = req.(*core.DataBrokerReq).Payload.MACPayload.FHDR.DevAddr
		case *core.JoinBrokerReq:
			devAddr = resp.Response.(*core.JoinBrokerRes).DevAddr
		}
		if err := r.Storage.Store(devAddr, resp.BrokerIndex); err != nil {
			r.Ctx.WithError(err).Warn("Failed to store accepted broker")
		}
	}
	return resp.Response, nil
}
