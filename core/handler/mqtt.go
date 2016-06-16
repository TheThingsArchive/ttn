package handler

import (
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
)

func (h *handler) HandleMQTT(username, password string, mqttBrokers ...string) error {
	h.mqttClient = mqtt.NewClient(h.Ctx, "ttnhdl", username, password, mqttBrokers...)

	err := h.mqttClient.Connect()
	if err != nil {
		return err
	}

	h.mqttUp = make(chan *mqtt.UplinkMessage)
	h.mqttActivation = make(chan *mqtt.Activation)

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
				token.Wait()
				if token.Error() != nil {
					h.Ctx.WithError(token.Error()).Warn("Could not publish Uplink")
				}
			}()
		}
	}()

	go func() {
		for activation := range h.mqttActivation {
			token := h.mqttClient.PublishActivation(activation.AppEUI, activation.DevEUI, *activation)
			go func() {
				token.Wait()
				if token.Error() != nil {
					h.Ctx.WithError(token.Error()).Warn("Could not publish Activation")
				}
			}()
		}
	}()

	return nil
}
