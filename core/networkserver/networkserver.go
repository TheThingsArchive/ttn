// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"errors"
	"time"

	"gopkg.in/redis.v3"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/fcnt"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/brocaar/lorawan"
)

// NetworkServer implements LoRaWAN-specific functionality for TTN
type NetworkServer interface {
	core.ComponentInterface
	core.ManagementInterface

	UsePrefix(prefixBytes []byte, length int) error
	HandleGetDevices(*pb.DevicesRequest) (*pb.DevicesResponse, error)
	HandlePrepareActivation(*pb_broker.DeduplicatedDeviceActivationRequest) (*pb_broker.DeduplicatedDeviceActivationRequest, error)
	HandleActivate(*pb_handler.DeviceActivationResponse) (*pb_handler.DeviceActivationResponse, error)
	HandleUplink(*pb_broker.DeduplicatedUplinkMessage) (*pb_broker.DeduplicatedUplinkMessage, error)
	HandleDownlink(*pb_broker.DownlinkMessage) (*pb_broker.DownlinkMessage, error)
}

// NewRedisNetworkServer creates a new Redis-backed NetworkServer
func NewRedisNetworkServer(client *redis.Client, netID int) NetworkServer {
	ns := &networkServer{
		devices: device.NewRedisDeviceStore(client),
	}
	ns.netID = [3]byte{byte(netID >> 16), byte(netID >> 8), byte(netID)}
	ns.prefix = [4]byte{ns.netID[2] << 1, 0, 0, 0}
	ns.prefixLength = 7
	return ns
}

type networkServer struct {
	*core.Component
	devices      device.Store
	netID        [3]byte
	prefix       [4]byte
	prefixLength int
}

func (n *networkServer) UsePrefix(prefixBytes []byte, length int) error {
	if length < 7 {
		return errors.New("ttn/networkserver: Invalid prefix length")
	}
	if prefixBytes[0]>>1 != n.netID[2] {
		return errors.New("ttn/networkserver: Invalid prefix")
	}
	copy(n.prefix[:], prefixBytes)
	n.prefixLength = length
	return nil
}

func (n *networkServer) Init(c *core.Component) error {
	n.Component = c
	err := n.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	return nil
}

func (n *networkServer) HandleGetDevices(req *pb.DevicesRequest) (*pb.DevicesResponse, error) {
	devices, err := n.devices.GetWithAddress(*req.DevAddr)
	if err != nil {
		return nil, err
	}

	// Return all devices with DevAddr with FCnt <= fCnt or Security off

	res := &pb.DevicesResponse{
		Results: make([]*pb_lorawan.Device, 0, len(devices)),
	}

	for _, device := range devices {
		fullFCnt := fcnt.GetFull(device.FCntUp, uint16(req.FCnt))
		dev := &pb_lorawan.Device{
			AppEui:           &device.AppEUI,
			AppId:            device.AppID,
			DevEui:           &device.DevEUI,
			NwkSKey:          &device.NwkSKey,
			FCntUp:           device.FCntUp,
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
		} else if device.Options.Uses32BitFCnt && device.FCntUp <= fullFCnt {
			res.Results = append(res.Results, dev)
			continue
		}
	}

	return res, nil
}

func (n *networkServer) HandlePrepareActivation(activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb_broker.DeduplicatedDeviceActivationRequest, error) {
	if activation.AppEui == nil || activation.DevEui == nil {
		return nil, errors.New("ttn/networkserver: Activation missing AppEUI or DevEUI")
	}
	dev, err := n.devices.Get(*activation.AppEui, *activation.DevEui)
	if err != nil {
		return nil, err
	}
	activation.AppId = dev.AppID

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

	// Generate random DevAddr with prefix
	var devAddr types.DevAddr
	copy(devAddr[:], random.New().Bytes(4))
	devAddr = devAddr.WithPrefix(types.DevAddr(n.prefix), n.prefixLength)

	// Set the DevAddr in the Activation Metadata
	lorawanMeta.DevAddr = &devAddr

	// Build JoinAccept Payload
	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			NetID:      n.netID,
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

	// Update FCntUp (from metadata if possible, because only 16lsb are marshaled in FHDR)
	if lorawan := message.GetProtocolMetadata().GetLorawan(); lorawan != nil {
		dev.FCntUp = lorawan.FCnt
	} else {
		dev.FCntUp = macPayload.FHDR.FCnt
	}
	dev.LastSeen = time.Now().UTC()
	err = n.devices.Set(dev, "f_cnt_up")
	if err != nil {
		return nil, err
	}

	// Prepare Downlink
	if message.ResponseTemplate == nil {
		return message, nil
	}
	message.ResponseTemplate.AppEui = message.AppEui
	message.ResponseTemplate.DevEui = message.DevEui

	// Add Full FCnt (avoiding nil pointer panics)
	if option := message.ResponseTemplate.DownlinkOption; option != nil {
		if protocol := option.ProtocolConfig; protocol != nil {
			if lorawan := protocol.GetLorawan(); lorawan != nil {
				lorawan.FCnt = dev.FCntDown
			}
		}
	}

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
