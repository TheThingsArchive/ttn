// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
)

func (h *handler) EnqueueDownlink(appDownlink *mqtt.DownlinkMessage) error {
	ctx := h.Ctx.WithFields(log.Fields{
		"DevEUI": appDownlink.DevEUI,
		"AppEUI": appDownlink.AppEUI,
	})
	var err error
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not enqueue downlink")
		}
	}()

	dev, err := h.devices.Get(appDownlink.AppEUI, appDownlink.DevEUI)
	if err != nil {
		return err
	}
	dev.NextDownlink = appDownlink
	err = h.devices.Set(dev, "next_downlink")
	if err != nil {
		return err
	}

	ctx.Debug("Enqueue Downlink")

	return nil
}

func (h *handler) HandleDownlink(appDownlink *mqtt.DownlinkMessage, downlink *pb_broker.DownlinkMessage) error {
	ctx := h.Ctx.WithFields(log.Fields{
		"DevEUI": appDownlink.DevEUI,
		"AppEUI": appDownlink.AppEUI,
	})
	var err error
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle downlink")
		}
	}()

	// Get Processors
	processors := []DownlinkProcessor{
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

	return nil
}
