package router

import (
	"errors"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/brocaar/lorawan"
)

func (r *router) HandleUplink(gatewayEUI types.GatewayEUI, uplink *pb.UplinkMessage) error {

	gateway := r.getGateway(gatewayEUI)
	gateway.Schedule.Sync(uplink.GatewayMetadata.Timestamp)
	gateway.Utilization.AddRx(uplink)

	downlinkOptions := r.buildDownlinkOptions(uplink, false, gateway)

	// LoRaWAN: Unmarshal
	var phyPayload lorawan.PHYPayload
	err := phyPayload.UnmarshalBinary(uplink.Payload)
	if err != nil {
		return err
	}

	switch phyPayload.MHDR.MType {
	case lorawan.JoinRequest:
		joinRequestPayload, ok := phyPayload.MACPayload.(*lorawan.JoinRequestPayload)
		if !ok {
			return errors.New("Join Request message does not contain a join payload.")
		}
		devEUI := types.DevEUI(joinRequestPayload.DevEUI)
		appEUI := types.AppEUI(joinRequestPayload.AppEUI)
		_, err := r.HandleActivation(gatewayEUI, &pb.DeviceActivationRequest{
			Payload:          uplink.Payload,
			DevEui:           &devEUI,
			AppEui:           &appEUI,
			ProtocolMetadata: uplink.ProtocolMetadata,
			GatewayMetadata:  uplink.GatewayMetadata,
		})
		return err
	case lorawan.UnconfirmedDataUp, lorawan.ConfirmedDataUp:
		// Just continue handling uplink
	default:
		return errors.New("ttn/router: Unhandled message type")
	}

	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return errors.New("Uplink message does not contain a MAC payload.")
	}
	devAddr := types.DevAddr(macPayload.FHDR.DevAddr)

	// Find Broker
	brokers, err := r.brokerDiscovery.Discover(devAddr)
	if err != nil {
		return err
	}

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
