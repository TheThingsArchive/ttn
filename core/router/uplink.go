// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

func (r *router) HandleUplink(gatewayID string, uplink *pb.UplinkMessage) (err error) {
	ctx := r.Ctx.WithField("GatewayID", gatewayID).WithFields(fields.Get(uplink))
	start := time.Now()
	defer func() {
		if err != nil {
			uplink.Trace = uplink.Trace.WithEvent(trace.DropEvent, "reason", err)
			ctx.WithError(err).Warn("Could not handle uplink")
		}
	}()
	r.status.uplink.Mark(1)

	uplink.Trace = uplink.Trace.WithEvent(trace.ReceiveEvent, "gateway", gatewayID)

	// LoRaWAN: Unmarshal
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(uplink.Payload)
	if err != nil {
		return err
	}

	if phyPayload.MHDR.MType == lorawan.JoinRequest {
		joinRequestPayload, ok := phyPayload.MACPayload.(*lorawan.JoinRequestPayload)
		if !ok {
			return errors.NewErrInvalidArgument("Join Request", "does not contain a JoinRequest payload")
		}
		devEUI := types.DevEUI(joinRequestPayload.DevEUI)
		appEUI := types.AppEUI(joinRequestPayload.AppEUI)
		ctx.WithFields(ttnlog.Fields{
			"DevEUI": devEUI,
			"AppEUI": appEUI,
		}).Debug("Handle Uplink as Activation")
		r.HandleActivation(gatewayID, &pb.DeviceActivationRequest{
			Payload:          uplink.Payload,
			DevEui:           &devEUI,
			AppEui:           &appEUI,
			ProtocolMetadata: uplink.ProtocolMetadata,
			GatewayMetadata:  uplink.GatewayMetadata,
			Trace:            uplink.Trace.WithEvent("handle uplink as activation"),
		})
		return nil
	}

	if phyPayload.MHDR.MType != lorawan.UnconfirmedDataUp && phyPayload.MHDR.MType != lorawan.ConfirmedDataUp {
		ctx.Warn("Accidentally received non-uplink message")
		return nil
	}

	if lorawan := uplink.ProtocolMetadata.GetLorawan(); lorawan != nil {
		ctx = ctx.WithField("Modulation", lorawan.Modulation.String())
		if lorawan.Modulation == pb_lorawan.Modulation_LORA {
			ctx = ctx.WithField("DataRate", lorawan.DataRate)
		} else {
			ctx = ctx.WithField("BitRate", lorawan.BitRate)
		}
	}

	if gateway := uplink.GatewayMetadata; gateway != nil {
		ctx = ctx.WithFields(ttnlog.Fields{
			"Frequency": gateway.Frequency,
			"RSSI":      gateway.Rssi,
			"SNR":       gateway.Snr,
		})
	}

	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return errors.NewErrInvalidArgument("Uplink", "does not contain a MAC payload")
	}
	devAddr := types.DevAddr(macPayload.FHDR.DevAddr)

	ctx = ctx.WithFields(ttnlog.Fields{
		"DevAddr": devAddr,
		"FCnt":    macPayload.FHDR.FCnt,
	})

	gateway := r.getGateway(gatewayID)

	if err = gateway.HandleUplink(uplink); err != nil {
		return err
	}

	var downlinkOptions []*pb_broker.DownlinkOption
	if gateway.Schedule.IsActive() {
		downlinkOptions = r.buildDownlinkOptions(uplink, false, gateway)
		uplink.Trace = uplink.Trace.WithEvent(trace.BuildDownlinkEvent,
			"options", len(downlinkOptions),
		)
	}

	ctx = ctx.WithField("DownlinkOptions", len(downlinkOptions))

	// Find Broker
	brokers, err := r.Discovery.GetAllBrokersForDevAddr(devAddr)
	if err != nil {
		return err
	}

	if len(brokers) == 0 {
		ctx.Debug("No brokers to forward message to")
		uplink.Trace = uplink.Trace.WithEvent(trace.DropEvent, "reason", "no brokers")
		return nil
	}

	ctx = ctx.WithField("NumBrokers", len(brokers))

	uplink.Trace = uplink.Trace.WithEvent(trace.ForwardEvent,
		"brokers", len(brokers),
	)

	// Forward to all brokers
	for _, broker := range brokers {
		broker, err := r.getBroker(broker)
		if err != nil {
			continue
		}
		broker.uplink <- &pb_broker.UplinkMessage{
			Payload:          uplink.Payload,
			ProtocolMetadata: uplink.ProtocolMetadata,
			GatewayMetadata:  uplink.GatewayMetadata,
			DownlinkOptions:  downlinkOptions,
			Trace:            uplink.Trace,
		}
	}

	ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled uplink")

	return nil
}
