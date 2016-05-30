package networkserver

import (
	"errors"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/core/fcnt"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/brocaar/lorawan"
)

// NetworkServer implements LoRaWAN-specific functionality for TTN
type NetworkServer interface {
	HandleGetDevices(*pb.DevicesRequest) (*pb.DevicesResponse, error)
	HandlePrepareActivation(*pb_broker.DeduplicatedDeviceActivationRequest) (*pb_broker.DeduplicatedDeviceActivationRequest, error)
	HandleActivate(*pb_handler.DeviceActivationResponse) (*pb_handler.DeviceActivationResponse, error)
	HandleUplink(*pb_broker.DeduplicatedUplinkMessage) (*pb_broker.DeduplicatedUplinkMessage, error)
	HandleDownlink(*pb_broker.DownlinkMessage) (*pb_broker.DownlinkMessage, error)
}

type networkServer struct {
	devices device.Store
}

func (n *networkServer) HandleGetDevices(req *pb.DevicesRequest) (*pb.DevicesResponse, error) {
	devices, err := n.devices.GetWithAddress(*req.DevAddr)
	if err != nil {
		return nil, err
	}

	// Return all devices with DevAddr with FCnt <= fCnt or Security off

	res := &pb.DevicesResponse{
		Results: make([]*pb.DevicesResponse_Device, 0, len(devices)),
	}

	for _, device := range devices {
		dev := &pb.DevicesResponse_Device{
			AppEui:           &device.AppEUI,
			DevEui:           &device.DevEUI,
			NwkSKey:          &device.NwkSKey,
			FCnt:             device.FCntUp,
			Uses32BitFCnt:    device.Options.Uses32BitFCnt,
			DisableFCntCheck: device.Options.DisableFCntCheck,
		}
		if device.Options.DisableFCntCheck {
			res.Results = append(res.Results, dev)
			continue
		}
		if device.FCntUp <= req.FCnt {
			res.Results = append(res.Results, dev)
			continue
		} else if device.Options.Uses32BitFCnt && device.FCntUp <= fcnt.GetFull(device.FCntUp, uint16(req.FCnt)) {
			res.Results = append(res.Results, dev)
			continue
		}
	}

	return res, nil
}

var netID = [3]byte{0x00, 0x00, 0x13}
var nwkID byte = 0x13

func (n *networkServer) HandlePrepareActivation(activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb_broker.DeduplicatedDeviceActivationRequest, error) {
	// Build activation metadata if not present
	if meta := activation.GetActivationMetadata(); meta == nil {
		activation.ActivationMetadata = &pb_protocol.ActivationMetadata{}
	}
	// Build lorawan metadata if not present
	if lorawan := activation.ActivationMetadata.GetLorawan(); lorawan == nil {
		return nil, errors.New("ttn/networkserver: Can only handle LoRaWAN activations")
	}

	// Build response template if not present
	if pld := activation.GetResponseTemplate(); pld == nil {
		return nil, errors.New("ttn/networkserver: Activation does not contain a response template")
	}
	lorawanMeta := activation.ActivationMetadata.GetLorawan()

	// Generate random DevAddr
	// TODO: Be smarter than just randomly generating addresses.
	var devAddr types.DevAddr
	copy(devAddr[:], random.Bytes(4))
	devAddr[0] = (nwkID << 1) | (devAddr[0] & 1) // DevAddr 7 msb are NetID 7 lsb

	// Set the DevAddr in the Activation Metadata
	lorawanMeta.DevAddr = &devAddr

	// Build JoinAccept Payload
	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			NetID:      netID,
			DLSettings: lorawan.DLSettings{RX2DataRate: uint8(lorawanMeta.Rx2Dr), RX1DROffset: uint8(lorawanMeta.Rx1DrOffset)},
			RXDelay:    uint8(lorawanMeta.RxDelay),
			DevAddr:    lorawan.DevAddr(devAddr),
		},
	}
	if len(lorawanMeta.CfList) == 5 {
		var cfList lorawan.CFList
		for i, cfListItem := range lorawanMeta.CfList {
			cfList[i] = uint32(cfListItem)
		}
		phy.MACPayload.(*lorawan.JoinAcceptPayload).CFList = &cfList
	}

	// Set the Payload
	phyBytes, err := phy.MarshalBinary()
	if err != nil {
		return nil, err
	}
	activation.ResponseTemplate.Payload = phyBytes

	return activation, nil
}

func (n *networkServer) HandleActivate(activation *pb_handler.DeviceActivationResponse) (*pb_handler.DeviceActivationResponse, error) {
	meta := activation.GetActivationMetadata()
	if meta == nil {
		return nil, errors.New("ttn/networkserver: invalid ActivationMetadata")
	}
	lorawan := meta.GetLorawan()
	if lorawan == nil {
		return nil, errors.New("ttn/networkserver: invalid LoRaWAN ActivationMetadata")
	}
	err := n.devices.Activate(*lorawan.AppEui, *lorawan.DevEui, *lorawan.DevAddr, *lorawan.NwkSKey)
	if err != nil {
		return nil, err
	}
	return activation, nil
}

func (n *networkServer) HandleUplink(message *pb_broker.DeduplicatedUplinkMessage) (*pb_broker.DeduplicatedUplinkMessage, error) {
	// Get Device
	dev, err := n.devices.Get(*message.AppEui, *message.DevEui)
	if err != nil {
		return nil, err
	}

	// Unmarshal LoRaWAN Payload
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(message.Payload)
	if err != nil {
		return nil, err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.New("ttn/networkserver: LoRaWAN message does not contain a MACPayload")
	}

	// Update FCntUp
	dev.FCntUp = macPayload.FHDR.FCnt
	dev.LastSeen = time.Now().UTC()
	err = n.devices.Set(dev, "f_cnt_up")
	if err != nil {
		return nil, err
	}

	// Prepare Downlink
	if message.ResponseTemplate == nil {
		message.ResponseTemplate = &pb_broker.DownlinkMessage{}
	}
	message.ResponseTemplate.AppEui = message.AppEui
	message.ResponseTemplate.DevEui = message.DevEui

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.UnconfirmedDataDown,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: macPayload.FHDR.DevAddr,
				FCtrl: lorawan.FCtrl{
					ACK: phyPayload.MHDR.MType == lorawan.ConfirmedDataUp,
				},
				FCnt: dev.FCntDown,
			},
			FPort: macPayload.FPort,
		},
	}
	phyBytes, err := phy.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// TODO: Maybe we need to add MAC commands on downlink

	message.ResponseTemplate.Payload = phyBytes

	return message, nil
}

func (n *networkServer) HandleDownlink(message *pb_broker.DownlinkMessage) (*pb_broker.DownlinkMessage, error) {
	// Get Device
	dev, err := n.devices.Get(*message.AppEui, *message.DevEui)
	if err != nil {
		return nil, err
	}

	// Unmarshal LoRaWAN Payload
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(message.Payload)
	if err != nil {
		return nil, err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.New("ttn/networkserver: LoRaWAN message does not contain a MACPayload")
	}

	// Set DevAddr
	macPayload.FHDR.DevAddr = lorawan.DevAddr(dev.DevAddr)

	// FIRST set and THEN increment FCntDown
	// TODO: For confirmed downlink, FCntDown should be incremented AFTER ACK
	macPayload.FHDR.FCnt = dev.FCntDown
	dev.FCntDown++
	err = n.devices.Set(dev, "f_cnt_down")
	if err != nil {
		return nil, err
	}

	// TODO: Maybe we need to add MAC commands on downlink

	// Sign MIC
	phyPayload.SetMIC(lorawan.AES128Key(dev.NwkSKey))

	// Update message
	bytes, err := phyPayload.MarshalBinary()
	if err != nil {
		return nil, err
	}
	message.Payload = bytes

	return message, nil
}
