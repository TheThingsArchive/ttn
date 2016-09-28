// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
)

type mqttEvent struct {
	AppID   string
	DevID   string
	Type    string
	Payload interface{}
}

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
	h.mqttEvent = make(chan *mqttEvent, MQTTBufferSize)

	token := h.mqttClient.SubscribeDownlink(func(client mqtt.Client, appID string, devID string, msg mqtt.DownlinkMessage) {
		down := &msg
		down.DevID = devID
		down.AppID = appID
		go h.EnqueueDownlink(down)
	})
	token.Wait()
	if token.Error() != nil {
		return err
	}

	go func() {
		for up := range h.mqttUp {
			h.Ctx.WithFields(log.Fields{
				"DevID": up.DevID,
				"AppID": up.AppID,
			}).Debug("Publish Uplink")
			upToken := h.mqttClient.PublishUplink(*up)
			go func() {
				if upToken.WaitTimeout(MQTTTimeout) {
					if upToken.Error() != nil {
						h.Ctx.WithError(upToken.Error()).Warn("Could not publish Uplink")
					}
				} else {
					h.Ctx.Warn("Uplink publish timeout")
				}
			}()
			if len(up.PayloadFields) > 0 {
				fieldsToken := h.mqttClient.PublishUplinkFields(up.AppID, up.DevID, up.PayloadFields)
				go func() {
					if fieldsToken.WaitTimeout(MQTTTimeout) {
						if fieldsToken.Error() != nil {
							h.Ctx.WithError(fieldsToken.Error()).Warn("Could not publish Uplink Fields")
						}
					} else {
						h.Ctx.Warn("Uplink Fields publish timeout")
					}
				}()
			}
		}
	}()

	go func() {
		for activation := range h.mqttActivation {
			h.Ctx.WithFields(log.Fields{
				"DevID":   activation.DevID,
				"AppID":   activation.AppID,
				"DevEUI":  activation.DevEUI,
				"AppEUI":  activation.AppEUI,
				"DevAddr": activation.DevAddr,
			}).Debug("Publish Activation")
			token := h.mqttClient.PublishActivation(*activation)
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

	go func() {
		for event := range h.mqttEvent {
			h.Ctx.WithFields(log.Fields{
				"DevID": event.DevID,
				"AppID": event.AppID,
			}).Debug("Publish Event")
			var token mqtt.Token
			if event.DevID == "" {
				token = h.mqttClient.PublishAppEvent(event.AppID, event.Type, event.Payload)
			} else {
				token = h.mqttClient.PublishDeviceEvent(event.AppID, event.DevID, event.Type, event.Payload)
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
