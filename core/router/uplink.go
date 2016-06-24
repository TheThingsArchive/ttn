// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"errors"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

func (r *router) HandleUplink(gatewayEUI types.GatewayEUI, uplink *pb.UplinkMessage) error {
	ctx := r.Ctx.WithField("GatewayEUI", gatewayEUI)
	var err error
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle uplink")
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
			return errors.New("Join Request message does not contain a join payload.")
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

	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return errors.New("Uplink message does not contain a MAC payload.")
	}
	devAddr := types.DevAddr(macPayload.FHDR.DevAddr)

	ctx = ctx.WithField("DevAddr", devAddr)

	gateway := r.getGateway(gatewayEUI)
	gateway.LastSeen = time.Now()
	gateway.Schedule.Sync(uplink.GatewayMetadata.Timestamp)
	gateway.Utilization.AddRx(uplink)

	downlinkOptions := r.buildDownlinkOptions(uplink, false, gateway)

	ctx = ctx.WithField("DownlinkOptions", len(downlinkOptions))

	// Find Broker
	brokers, err := r.brokerDiscovery.Discover(devAddr)
	if err != nil {
		return err
	}

	ctx = ctx.WithField("NumBrokers", len(brokers))
	ctx.Debug("Forward Uplink")

	// Forward to all brokers
	for _, broker := range brokers {
		broker, err := r.getBroker(broker)
		if err != nil {
			continue
		}
		broker.association.Send(&pb_broker.UplinkMessage{
			Payload:          uplink.Payload,
			ProtocolMetadata: uplink.ProtocolMetadata,
			GatewayMetadata:  uplink.GatewayMetadata,
			DownlinkOptions:  downlinkOptions,
		})
	}

	return nil
}
