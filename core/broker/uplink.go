// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/fields"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/fcnt"
	"github.com/brocaar/lorawan"
)

const maxFCntGap = 16384

func (b *broker) HandleUplink(uplink *pb.UplinkMessage) (err error) {
	ctx := b.Ctx.WithFields(fields.Get(uplink))
	start := time.Now()
	deduplicatedUplink := new(pb.DeduplicatedUplinkMessage)
	deduplicatedUplink.ServerTime = start.UnixNano()
	defer func() {
		if err != nil {
			if deduplicatedUplink != nil {
				deduplicatedUplink.Trace = deduplicatedUplink.Trace.WithEvent(trace.DropEvent, "reason", err)
			}
			ctx.WithError(err).Warn("Could not handle uplink")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled uplink")
		}
		if deduplicatedUplink != nil && b.monitorStream != nil {
			b.monitorStream.Send(deduplicatedUplink)
		}
	}()

	b.status.uplink.Mark(1)

	uplink.Trace = uplink.Trace.WithEvent(trace.ReceiveEvent)

	// De-duplicate uplink messages
	duplicates := b.deduplicateUplink(uplink)
	if len(duplicates) == 0 {
		return nil
	}
	ctx = ctx.WithField("Duplicates", len(duplicates))

	b.status.uplinkUnique.Mark(1)

	deduplicatedUplink.Payload = duplicates[0].Payload
	deduplicatedUplink.ProtocolMetadata = duplicates[0].ProtocolMetadata
	deduplicatedUplink.Trace = deduplicatedUplink.Trace.WithEvent(trace.DeduplicateEvent,
		"duplicates", len(duplicates),
	)
	for _, duplicate := range duplicates {
		if duplicate.Trace != nil {
			deduplicatedUplink.Trace.Parents = append(deduplicatedUplink.Trace.Parents, duplicate.Trace)
		}
	}

	if deduplicatedUplink.ProtocolMetadata.GetLorawan() == nil {
		return errors.NewErrInvalidArgument("Uplink", "does not contain LoRaWAN metadata")
	}

	// LoRaWAN: Unmarshal
	var phyPayload lorawan.PHYPayload
	err = phyPayload.UnmarshalBinary(deduplicatedUplink.Payload)
	if err != nil {
		return err
	}
	macPayload, ok := phyPayload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return errors.NewErrInvalidArgument("Uplink", "does not contain a MAC payload")
	}

	// Request devices from NS
	devAddr := types.DevAddr(macPayload.FHDR.DevAddr)
	ctx = ctx.WithFields(ttnlog.Fields{
		"DevAddr": devAddr,
		"FCnt":    macPayload.FHDR.FCnt,
	})
	var getDevicesResp *networkserver.DevicesResponse
	getDevicesResp, err = b.ns.GetDevices(b.Component.GetContext(b.nsToken), &networkserver.DevicesRequest{
		DevAddr: &devAddr,
		FCnt:    macPayload.FHDR.FCnt,
	})
	if err != nil {
		return errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not return devices")
	}
	b.status.deduplication.Update(int64(len(getDevicesResp.Results)))
	if len(getDevicesResp.Results) == 0 {
		return errors.NewErrNotFound(fmt.Sprintf("Device with DevAddr %s and FCnt <= %d", devAddr, macPayload.FHDR.FCnt))
	}
	ctx = ctx.WithField("DevAddrResults", len(getDevicesResp.Results))
	deduplicatedUplink.Trace = deduplicatedUplink.Trace.WithEvent("got devices from networkserver",
		"devices", len(getDevicesResp.Results),
	)

	// Sort by FCntUp to optimize the number of MIC checks
	sort.Sort(ByFCntUp(getDevicesResp.Results))

	// Find AppEUI/DevEUI through MIC check
	var device *pb_lorawan.Device
	var micChecks int
	originalFCnt := macPayload.FHDR.FCnt
	for _, candidate := range getDevicesResp.Results {
		nwkSKey := lorawan.AES128Key(*candidate.NwkSKey)

		// First check with the 16 bit counter
		micChecks++
		ok, err = phyPayload.ValidateMIC(nwkSKey)
		if err != nil {
			return err
		}
		if ok {
			device = candidate
			break
		}

		if fullFCnt := fcnt.GetFull(candidate.FCntUp, uint16(originalFCnt)); fullFCnt != originalFCnt && candidate.Uses32BitFCnt {
			macPayload.FHDR.FCnt = fullFCnt

			// Then check again with the 32 bit counter
			if macPayload.FHDR.FCnt != originalFCnt {
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

			macPayload.FHDR.FCnt = originalFCnt
		}
	}
	if device == nil {
		return errors.NewErrNotFound("device that validates MIC")
	}

	ctx = ctx.WithFields(ttnlog.Fields{
		"MICChecks": micChecks,
		"DevEUI":    device.DevEui,
		"AppEUI":    device.AppEui,
		"AppID":     device.AppId,
		"DevID":     device.DevId,
		"FCnt":      originalFCnt,
	})
	deduplicatedUplink.DevEui = device.DevEui
	deduplicatedUplink.AppEui = device.AppEui
	deduplicatedUplink.AppId = device.AppId
	deduplicatedUplink.DevId = device.DevId
	deduplicatedUplink.Trace = deduplicatedUplink.Trace.WithEvent(trace.CheckMICEvent, "mic checks", micChecks)
	if macPayload.FHDR.FCnt != originalFCnt {
		ctx = ctx.WithField("RealFCnt", macPayload.FHDR.FCnt)
	}

	switch {
	case macPayload.FHDR.FCnt > device.FCntUp && macPayload.FHDR.FCnt-device.FCntUp <= maxFCntGap:
		// FCnt higher than latest and within max FCnt gap (normal case)
	case device.DisableFCntCheck:
		// FCnt Check disabled. Rely on MIC check only
	case device.FCntUp == 0:
		// FCntUp is reset. We don't know where the device will start sending.
	case macPayload.FHDR.FCnt == device.FCntUp:
		if phyPayload.MHDR.MType == lorawan.ConfirmedDataUp {
			// Retry of confirmed uplink
			break
		}
		fallthrough
	case macPayload.FHDR.FCnt <= device.FCntUp:
		return errors.NewErrInvalidArgument("FCnt", "not high enough")
	case macPayload.FHDR.FCnt-device.FCntUp > maxFCntGap:
		return errors.NewErrInvalidArgument("FCnt", "too high")
	default:
		return errors.NewErrInternal("FCnt check failed")
	}

	// Add FCnt to Metadata (because it's not marshaled in lorawan payload)
	deduplicatedUplink.ProtocolMetadata.GetLorawan().FCnt = macPayload.FHDR.FCnt

	// Collect GatewayMetadata and DownlinkOptions
	var downlinkOptions []downlinkOption
	for _, duplicate := range duplicates {
		deduplicatedUplink.GatewayMetadata = append(deduplicatedUplink.GatewayMetadata, duplicate.GatewayMetadata)
		for _, option := range duplicate.DownlinkOptions {
			if option.RxDelay != 1 {
				continue // The Downlink needs to have an RX delay of 1 seconds
			}
			downlinkOptions = append(downlinkOptions, downlinkOption{
				uplinkMetadata: duplicate.GatewayMetadata,
				option:         option,
			})
		}
	}

	// Select best DownlinkOption
	if len(downlinkOptions) > 0 {
		deduplicatedUplink.ResponseTemplate = &pb.DownlinkMessage{
			DevEui:         device.DevEui,
			AppEui:         device.AppEui,
			AppId:          device.AppId,
			DevId:          device.DevId,
			DownlinkOption: selectBestDownlink(downlinkOptions),
		}
	}

	// Pass Uplink through NS
	deduplicatedUplink, err = b.ns.Uplink(b.Component.GetContext(b.nsToken), deduplicatedUplink)
	if err != nil {
		return errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not handle uplink")
	}

	var announcements []*pb_discovery.Announcement
	announcements, err = b.Discovery.GetAllHandlersForAppID(device.AppId)
	if err != nil {
		return err
	}
	if len(announcements) == 0 {
		return errors.NewErrNotFound(fmt.Sprintf("Handler for AppID %s", device.AppId))
	}
	if len(announcements) > 1 {
		return errors.NewErrInternal(fmt.Sprintf("Multiple Handlers for AppID %s", device.AppId))
	}

	var handler chan<- *pb.DeduplicatedUplinkMessage
	handler, err = b.getHandlerUplink(announcements[0].Id)
	if err != nil {
		return err
	}

	deduplicatedUplink.Trace = deduplicatedUplink.Trace.WithEvent(trace.ForwardEvent,
		"handler", announcements[0].Id,
	)

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
