// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	"github.com/TheThingsNetwork/api/logfields"
	"github.com/TheThingsNetwork/api/trace"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// ResponseDeadline indicates how long
var ResponseDeadline = 100 * time.Millisecond

func (h *handler) HandleUplink(uplink *pb_broker.DeduplicatedUplinkMessage) (err error) {
	appID, devID := uplink.AppID, uplink.DevID
	ctx := h.Ctx.WithFields(logfields.ForMessage(uplink))
	start := time.Now()

	h.RegisterReceived(uplink)
	defer func() {
		if err != nil {
			select {
			case h.qEvent <- &types.DeviceEvent{
				AppID: appID,
				DevID: devID,
				Event: types.UplinkErrorEvent,
				Data:  types.ErrorEventData{Error: err.Error()},
			}:
			case <-time.After(eventPublishTimeout):
				ctx.Warnf("Could not emit %q event", types.UplinkErrorEvent)
			}
			ctx.WithError(err).Warn("Could not handle uplink")
			uplink.Trace = uplink.Trace.WithEvent(trace.DropEvent, "reason", err)
		} else {
			h.RegisterHandled(uplink)
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled uplink")
		}
		// if uplink != nil && h.monitorStream != nil {
		// 	h.monitorStream.Send(uplink)
		// }
	}()
	h.status.uplink.Mark(1)

	uplink.Trace = uplink.Trace.WithEvent(trace.ReceiveEvent)

	dev, err := h.devices.Get(appID, devID)
	if err != nil {
		return err
	}
	dev.StartUpdate()

	// Build AppUplink
	appUplink := &types.UplinkMessage{
		AppID: appID,
		DevID: devID,
		Attributes: map[string]string{
			"warning": "The Things Network V2 is shutting down in December 2021. Migrate to V3 NOW!",
		},
	}

	// Get Uplink Processors
	processors := []UplinkProcessor{
		h.ConvertFromLoRaWAN,
		h.ConvertMetadata,
		h.ConvertFieldsUp,
	}

	ctx.WithField("NumProcessors", len(processors)).Debug("Running Uplink Processors")
	uplink.Trace = uplink.Trace.WithEvent("process uplink")

	// Run Uplink Processors
	for _, processor := range processors {
		err = processor(ctx, uplink, appUplink, dev)
		if err == ErrNotNeeded {
			err = nil
			return nil
		} else if err != nil {
			return err
		}
	}

	err = h.devices.Set(dev)
	if err != nil {
		return err
	}
	dev.StartUpdate()

	// Publish Uplink
	h.qUp <- appUplink

	noDownlinkErrEvent := &types.DeviceEvent{
		AppID: appID,
		DevID: devID,
		Event: types.DownlinkErrorEvent,
		Data:  types.ErrorEventData{Error: "No gateways available for downlink"},
	}

	if dev.CurrentDownlink == nil {
		<-time.After(ResponseDeadline)

		queue, err := h.devices.DownlinkQueue(appID, devID)
		if err != nil {
			return err
		}

		if len, _ := queue.Length(); len > 0 {
			if uplink.ResponseTemplate != nil {
				next, err := queue.Next()
				if err != nil {
					return err
				}
				dev.CurrentDownlink = next
			} else {
				select {
				case h.qEvent <- noDownlinkErrEvent:
				case <-time.After(eventPublishTimeout):
					ctx.Warnf("Could not emit %q event", noDownlinkErrEvent.Event)
				}
				return nil
			}
		}
	}

	if uplink.ResponseTemplate == nil {
		if dev.CurrentDownlink != nil {
			select {
			case h.qEvent <- noDownlinkErrEvent:
			case <-time.After(eventPublishTimeout):
				ctx.Warnf("Could not emit %q event", noDownlinkErrEvent.Event)
			}
		}
		return nil
	}

	// Save changes (if any)
	err = h.devices.Set(dev)
	if err != nil {
		return err
	}

	// Prepare Downlink
	var appDownlink types.DownlinkMessage
	if dev.CurrentDownlink != nil {
		appDownlink = *dev.CurrentDownlink
	}
	appDownlink.AppID = uplink.AppID
	appDownlink.DevID = uplink.DevID
	downlink := uplink.ResponseTemplate
	downlink.Trace = uplink.Trace.WithEvent("prepare downlink")

	// Handle Downlink
	err = h.HandleDownlink(&appDownlink, downlink)
	if err != nil {
		return err
	}

	return nil
}
