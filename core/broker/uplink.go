// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/fcnt"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

const maxFCntGap = 16384

func (b *broker) HandleUplink(uplink *pb.UplinkMessage) error {
	ctx := b.Ctx.WithField("GatewayID", uplink.GatewayMetadata.GatewayId)
	var err error
	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle uplink")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled uplink")
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
		err = errors.NewErrInvalidArgument("Uplink", "does not contain LoRaWAN metadata")
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
		err = errors.NewErrInvalidArgument("Uplink", "does not contain a MAC payload")
		return err
	}

	// Request devices from NS
	devAddr := types.DevAddr(macPayload.FHDR.DevAddr)
	ctx = ctx.WithField("DevAddr", devAddr)
	var getDevicesResp *networkserver.DevicesResponse
	getDevicesResp, err = b.ns.GetDevices(b.Component.GetContext(b.nsToken), &networkserver.DevicesRequest{
		DevAddr: &devAddr,
		FCnt:    macPayload.FHDR.FCnt,
	})
	if err != nil {
		return errors.BuildGRPCError(errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not return devices"))
	}
	if len(getDevicesResp.Results) == 0 {
		err = errors.NewErrNotFound(fmt.Sprintf("Device with DevAddr %s and FCnt <= %d", devAddr, macPayload.FHDR.FCnt))
		return err
	}
	ctx = ctx.WithField("DevAddrResults", len(getDevicesResp.Results))

	// Sort by FCntUp to optimize the number of MIC checks
	sort.Sort(ByFCntUp(getDevicesResp.Results))

	// Find AppEUI/DevEUI through MIC check
	var device *pb_lorawan.Device
	var micChecks int
	for _, candidate := range getDevicesResp.Results {
		nwkSKey := lorawan.AES128Key(*candidate.NwkSKey)
		if candidate.Uses32BitFCnt {
			macPayload.FHDR.FCnt = fcnt.GetFull(candidate.FCntUp, uint16(macPayload.FHDR.FCnt))
		}
		micChecks++
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
		err = errors.NewErrNotFound("device that validates MIC")
		return err
	}
	ctx = ctx.WithFields(log.Fields{
		"MICChecks": micChecks,
		"DevEUI":    device.DevEui,
		"AppEUI":    device.AppEui,
		"AppID":     device.AppId,
		"DevID":     device.DevId,
	})

	if device.DisableFCntCheck {
		// TODO: Add warning to message?
	} else if device.FCntUp == 0 {

	} else if macPayload.FHDR.FCnt <= device.FCntUp || macPayload.FHDR.FCnt-device.FCntUp > maxFCntGap {
		// Replay attack or FCnt gap too big
		err = errors.NewErrNotFound("device with matching FCnt")
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
		DevId:            device.DevId,
		AppEui:           device.AppEui,
		AppId:            device.AppId,
		ProtocolMetadata: base.ProtocolMetadata,
		GatewayMetadata:  gatewayMetadata,
		ServerTime:       time.UnixNano(),
		ResponseTemplate: downlinkMessage,
	}

	// Pass Uplink through NS
	deduplicatedUplink, err = b.ns.Uplink(b.Component.GetContext(b.nsToken), deduplicatedUplink)
	if err != nil {
		return errors.BuildGRPCError(errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not handle uplink"))
	}

	var announcements []*pb_discovery.Announcement
	announcements, err = b.Discovery.GetAllHandlersForAppID(device.AppId)
	if err != nil {
		return err
	}
	if len(announcements) == 0 {
		err = errors.NewErrNotFound(fmt.Sprintf("Handler for AppID %s", device.AppId))
		return err
	}
	if len(announcements) > 1 {
		err = errors.NewErrInternal(fmt.Sprintf("Multiple Handlers for AppID %s", device.AppId))
		return err
	}

	var handler chan<- *pb.DeduplicatedUplinkMessage
	handler, err = b.getHandler(announcements[0].Id)
	if err != nil {
		return err
	}

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

// ByFCntUp implements sort.Interface for []*pb_lorawan.Device based on FCnt
type ByFCntUp []*pb_lorawan.Device

func (a ByFCntUp) Len() int      { return len(a) }
func (a ByFCntUp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByFCntUp) Less(i, j int) bool {
	// Devices that disable the FCnt check have low priority
	if a[i].DisableFCntCheck {
		return 2*int(a[i].FCntUp)+100 < int(a[j].FCntUp)
	}
	return int(a[i].FCntUp) < int(a[j].FCntUp)
}
