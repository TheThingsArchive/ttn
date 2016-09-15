// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

func (r *router) HandleUplink(gatewayEUI types.GatewayEUI, uplink *pb.UplinkMessage) error {
	ctx := r.Ctx.WithField("GatewayEUI", gatewayEUI)
	var err error
	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle uplink")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled uplink")
		}
	}()

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
		ctx.WithFields(log.Fields{
			"DevEUI": devEUI,
			"AppEUI": appEUI,
		}).Debug("Handle Uplink as Activation")
		_, err := r.HandleActivation(gatewayEUI, &pb.DeviceActivationRequest{
			Payload:          uplink.Payload,
			DevEui:           &devEUI,
			AppEui:           &appEUI,
			ProtocolMetadata: uplink.ProtocolMetadata,
			GatewayMetadata:  uplink.GatewayMetadata,
		})
		return err
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
		ctx = ctx.WithFields(log.Fields{
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

	ctx = ctx.WithField("DevAddr", devAddr)

	gateway := r.getGateway(gatewayEUI)
	gateway.LastSeen = time.Now()
	gateway.Schedule.Sync(uplink.GatewayMetadata.Timestamp)
	gateway.Utilization.AddRx(uplink)

	var downlinkOptions []*pb_broker.DownlinkOption
	if gateway.Schedule.IsActive() {
		downlinkOptions = r.buildDownlinkOptions(uplink, false, gateway)
	}

	ctx = ctx.WithField("DownlinkOptions", len(downlinkOptions))

	// Find Broker
	brokers, err := r.Discovery.GetAllBrokersForDevAddr(devAddr)
	if err != nil {
		return err
	}

	ctx = ctx.WithField("NumBrokers", len(brokers))

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
		}
	}

	return nil
}
