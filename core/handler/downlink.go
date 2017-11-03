// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/api/trace"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/toa"
)

func (h *handler) EnqueueDownlink(appDownlink *types.DownlinkMessage) (err error) {
	appID, devID := appDownlink.AppID, appDownlink.DevID
	ctx := h.Ctx.WithFields(ttnlog.Fields{
		"AppID": appID,
		"DevID": devID,
	})

	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not enqueue downlink")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Debug("Enqueued downlink")
		}
	}()

	// Check if device exists
	dev, err := h.devices.Get(appID, devID)
	if err != nil {
		return err
	}
	dev.StartUpdate()

	defer func() {
		if err != nil {
			h.qEvent <- &types.DeviceEvent{
				AppID: appID,
				DevID: devID,
				Event: types.DownlinkErrorEvent,
				Data: types.DownlinkEventData{
					ErrorEventData: types.ErrorEventData{Error: err.Error()},
					Message:        appDownlink,
				},
			}
		}
	}()

	if len(appDownlink.PayloadRaw) == 0 && len(appDownlink.PayloadFields) == 0 {
		return errors.NewErrInvalidArgument("Downlink Payload", "empty")
	}

	// Clear redundant fields
	appDownlink.AppID = ""
	appDownlink.DevID = ""

	queue, err := h.devices.DownlinkQueue(appID, devID)
	if err != nil {
		return err
	}

	schedule := appDownlink.Schedule
	appDownlink.Schedule = ""

	switch schedule {
	case types.ScheduleReplace, "": // Empty string for default
		dev.CurrentDownlink = nil
		err = queue.Replace(appDownlink)
	case types.ScheduleFirst:
		err = queue.PushFirst(appDownlink)
	case types.ScheduleLast:
		err = queue.PushLast(appDownlink)
	default:
		return errors.NewErrInvalidArgument("ScheduleType", "unknown")
	}

	if err != nil {
		return err
	}

	if err := h.devices.Set(dev); err != nil {
		return err
	}

	h.qEvent <- &types.DeviceEvent{
		AppID: appID,
		DevID: devID,
		Event: types.DownlinkScheduledEvent,
		Data: types.DownlinkEventData{
			Message: appDownlink,
		},
	}
	return nil
}

func (h *handler) HandleDownlink(appDownlink *types.DownlinkMessage, downlink *pb_broker.DownlinkMessage) (err error) {
	appID, devID := appDownlink.AppID, appDownlink.DevID

	ctx := h.Ctx.WithFields(ttnlog.Fields{
		"AppID":  appID,
		"DevID":  devID,
		"AppEUI": downlink.AppEUI,
		"DevEUI": downlink.DevEUI,
	})

	defer func() {
		if err != nil {
			h.qEvent <- &types.DeviceEvent{
				AppID: appID,
				DevID: devID,
				Event: types.DownlinkErrorEvent,
				Data: types.DownlinkEventData{
					ErrorEventData: types.ErrorEventData{Error: err.Error()},
					Message:        appDownlink,
				},
			}
			ctx.WithError(err).Warn("Could not handle downlink")
			downlink.Trace = downlink.Trace.WithEvent(trace.DropEvent, "reason", err)
		}
		if downlink != nil && h.monitorStream != nil {
			h.monitorStream.Send(downlink)
		}
	}()

	dev, err := h.devices.Get(appID, devID)
	if err != nil {
		return err
	}
	dev.StartUpdate()
	defer func() {
		setErr := h.devices.Set(dev)
		if err == nil {
			err = setErr
		}
	}()

	// Get Processors
	processors := []DownlinkProcessor{
		h.ConvertFieldsDown,
		h.ConvertToLoRaWAN,
	}

	ctx.WithField("NumProcessors", len(processors)).Debug("Running Downlink Processors")
	downlink.Trace = downlink.Trace.WithEvent("process downlink")

	// Run Processors
	for _, processor := range processors {
		err = processor(ctx, appDownlink, downlink, dev)
		if err == ErrNotNeeded {
			err = nil
			return nil
		} else if err != nil {
			return err
		}
	}

	downlink.Message = nil
	downlink.UnmarshalPayload()

	h.RegisterHandled(downlink)
	h.status.downlink.Mark(1)

	ctx.Debug("Send Downlink")

	downlink.Trace = downlink.Trace.WithEvent(trace.ForwardEvent, "broker", h.ttnBrokerID)

	h.downlink <- downlink

	downlinkConfig := types.DownlinkEventConfigInfo{}

	if lorawan := downlink.DownlinkOption.ProtocolConfiguration.GetLoRaWAN(); lorawan != nil {
		downlinkConfig.Modulation = lorawan.Modulation.String()
		downlinkConfig.DataRate = lorawan.DataRate
		downlinkConfig.BitRate = uint(lorawan.BitRate)
		downlinkConfig.FCnt = uint(lorawan.FCnt)
		switch lorawan.Modulation {
		case pb_lorawan.Modulation_LORA:
			downlinkConfig.Airtime, _ = toa.ComputeLoRa(uint(len(downlink.Payload)), lorawan.DataRate, lorawan.CodingRate)
		case pb_lorawan.Modulation_FSK:
			downlinkConfig.Airtime, _ = toa.ComputeFSK(uint(len(downlink.Payload)), int(lorawan.BitRate))
		}
	}
	downlinkConfig.Frequency = uint(downlink.DownlinkOption.GatewayConfiguration.Frequency)
	downlinkConfig.Power = int(downlink.DownlinkOption.GatewayConfiguration.Power)

	h.qEvent <- &types.DeviceEvent{
		AppID: appDownlink.AppID,
		DevID: appDownlink.DevID,
		Event: types.DownlinkSentEvent,
		Data: types.DownlinkEventData{
			Payload:   downlink.Payload,
			Message:   appDownlink,
			GatewayID: downlink.DownlinkOption.GatewayID,
			Config:    &downlinkConfig,
		},
	}
	return nil
}
