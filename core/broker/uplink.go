package broker

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sort"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_networkserver "github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/core/broker/application"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

const maxFCntGap = 16384

var (
	ErrNotFound    = errors.New("ttn/broker: Device not found")
	ErrNoMatch     = errors.New("ttn/broker: No matching device")
	ErrInvalidFCnt = errors.New("ttn/broker: Invalid Frame Counter")
)

func (b *broker) HandleUplink(uplink *pb.UplinkMessage) error {
	ctx := b.Ctx.WithField("GatewayEUI", *uplink.GatewayMetadata.GatewayEui)
	var err error
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle uplink")
		}
	}()

	time := time.Now()

	// De-duplicate uplink messages
	duplicates := b.deduplicateUplink(uplink)
	if len(duplicates) == 0 {
		return nil
	}

	base := duplicates[0]

	if base.ProtocolMetadata.GetLorawan() == nil {
		return errors.New("ttn/broker: Can not handle uplink from non-LoRaWAN device")
	}

	// LoRaWAN: Unmarshal
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(base.Payload)
	if err != nil {
		return err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		err = errors.New("Uplink message does not contain a MAC payload.")
		return err
	}

	// Request devices from NS
	devAddr := types.DevAddr(macPayload.FHDR.DevAddr)
	ctx = ctx.WithField("DevAddr", devAddr)
	var getDevicesResp *networkserver.DevicesResponse
	getDevicesResp, err = b.ns.GetDevices(b.Component.GetContext(), &networkserver.DevicesRequest{
		DevAddr: &devAddr,
		FCnt:    macPayload.FHDR.FCnt,
	})
	if err != nil {
		return err
	}
	if len(getDevicesResp.Results) == 0 {
		err = ErrNotFound
		return err
	}
	ctx = ctx.WithField("DevAddrResults", len(getDevicesResp.Results))

	// Find AppEUI/DevEUI through MIC check
	var device *pb_networkserver.DevicesResponse_Device
	for _, candidate := range getDevicesResp.Results {
		nwkSKey := lorawan.AES128Key(*candidate.NwkSKey)
		if candidate.Uses32BitFCnt {
			macPayload.FHDR.FCnt = candidate.FullFCnt
		}
		ok, err = phyPayload.ValidateMIC(nwkSKey)
		if err != nil {
			return err
		}
		if ok {
			device = candidate
			break
		}
	}
	if device == nil {
		err = ErrNoMatch
		return err
	}
	ctx = ctx.WithFields(log.Fields{
		"DevEUI": device.DevEui,
		"AppEUI": device.AppEui,
		"FCnt":   device.FullFCnt,
	})

	if device.DisableFCntCheck {
		// TODO: Add warning to message?
	} else if macPayload.FHDR.FCnt <= device.StoredFCnt || macPayload.FHDR.FCnt-device.StoredFCnt > maxFCntGap {
		// Replay attack or FCnt gap too big
		err = ErrInvalidFCnt
		return err
	}

	// Add FCnt to Metadata (because it's not marshaled in lorawan payload)
	base.ProtocolMetadata.GetLorawan().FCnt = device.FullFCnt

	// Collect GatewayMetadata and DownlinkOptions
	var gatewayMetadata []*gateway.RxMetadata
	var downlinkOptions []*pb.DownlinkOption
	var downlinkMessage *pb.DownlinkMessage
	for _, duplicate := range duplicates {
		gatewayMetadata = append(gatewayMetadata, duplicate.GatewayMetadata)
		downlinkOptions = append(downlinkOptions, duplicate.DownlinkOptions...)
	}

	// Select best DownlinkOption
	if len(downlinkOptions) > 0 {
		downlinkMessage = &pb.DownlinkMessage{
			DownlinkOption: selectBestDownlink(downlinkOptions),
		}
	}

	// Build Uplink
	deduplicatedUplink := &pb.DeduplicatedUplinkMessage{
		Payload:          base.Payload,
		DevEui:           device.DevEui,
		AppEui:           device.AppEui,
		ProtocolMetadata: base.ProtocolMetadata,
		GatewayMetadata:  gatewayMetadata,
		ServerTime:       time.UnixNano(),
		ResponseTemplate: downlinkMessage,
	}

	// Pass Uplink through NS
	deduplicatedUplink, err = b.ns.Uplink(b.Component.GetContext(), deduplicatedUplink)
	if err != nil {
		return err
	}

	var application *application.Application
	application, err = b.applications.Get(*device.AppEui)
	if err != nil {
		return err
	}

	ctx = ctx.WithField("HandlerID", application.HandlerID)

	var handler chan<- *pb.DeduplicatedUplinkMessage
	handler, err = b.getHandler(application.HandlerID)
	if err != nil {
		return err
	}

	ctx.Debug("Forward Uplink")

	handler <- deduplicatedUplink

	return nil
}

func (b *broker) deduplicateUplink(duplicate *pb.UplinkMessage) (uplinks []*pb.UplinkMessage) {
	sum := md5.Sum(duplicate.Payload)
	key := hex.EncodeToString(sum[:])
	list := b.uplinkDeduplicator.Deduplicate(key, duplicate)
	if len(list) == 0 {
		return
	}
	for _, duplicate := range list {
		uplinks = append(uplinks, duplicate.(*pb.UplinkMessage))
	}
	return
}

func selectBestDownlink(options []*pb.DownlinkOption) *pb.DownlinkOption {
	sort.Sort(ByScore(options))
	return options[0]
}
