// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sort"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/fcnt"
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
		err = errors.New("ttn/broker: Can not handle uplink from non-LoRaWAN device")
		return err
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
	var device *pb_lorawan.Device
	for _, candidate := range getDevicesResp.Results {
		nwkSKey := lorawan.AES128Key(*candidate.NwkSKey)
		if candidate.Uses32BitFCnt {
			macPayload.FHDR.FCnt = fcnt.GetFull(macPayload.FHDR.FCnt, uint16(candidate.FCntUp))
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
		"AppID":  device.AppId,
	})

	if device.DisableFCntCheck {
		// TODO: Add warning to message?
	} else if device.FCntUp == 0 {

	} else if macPayload.FHDR.FCnt <= device.FCntUp || macPayload.FHDR.FCnt-device.FCntUp > maxFCntGap {
		// Replay attack or FCnt gap too big
		err = ErrInvalidFCnt
		return err
	}

	// Add FCnt to Metadata (because it's not marshaled in lorawan payload)
	base.ProtocolMetadata.GetLorawan().FCnt = macPayload.FHDR.FCnt

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

	var announcements []*pb_discovery.Announcement
	announcements, err = b.handlerDiscovery.ForAppID(device.AppId)
	if err != nil {
		return err
	}
	if len(announcements) == 0 {
		err = errors.New("ttn/broker: No Handlers")
		return err
	}
	if len(announcements) > 1 {
		err = errors.New("ttn/broker: Can't forward to multiple Handlers")
		return err
	}

	var handler chan<- *pb.DeduplicatedUplinkMessage
	handler, err = b.getHandler(announcements[0].Id)
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
