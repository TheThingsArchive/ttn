// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
)

var ResponseDeadline = 100 * time.Millisecond

func (h *handler) HandleUplink(uplink *pb_broker.DeduplicatedUplinkMessage) error {
	ctx := h.Ctx.WithFields(log.Fields{
		"AppID": uplink.AppId,
		"DevID": uplink.DevId,
	})
	var err error
	start := time.Now()
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle uplink")
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled uplink")
		}
	}()

	// Build AppUplink
	appUplink := &mqtt.UplinkMessage{
		AppID: uplink.AppId,
		DevID: uplink.DevId,
	}

	// Get Uplink Processors
	processors := []UplinkProcessor{
		h.ConvertFromLoRaWAN,
		h.ConvertMetadata,
		h.ConvertFields,
	}

	ctx.WithField("NumProcessors", len(processors)).Debug("Running Uplink Processors")

	// Run Uplink Processors
	for _, processor := range processors {
		err = processor(ctx, uplink, appUplink)
		if err == ErrNotNeeded {
			err = nil
			return nil
		} else if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// Publish Uplink
	h.mqttUp <- appUplink

	<-time.After(ResponseDeadline)

	// Find Device and scheduled downlink
	var appDownlink mqtt.DownlinkMessage
	dev, err := h.devices.Get(uplink.AppId, uplink.DevId)
	if err != nil {
		return err
	}
	if dev.NextDownlink != nil {
		appDownlink = *dev.NextDownlink
	}

	if uplink.ResponseTemplate == nil {
		ctx.Debug("No Downlink Available")
		return nil
	}

	// Prepare Downlink
	downlink := uplink.ResponseTemplate
	appDownlink.AppID = uplink.AppId
	appDownlink.DevID = uplink.DevId

	// Handle Downlink
	err = h.HandleDownlink(&appDownlink, downlink)
	if err != nil {
		return err
	}

	// Clear Downlink
	dev.NextDownlink = nil
	err = h.devices.Set(dev, "next_downlink")
	if err != nil {
		return err
	}

	return nil
}
