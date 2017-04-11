// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
)

// MQTTTimeout indicates how long we should wait for an MQTT publish
var MQTTTimeout = 2 * time.Second

// MQTTBufferSize indicates the size for uplink channel buffers
var MQTTBufferSize = 10

func (h *handler) HandleMQTT(username, password string, mqttBrokers ...string) error {
	h.mqttClient = mqtt.NewClient(h.Ctx, "ttnhdl", username, password, mqttBrokers...)

	err := h.mqttClient.Connect()
	if err != nil {
		return err
	}

	h.mqttUp = make(chan *types.UplinkMessage, MQTTBufferSize)
	h.mqttEvent = make(chan *types.DeviceEvent, MQTTBufferSize)

	token := h.mqttClient.SubscribeDownlink(func(client mqtt.Client, appID string, devID string, msg types.DownlinkMessage) {
		down := &msg
		down.DevID = devID
		down.AppID = appID
		go h.EnqueueDownlink(down)
	})
	token.Wait()
	if token.Error() != nil {
		return err
	}

	ctx := h.Ctx.WithField("Protocol", "MQTT")

	go func() {
		for up := range h.mqttUp {
			ctx.WithFields(ttnlog.Fields{
				"DevID": up.DevID,
				"AppID": up.AppID,
			}).Debug("Publish Uplink")
			upToken := h.mqttClient.PublishUplink(*up)
			go func(ctx ttnlog.Interface) {
				if upToken.WaitTimeout(MQTTTimeout) {
					if upToken.Error() != nil {
						ctx.WithError(upToken.Error()).Warn("Could not publish Uplink")
					}
				} else {
					ctx.Warn("Uplink publish timeout")
				}
			}(ctx)
			if len(up.PayloadFields) > 0 {
				fieldsToken := h.mqttClient.PublishUplinkFields(up.AppID, up.DevID, up.PayloadFields)
				go func(ctx ttnlog.Interface) {
					if fieldsToken.WaitTimeout(MQTTTimeout) {
						if fieldsToken.Error() != nil {
							ctx.WithError(fieldsToken.Error()).Warn("Could not publish Uplink Fields")
						}
					} else {
						ctx.Warn("Uplink Fields publish timeout")
					}
				}(ctx)
			}
		}
	}()

	go func() {
		for event := range h.mqttEvent {
			ctx.WithFields(ttnlog.Fields{
				"DevID": event.DevID,
				"AppID": event.AppID,
				"Event": event.Event,
			}).Debug("Publish Event")
			var token mqtt.Token
			if event.DevID == "" {
				token = h.mqttClient.PublishAppEvent(event.AppID, event.Event, event.Data)
			} else {
				token = h.mqttClient.PublishDeviceEvent(event.AppID, event.DevID, event.Event, event.Data)
			}
			go func() {
				if token.WaitTimeout(MQTTTimeout) {
					if token.Error() != nil {
						h.Ctx.WithError(token.Error()).Warn("Could not publish Event")
					}
				} else {
					h.Ctx.Warn("Event publish timeout")
				}
			}()
		}
	}()

	return nil
}
