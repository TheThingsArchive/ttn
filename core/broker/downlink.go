// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sort"
	"strings"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/core/band"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/toa"
)

type downlinkOption struct {
	gatewayPreference bool
	uplinkMetadata    *pb_gateway.RxMetadata
	option            *pb.DownlinkOption
}

// ByScore is used to sort a list of DownlinkOptions based on Score
type ByScore []downlinkOption

func (a ByScore) Len() int      { return len(a) }
func (a ByScore) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool {
	var pointsI, pointsJ int

	gradeBool := func(i, j bool, weight int) {
		if i {
			pointsI += weight
		}
		if j {
			pointsJ += weight
		}
	}

	gradeHighest := func(i, j float32, weight int) {
		if i > j {
			pointsI += weight
		}
		if i < j {
			pointsJ += weight
		}
	}

	gradeLowest := func(i, j float32, weight int) {
		if i < j {
			pointsI += weight
		}
		if i > j {
			pointsJ += weight
		}
	}

	// TODO: Score is deprecated, remove it
	if a[i].option.GetScore() != 0 && a[j].option.GetScore() != 0 {
		gradeLowest(float32(a[i].option.GetScore()), float32(a[j].option.GetScore()), 10)
		return pointsI > pointsJ
	}

	gradeBool(a[i].gatewayPreference, a[j].gatewayPreference, 10)
	gradeHighest(a[i].uplinkMetadata.GetSnr(), a[j].uplinkMetadata.GetSnr(), 1)
	gradeHighest(a[i].uplinkMetadata.GetRssi(), a[j].uplinkMetadata.GetRssi(), 1)
	gradeLowest(float32(a[i].option.GetPossibleConflicts()), float32(a[j].option.GetPossibleConflicts()), 1)
	gradeLowest(a[i].option.GetUtilization(), a[j].option.GetUtilization(), 1)
	gradeLowest(a[i].option.GetDutyCycle(), a[j].option.GetDutyCycle(), 1)

	if a[i].option != nil && a[j].option != nil {
		toaI, _ := toa.Compute(a[i].option)
		toaJ, _ := toa.Compute(a[j].option)
		gradeLowest(float32(toaI.Seconds()), float32(toaJ.Seconds()), 1)
	}

	return pointsI > pointsJ
}

func selectBestDownlink(options []downlinkOption) *pb.DownlinkOption {
	scored := ByScore(options)
	sort.Sort(scored)
	return scored[0].option
}

func selectBestDownlinkOption(options []*pb_router.DownlinkOptionResponse) *pb_router.DownlinkOptionResponse {
	return options[0] // TODO
}

func (b *broker) PrepareDownlink(req *pb.PrepareDownlinkRequest) (downlink *pb.DownlinkMessage, err error) {
	dev, err := b.ns.GetDevice(b.Component.GetContext(b.nsToken), &pb_lorawan.DeviceIdentifier{
		AppEui: req.AppEui,
		DevEui: req.DevEui,
	})
	if err != nil {
		return nil, err
	}
	if dev.AppId != req.AppId || dev.DevId != req.DevId {
		return nil, errors.NewErrInvalidArgument("Device ID", "does not match EUI")
	}
	if dev.GetClass() != pb_lorawan.Class_C {
		return nil, errors.NewErrInvalidArgument("Device", "can not handle Class C downlink")
	}
	if len(dev.PreferredGateways) == 0 {
		return nil, errors.NewErrInvalidArgument("Device", "does not have PreferredGateways for Class C downlink")
	}
	if dev.DevAddr == nil {
		return nil, errors.NewErrInvalidArgument("Device", "does not have a DevAddr")
	}

	// Frequency Plan
	fp, err := band.Get(dev.FrequencyPlan.String())
	if err != nil {
		return nil, err
	}
	drIdx := fp.RX2DataRate
	dr, _ := types.ConvertDataRate(fp.DataRates[drIdx])
	dataRate := dr.String()
	frequency := uint64(fp.RX2Frequency)
	if dev.Rx2DataRate != "" {
		dataRate = dev.Rx2DataRate
		drIdx, err = fp.GetDataRateIndexFor(dataRate)
		if err != nil {
			return nil, err
		}
	}
	if dev.Rx2Frequency != 0 {
		frequency = dev.Rx2Frequency
	}
	duration, _ := toa.ComputeLoRa(uint(fp.MaxPayloadSize[drIdx].M), dataRate, "4/5")

	downlink = &pb.DownlinkMessage{
		AppEui:  req.AppEui,
		DevEui:  req.DevEui,
		AppId:   req.AppId,
		DevId:   req.DevId,
		Message: new(pb_protocol.Message),
		DownlinkOption: &pb.DownlinkOption{
			ProtocolConfig: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{
				Modulation: pb_lorawan.Modulation_LORA,
				DataRate:   dataRate,
				CodingRate: "4/5",
				FCnt:       dev.FCntDown,
			}}},
			GatewayConfig: &pb_gateway.TxConfiguration{
				Frequency: frequency,
				Power:     int32(fp.DefaultTXPower),
				PolarizationInversion: true,
			},
		},
	}

	options := make([]downlinkOption, 0, len(dev.PreferredGateways))
	for _, gatewayID := range dev.PreferredGateways {
		routerID := b.getRouterForGateway(gatewayID)
		if routerID == "" {
			continue
		}
		router, err := b.Discover("router", routerID)
		if err != nil {
			continue
		}
		routerConn, err := router.Dial(b.Component.Pool)
		if err != nil {
			continue
		}
		defer b.Component.Pool.CloseConn(routerConn)
		routerCli := pb_router.NewRouterClient(routerConn)

		option, err := routerCli.GetDownlinkOption(b.Component.GetContext(""), &pb_router.DownlinkOptionRequest{
			GatewayId: gatewayID,
			Frequency: downlink.GetDownlinkOption().GetGatewayConfig().GetFrequency(),
			Duration:  duration,
		})
		if err != nil {
			continue
		}
		optionWithInfo := *downlink.DownlinkOption
		optionWithInfo.Identifier = router.Id + ":" + option.Identifier
		optionWithInfo.GatewayId = gatewayID
		optionWithInfo.DutyCycle = option.DutyCycle
		optionWithInfo.PossibleConflicts = option.PossibleConflicts
		optionWithInfo.Utilization = option.Utilization
		options = append(options, downlinkOption{option: &optionWithInfo})
	}
	if len(options) == 0 {
		return nil, errors.NewErrNotFound("Gateway with downlink availability")
	}

	lorawanDownlinkMsg := downlink.Message.InitLoRaWAN()
	lorawanDownlinkMac := lorawanDownlinkMsg.InitDownlink()
	lorawanDownlinkMac.DevAddr = *dev.DevAddr
	lorawanDownlinkMac.FCnt = dev.FCntDown
	downlink.Payload, _ = lorawanDownlinkMsg.PHYPayload().MarshalBinary()

	downlink.DownlinkOption = selectBestDownlink(options)
	return downlink, nil
}

func (b *broker) HandleDownlink(downlink *pb.DownlinkMessage) (err error) {
	ctx := b.Ctx.WithFields(fields.Get(downlink))
	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle downlink")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled downlink")
		}
		if downlink != nil && b.monitorStream != nil {
			b.monitorStream.Send(downlink)
		}
	}()

	b.status.downlink.Mark(1)

	downlink.Trace = downlink.Trace.WithEvent(trace.ReceiveEvent)

	downlink, err = b.ns.Downlink(b.Component.GetContext(b.nsToken), downlink)
	if err != nil {
		return errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not handle downlink")
	}

	var routerID string
	if id := strings.Split(downlink.DownlinkOption.Identifier, ":"); len(id) == 2 {
		routerID = id[0]
	} else {
		return errors.NewErrInvalidArgument("DownlinkOption Identifier", "invalid format")
	}
	ctx = ctx.WithField("RouterID", routerID)

	router, err := b.getRouter(routerID)
	if err != nil {
		return err
	}

	downlink.Trace = downlink.Trace.WithEvent(trace.ForwardEvent, "router", routerID)

	router <- downlink

	return nil
}
