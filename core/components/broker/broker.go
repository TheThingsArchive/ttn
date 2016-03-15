// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// component implements the core.Component interface
type component struct {
	NetworkController
	ctx log.Interface
}

// New construct a new Broker component
func New(controller NetworkController, ctx log.Interface) core.BrokerServer {
	return component{NetworkController: controller, ctx: ctx}
}

// HandleData implements the core.RouterClient interface
func (b component) HandleData(bctx context.Context, req *core.DataBrokerReq) (*core.DataBrokerRes, error) {
	// Get some logs / analytics
	stats.MarkMeter("broker.uplink.in")
	b.ctx.Debug("Handling uplink packet")

	// Validate incoming data
	uplinkPayload, err := core.NewLoRaWANData(req.Payload, true)
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	devAddr := req.Payload.MACPayload.FHDR.DevAddr // No nil ref, ensured by NewLoRaWANData()
	ctx := b.ctx.WithField("DevAddr", devAddr)

	// Check whether we should handle it
	entries, err := b.LookupDevices(devAddr)
	if err != nil {
		switch err.(errors.Failure).Nature {
		case errors.NotFound:
			stats.MarkMeter("broker.uplink.handler_lookup.device_not_found")
			ctx.Debug("Uplink device not found")
		default:
			b.ctx.Warn("Database lookup failed")
		}
		return nil, err
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
		ok, err := uplinkPayload.ValidateMIC(key)
		if err != nil {
			continue
		}

		if !ok && entry.FCntUp > 65535 { // Check with 32-bits counter
			fcnt32, err := b.WholeCounter(fcnt16, entry.FCntUp)
			if err != nil {
				continue
			}
			fhdr.FCnt = fcnt32
			ok, err = uplinkPayload.ValidateMIC(key)
		}

		if ok {
			mEntry = &entry
			stats.MarkMeter("broker.uplink.handler_lookup.mic_match")
			ctx.WithField("handler", entry.HandlerNet).Debug("MIC check succeeded")
			break // We stop at the first valid check ...
		}
	}

	if mEntry == nil {
		stats.MarkMeter("broker.uplink.handler_lookup.no_mic_match")
		err := errors.New(errors.NotFound, "MIC check returned no matches")
		ctx.Debug(err.Error())
		return nil, err
	}

	// It does matter here to use the DevEUI from the entry and not from the packet.
	// The packet actually holds a DevAddr and the real DevEUI has been determined thanks
	// to the MIC check + persistence
	if err := b.UpdateFCnt(mEntry.AppEUI, mEntry.DevEUI, fhdr.FCnt); err != nil {
		return nil, err
	}

	// Then we forward the packet to the handler and wait for the response
	conn, err := grpc.Dial(mEntry.HandlerNet, grpc.WithInsecure(), grpc.WithTimeout(time.Second*2))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	handler := core.NewHandlerClient(conn)
	resp, err := handler.HandleData(context.Background(), &core.DataHandlerReq{
		Payload:  req.Payload.MACPayload.FRMPayload,
		DevEUI:   mEntry.DevEUI,
		AppEUI:   mEntry.AppEUI,
		FCnt:     fhdr.FCnt,
		MType:    req.Payload.MHDR.MType,
		Metadata: req.Metadata,
	})

	if err != nil && !strings.Contains(err.Error(), string(errors.NotFound)) { // FIXME Find better way to analyze error
		stats.MarkMeter("broker.uplink.bad_handler_response")
		return nil, errors.New(errors.Operational, err)
	}
	stats.MarkMeter("broker.uplink.ok")

	// No response, we stop here and propagate the "no answer".
	// In case of confirmed data, the handler is in charge of creating the confirmation
	if resp == nil {
		return nil, nil
	}

	// If a response was sent, i.e. a downlink data, we need to compute the right MIC
	downlinkPayload, err := core.NewLoRaWANData(resp.Payload, false)
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	stats.MarkMeter("broker.downlink.in")
	if err := downlinkPayload.SetMIC(lorawan.AES128Key(mEntry.NwkSKey)); err != nil {
		return nil, errors.New(errors.Structural, "Unable to set response MIC")
	}
	resp.Payload.MIC = downlinkPayload.MIC[:]

	// And finally, we acknowledge the answer
	return &core.DataBrokerRes{
		Payload:  resp.Payload,
		Metadata: resp.Metadata,
	}, nil
}

// Register implements the core.BrokerServer interface
func (b component) SubscribePersonalized(bctx context.Context, req *core.SubBrokerReq) (*core.SubBrokerRes, error) {
	b.ctx.Debug("Handling personalized subscription")

	// Ensure the entry is valid
	if len(req.AppEUI) != 8 {
		return nil, errors.New(errors.Structural, "Invalid Application EUI")
	}

	if len(req.DevEUI) != 8 {
		return nil, errors.New(errors.Structural, "Invalid Device EUI")
	}

	var nwkSKey [16]byte
	if len(req.NwkSKey) != 16 {
		return nil, errors.New(errors.Structural, "Invalid Network Session Key")
	}
	copy(nwkSKey[:], req.NwkSKey)

	re := regexp.MustCompile("^[-\\w]+\\.([-\\w]+\\.?)+:\\d+$")
	if !re.MatchString(req.HandlerNet) {
		return nil, errors.New(errors.Structural, fmt.Sprintf("Invalid Handler Net Address. Should match: %s", re))
	}

	return nil, b.StoreDevice(devEntry{
		HandlerNet: req.HandlerNet,
		AppEUI:     req.AppEUI,
		DevEUI:     req.DevEUI,
		NwkSKey:    nwkSKey,
		FCntUp:     0,
	})
}
