// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
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

	h.mqttUp = make(chan *mqtt.UplinkMessage, MQTTBufferSize)
	h.mqttActivation = make(chan *mqtt.Activation, MQTTBufferSize)

	token := h.mqttClient.SubscribeDownlink(func(client mqtt.Client, appEUI types.AppEUI, devEUI types.DevEUI, msg mqtt.DownlinkMessage) {
		down := &msg
		down.DevEUI = devEUI
		down.AppEUI = appEUI
		go h.EnqueueDownlink(down)
	})
	token.Wait()
	if token.Error() != nil {
		return err
	}

	go func() {
		for up := range h.mqttUp {
			h.Ctx.WithFields(log.Fields{
				"DevEUI": up.DevEUI,
				"AppEUI": up.AppEUI,
			}).Debug("Publish Uplink")
			token := h.mqttClient.PublishUplink(up.AppEUI, up.DevEUI, *up)
			go func() {
				if token.WaitTimeout(MQTTTimeout) {
					if token.Error() != nil {
						h.Ctx.WithError(token.Error()).Warn("Could not publish Uplink")
					}
				} else {
					h.Ctx.Warn("Uplink publish timeout")
				}
			}()
		}
	}()

	go func() {
		for activation := range h.mqttActivation {
			token := h.mqttClient.PublishActivation(activation.AppEUI, activation.DevEUI, *activation)
			go func() {
				if token.WaitTimeout(MQTTTimeout) {
					if token.Error() != nil {
						h.Ctx.WithError(token.Error()).Warn("Could not publish Activation")
					}
				} else {
					h.Ctx.Warn("Activation publish timeout")
				}
			}()
		}
	}()

	return nil
}
