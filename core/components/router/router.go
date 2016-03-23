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
	BrkStorage  BrkStorage
	GtwStorage  GtwStorage
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

// Server defines the Router Server interface
type Server interface {
	core.RouterServer
	Start() error
}

// New constructs a new router
func New(c Components, o Options) Server {
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
	return new(core.StatsRes), r.GtwStorage.upsert(gtwEntry{
		GatewayID: req.GatewayID,
		Metadata:  *req.Metadata,
	})
}

// HandleJoin implements the core.RouterClient interface
func (r component) HandleJoin(ctx context.Context, req *core.JoinRouterReq) (routerRes *core.JoinRouterRes, err error) {
	if len(req.GatewayID) != 8 || len(req.AppEUI) != 8 || len(req.DevEUI) != 8 || len(req.DevNonce) != 2 || len(req.MIC) != 4 || req.Metadata == nil {
		r.Ctx.Debug("Invalid request. Parameters are incorrects")
		return new(core.JoinRouterRes), errors.New(errors.Structural, "Invalid Request")
	}

	// Update Metadata with Gateway infos
	req.Metadata, err = r.injectMetadata(req.GatewayID, *req.Metadata)
	if err != nil {
		return new(core.JoinRouterRes), err
	}

	// Broadcast the join request
	bpacket := &core.JoinBrokerReq{
		AppEUI:   req.AppEUI,
		DevEUI:   req.DevEUI,
		DevNonce: req.DevNonce,
		MIC:      req.MIC,
		Metadata: req.Metadata,
	}
	response, err := r.send(bpacket, true, r.Brokers...)
	if err != nil {
		return new(core.JoinRouterRes), err
	}

	// Update Gateway Duty cycle with response metadata
	res := response.(*core.JoinBrokerRes)
	if res == nil || res.Payload == nil { // No response
		r.Ctx.Debug("No Join-Accept received")
		return new(core.JoinRouterRes), nil
	}
	if err := r.handleDown(req.GatewayID, res.Metadata); err != nil {
		return new(core.JoinRouterRes), err
	}
	return &core.JoinRouterRes{Payload: res.Payload, Metadata: res.Metadata}, nil
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

	// Update Metadata with Gateway infos
	req.Metadata, err = r.injectMetadata(req.GatewayID, *req.Metadata)
	if err != nil {
		return new(core.DataRouterRes), err
	}

	// Lookup for an existing broker
	entries, err := r.BrkStorage.read(fhdr.DevAddr)
	if err != nil && err.(errors.Failure).Nature != errors.NotFound {
		r.Ctx.Warn("Database lookup failed")
		return new(core.DataRouterRes), errors.New(errors.Operational, err)
	}
	shouldBroadcast := err != nil
	r.Ctx.WithField("Should Broadcast?", shouldBroadcast).Debug("Storage Lookup done")

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

	res := response.(*core.DataBrokerRes)
	if res == nil || res.Payload == nil { // No response
		r.Ctx.Debug("Packet sent. No downlink received.")
		return new(core.DataRouterRes), nil
	}

	// Update Gateway Duty cycle with response metadata
	if err := r.handleDown(req.GatewayID, res.Metadata); err != nil {
		return new(core.DataRouterRes), err
	}

	// Send Back the response
	return &core.DataRouterRes{Payload: res.Payload, Metadata: res.Metadata}, nil
}

func (r component) injectMetadata(gid []byte, metadata core.Metadata) (*core.Metadata, error) {
	// Add Gateway location metadata
	if entry, err := r.GtwStorage.read(gid); err == nil {
		r.Ctx.Debug("Adding Gateway Metadata to packet")
		metadata.Latitude = entry.Metadata.Latitude
		metadata.Longitude = entry.Metadata.Longitude
		metadata.Altitude = entry.Metadata.Altitude
	}

	// Add Gateway duty metadata
	cycles, err := r.DutyManager.Lookup(gid)
	if err != nil {
		r.Ctx.WithError(err).Debug("Unable to get any metadata about duty-cycles")
		cycles = make(dutycycle.Cycles)
	}

	sb1, err := dutycycle.GetSubBand(float32(metadata.Frequency))
	if err != nil {
		stats.MarkMeter("router.uplink.not_supported")
		return nil, errors.New(errors.Structural, "Unhandled uplink signal frequency")
	}

	rx1, rx2 := dutycycle.StateFromDuty(cycles[sb1]), dutycycle.StateFromDuty(cycles[dutycycle.EuropeG3])
	metadata.DutyRX1, metadata.DutyRX2 = uint32(rx1), uint32(rx2)
	r.Ctx.WithField("DutyRX1", rx1).WithField("DutyRX2", rx2).Debug("Adding Duty values for RX1 & RX2")
	return &metadata, nil

}

func (r component) handleDown(gatewayID []byte, metadata *core.Metadata) error {
	r.Ctx.Debug("Handling downlink response")

	// Update downlink metadata for the related gateway
	if metadata == nil {
		stats.MarkMeter("router.uplink.bad_broker_response")
		r.Ctx.Warn("Missing mandatory Metadata in response")
		return errors.New(errors.Structural, "Missing mandatory Metadata in response")
	}
	freq := metadata.Frequency
	datr := metadata.DataRate
	codr := metadata.CodingRate
	size := metadata.PayloadSize
	if err := r.DutyManager.Update(gatewayID, freq, size, datr, codr); err != nil {
		r.Ctx.WithError(err).Debug("Unable to update DutyManager")
		return errors.New(errors.Operational, err)
	}
	return nil
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
		BrokerIndex uint16
	}, nb)
	wg := sync.WaitGroup{}
	wg.Add(nb)

	// Run each request
	for i, broker := range brokers {
		go func(index uint16, broker core.BrokerClient) {
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
				BrokerIndex uint16
			}{resp, index}
		}(uint16(i), broker)
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
		err := r.BrkStorage.create(brkEntry{
			DevAddr:     devAddr,
			BrokerIndex: resp.BrokerIndex,
		})
		if err != nil {
			r.Ctx.WithError(err).Warn("Failed to store accepted broker")
		}
	}
	return resp.Response, nil
}
