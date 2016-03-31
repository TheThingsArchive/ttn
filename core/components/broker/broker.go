// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"net"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/TheThingsNetwork/ttn/utils/tokenkey"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// component implements the core.BrokerServer interface
type component struct {
	Components
	NetAddrUp        string
	NetAddrDown      string
	TokenKeyProvider tokenkey.Provider
	MaxDevNonces     uint
}

// Components defines a structure to make the instantiation easier to read
type Components struct {
	NetworkController NetworkController
	AppStorage        AppStorage
	Ctx               log.Interface
}

// Options defines a structure to make the instantiation easier to read
type Options struct {
	NetAddrUp        string
	NetAddrDown      string
	TokenKeyProvider tokenkey.Provider
}

// Interface defines the Broker interface
type Interface interface {
	core.BrokerServer
	core.BrokerManagerServer
	Start() error
}

// New construct a new Broker component
func New(c Components, o Options) Interface {
	return component{
		Components:       c,
		NetAddrUp:        o.NetAddrUp,
		NetAddrDown:      o.NetAddrDown,
		TokenKeyProvider: o.TokenKeyProvider,
		MaxDevNonces:     10,
	}
}

// Start actually runs the component and starts the rpc server
func (b component) Start() error {
	connUp, err := net.Listen("tcp", b.NetAddrUp)
	if err != nil {
		return errors.New(errors.Operational, err)
	}

	connDown, err := net.Listen("tcp", b.NetAddrDown)
	if err != nil {
		return errors.New(errors.Operational, err)
	}

	if b.TokenKeyProvider != nil {
		tokenKey, err := b.TokenKeyProvider.Get(true)
		if err != nil {
			return errors.New(errors.Operational, fmt.Sprintf("Failed to refresh token key: %s", err.Error()))
		}
		b.Ctx.WithField("provider", b.TokenKeyProvider).Infof("Got token key for algorithm %v", tokenKey.Algorithm)
	}

	server := grpc.NewServer()
	core.RegisterBrokerServer(server, b)
	core.RegisterBrokerManagerServer(server, b)

	cherr := make(chan error)

	go func() {
		cherr <- server.Serve(connUp)
	}()

	go func() {
		cherr <- server.Serve(connDown)
	}()

	if err := <-cherr; err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// HandleJoin implements the core.BrokerServer interface
func (b component) HandleJoin(bctx context.Context, req *core.JoinBrokerReq) (*core.JoinBrokerRes, error) {
	// Validate incoming data
	if req == nil || len(req.DevEUI) != 8 || len(req.AppEUI) != 8 || len(req.DevNonce) != 2 || len(req.MIC) != 4 || req.Metadata == nil {
		b.Ctx.Debug("Invalid join request. Parameters are incorrect.")
		return new(core.JoinBrokerRes), errors.New(errors.Structural, "Invalid parameters")
	}

	ctx := b.Ctx.WithFields(log.Fields{
		"AppEUI": req.AppEUI,
		"DevEUI": req.DevEUI,
	})

	ctx.Debug("Handle join request")

	// Check if devNonce already referenced
	appEntry, err := b.AppStorage.read(req.AppEUI)
	if err != nil {
		ctx.WithError(err).Debug("Unable to lookup activation")
		return new(core.JoinBrokerRes), err
	}
	nonces, err := b.NetworkController.readNonces(req.AppEUI, req.DevEUI)
	if err != nil {
		if err.(errors.Failure).Nature != errors.NotFound {
			ctx.WithError(err).Debug("Unable to retrieve associated devNonces")
			return new(core.JoinBrokerRes), err
		}
		nonces = noncesEntry{
			AppEUI: req.AppEUI,
			DevEUI: req.DevEUI,
		}
	}

	var found bool
	for _, n := range nonces.DevNonces {
		if n[0] == req.DevNonce[0] && n[1] == req.DevNonce[1] {
			found = true
			break
		}
	}
	if found {
		ctx.Debug("DevNonce already used in the past")
		return new(core.JoinBrokerRes), errors.New(errors.Structural, "DevNonce used by the past")
	}

	// Forward the registration to the handler
	handler, closer, err := appEntry.Dialer.Dial()
	if err != nil {
		ctx.WithError(err).Debug("Unable to dial handler")
		return new(core.JoinBrokerRes), err
	}
	defer closer.Close()
	res, err := handler.HandleJoin(context.Background(), &core.JoinHandlerReq{
		DevEUI:   req.DevEUI,
		AppEUI:   req.AppEUI,
		DevNonce: req.DevNonce,
		MIC:      req.MIC,
		Metadata: req.Metadata,
	})
	if err != nil {
		ctx.WithError(err).Debug("Error while contacting handler")
		return new(core.JoinBrokerRes), errors.New(errors.Operational, err)
	}

	// Handle the response
	if res == nil || res.Payload == nil {
		ctx.Debug("No join-accept received")
		return new(core.JoinBrokerRes), nil
	}

	if len(res.DevAddr) != 4 || len(res.NwkSKey) != 16 {
		ctx.Debug("Invalid response from handler")
		return new(core.JoinBrokerRes), errors.New(errors.Operational, "Invalid response from handler")
	}

	ctx.WithField("DevAddr", res.DevAddr).Debug("Handle join-accept")

	// Update the DevNonce
	nonces.DevNonces = append(nonces.DevNonces, req.DevNonce)
	if uint(len(nonces.DevNonces)) > b.MaxDevNonces {
		nonces.DevNonces = nonces.DevNonces[1:]
	}
	err = b.NetworkController.upsertNonces(nonces)
	if err != nil {
		ctx.WithError(err).Debug("Unable to update activation entry")
		return new(core.JoinBrokerRes), err
	}

	var nwkSKey [16]byte
	copy(nwkSKey[:], res.NwkSKey)
	err = b.NetworkController.upsert(devEntry{
		Dialer:  appEntry.Dialer,
		DevAddr: res.DevAddr,
		AppEUI:  req.AppEUI,
		DevEUI:  req.DevEUI,
		NwkSKey: nwkSKey,
		FCntUp:  0,
	})
	if err != nil {
		ctx.WithError(err).Debug("Unable to store device")
		return new(core.JoinBrokerRes), err
	}
	return &core.JoinBrokerRes{
		Payload:  res.Payload,
		Metadata: res.Metadata,
	}, nil
}

// HandleData implements the core.BrokerServer interface
func (b component) HandleData(bctx context.Context, req *core.DataBrokerReq) (*core.DataBrokerRes, error) {
	// Get some logs / analytics
	stats.MarkMeter("broker.uplink.in")

	// Validate incoming data
	uplinkPayload, err := core.NewLoRaWANData(req.Payload, true)
	if err != nil {
		b.Ctx.WithError(err).Debug("Unable to interpret LoRaWAN payload")
		return new(core.DataBrokerRes), errors.New(errors.Structural, err)
	}
	devAddr := req.Payload.MACPayload.FHDR.DevAddr // No nil ref, ensured by NewLoRaWANData()

	ctx := b.Ctx.WithField("DevAddr", devAddr)
	ctx.Debug("Handle uplink")

	// Check whether we should handle it
	entries, err := b.NetworkController.read(devAddr)
	if err != nil {
		switch err.(errors.Failure).Nature {
		case errors.NotFound:
			stats.MarkMeter("broker.uplink.handler_lookup.device_not_found")
			ctx.Debug("Uplink device not found")
		default:
			ctx.WithError(err).Warn("Database lookup failed")
		}
		return new(core.DataBrokerRes), err
	}
	stats.UpdateHistogram("broker.uplink.handler_lookup.entries", int64(len(entries)))

	// Several handlers might be associated to the same device, we distinguish them using
	// MIC check. Only one should verify the MIC check
	// The device only stores a 16-bits counter but could reflect a 32-bits one.
	// The counter is used for the MIC computation, thus, we're gonna try both 16-bits and
	// 32-bits counters.
	// We keep track of the real counter in the network controller.
	fhdr := &uplinkPayload.MACPayload.(*lorawan.MACPayload).FHDR // No nil ref, ensured by NewLoRaWANData()
	fcnt16 := fhdr.FCnt                                          // Keep a reference to the original counter

	var mEntry *devEntry
	for _, entry := range entries {
		// retrieve the network session key
		key := lorawan.AES128Key(entry.NwkSKey)

		// Check with 16-bits counters
		fhdr.FCnt = fcnt16
		fcnt32, err := b.NetworkController.wholeCounter(fcnt16, entry.FCntUp)
		if err != nil {
			continue
		}

		ok, err := uplinkPayload.ValidateMIC(key)
		if err != nil {
			continue
		}
		fhdr.FCnt = fcnt32
		if !ok { // Check with 32-bits counter
			ok, err = uplinkPayload.ValidateMIC(key)
			if err != nil {
				continue
			}
		}

		if ok {
			mEntry = &entry
			stats.MarkMeter("broker.uplink.handler_lookup.mic_match")
			ctx = ctx.WithFields(log.Fields{
				"DevEUI":  mEntry.DevEUI,
				"Handler": string(entry.Dialer.MarshalSafely()),
			})
			ctx.Debug("MIC check succeeded")
			break // We stop at the first valid check ...
		}
	}

	if mEntry == nil {
		stats.MarkMeter("broker.uplink.handler_lookup.no_mic_match")
		err := errors.New(errors.NotFound, "FCntUp or MIC check did not match")
		ctx.WithError(err).Debug("Unable to handle uplink")
		return new(core.DataBrokerRes), err
	}

	// It does matter here to use the DevEUI from the entry and not from the packet.
	// The packet actually holds a DevAddr and the real DevEUI has been determined thanks
	// to the MIC check + persistence
	mEntry.FCntUp = fhdr.FCnt
	if err := b.NetworkController.upsert(*mEntry); err != nil {
		ctx.WithError(err).Debug("Unable to update Frame Counter")
		return new(core.DataBrokerRes), err
	}

	// Then we forward the packet to the handler and wait for the response
	handler, closer, err := mEntry.Dialer.Dial()
	if err != nil {
		ctx.WithError(err).Debug("Unable to dial handler")
		return new(core.DataBrokerRes), err
	}
	defer closer.Close()
	resp, err := handler.HandleDataUp(context.Background(), &core.DataUpHandlerReq{
		Payload:  req.Payload.MACPayload.FRMPayload,
		DevEUI:   mEntry.DevEUI,
		AppEUI:   mEntry.AppEUI,
		FCnt:     fhdr.FCnt,
		MType:    req.Payload.MHDR.MType,
		Metadata: req.Metadata,
	})

	if err != nil {
		stats.MarkMeter("broker.uplink.bad_handler_response")
		ctx.WithError(err).Debug("Unexpected answer from handler")
		return new(core.DataBrokerRes), errors.New(errors.Operational, err)
	}
	stats.MarkMeter("broker.uplink.ok")

	// No response, we stop here and propagate the "no answer".
	// In case of confirmed data, the handler is in charge of creating the confirmation
	if resp == nil || resp.Payload == nil {
		ctx.Debug("No downlink response")
		return new(core.DataBrokerRes), nil
	}
	ctx.Debug("Handle downlink response")

	// If a response was sent, i.e. a downlink data, we need to compute the right MIC
	downlinkPayload, err := core.NewLoRaWANData(resp.Payload, false)
	if err != nil {
		ctx.WithError(err).Debug("Unable to interpret LoRaWAN downlink datagram")
		return new(core.DataBrokerRes), errors.New(errors.Structural, err)
	}
	stats.MarkMeter("broker.downlink.in")
	if err := downlinkPayload.SetMIC(lorawan.AES128Key(mEntry.NwkSKey)); err != nil {
		ctx.WithError(err).Debug("Unable to set MIC")
		return new(core.DataBrokerRes), errors.New(errors.Structural, "Unable to set response MIC")
	}
	resp.Payload.MIC = downlinkPayload.MIC[:]

	// And finally, we acknowledge the answer
	return &core.DataBrokerRes{
		Payload:  resp.Payload,
		Metadata: resp.Metadata,
	}, nil
}
