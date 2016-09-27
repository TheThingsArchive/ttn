// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
)

func (h *handler) EnqueueDownlink(appDownlink *mqtt.DownlinkMessage) error {
	appID, devID := appDownlink.AppID, appDownlink.DevID

	ctx := h.Ctx.WithFields(log.Fields{
		"AppID": appID,
		"DevID": devID,
	})
	var err error
	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not enqueue downlink")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Debug("Enqueued downlink")
		}
	}()

	var dev *device.Device
	dev, err = h.devices.Get(appID, devID)
	if err != nil {
		return err
	}
	// Clear redundant fields
	appDownlink.AppID = ""
	appDownlink.DevID = ""
	dev.NextDownlink = appDownlink
	err = h.devices.Set(dev, "next_downlink")
	if err != nil {
		return err
	}

	h.mqttEvent <- &mqttEvent{
		AppID:   appID,
		DevID:   devID,
		Type:    "down/scheduled",
		Payload: appDownlink,
	}

	return nil
}

func (h *handler) HandleDownlink(appDownlink *mqtt.DownlinkMessage, downlink *pb_broker.DownlinkMessage) error {
	appID, devID := appDownlink.AppID, appDownlink.DevID

	ctx := h.Ctx.WithFields(log.Fields{
		"AppID":  appID,
		"DevID":  devID,
		"AppEUI": downlink.AppEui,
		"DevEUI": downlink.DevEui,
	})

	var err error
	defer func() {
		if err != nil {
			h.mqttEvent <- &mqttEvent{
				AppID:   appID,
				DevID:   devID,
				Type:    "down/errors",
				Payload: map[string]string{"error": err.Error()},
			}
			ctx.WithError(err).Warn("Could not handle downlink")
		}
	}()

	// Get Processors
	processors := []DownlinkProcessor{
		h.ConvertFieldsDown,
		h.ConvertToLoRaWAN,
	}

	ctx.WithField("NumProcessors", len(processors)).Debug("Running Downlink Processors")

	// Run Processors
	for _, processor := range processors {
		err = processor(ctx, appDownlink, downlink)
		if err == ErrNotNeeded {
			err = nil
			return nil
		} else if err != nil {
			return err
		}
	}

	ctx.Debug("Send Downlink")

	h.downlink <- downlink

	appDownlinkCopy := *appDownlink
	appDownlinkCopy.AppID = ""
	appDownlinkCopy.DevID = ""
	appDownlinkCopy.Fields = make(map[string]interface{})

	h.mqttEvent <- &mqttEvent{
		AppID:   appDownlink.AppID,
		DevID:   appDownlink.DevID,
		Type:    "down/sent",
		Payload: appDownlinkCopy,
	}

	return nil
}
