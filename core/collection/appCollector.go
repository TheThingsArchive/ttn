// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collection

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
)

// AppCollector represents a collector for application data
type AppCollector interface {
	Start() error
	Stop()
}

type appCollector struct {
	ctx        log.Interface
	eui        types.AppEUI
	mqttBroker string
	client     mqtt.Client
	storage    DataStorage
}

// NewMqttAppCollector instantiates a new AppCollector instance using MQTT
func NewMqttAppCollector(ctx log.Interface, mqttBroker string, eui types.AppEUI, key string, storage DataStorage) AppCollector {
	return &appCollector{
		ctx:        ctx,
		eui:        eui,
		mqttBroker: mqttBroker,
		client:     mqtt.NewClient(ctx, "collector", eui.String(), key, fmt.Sprintf("tcp://%s", mqttBroker)),
		storage:    storage,
	}
}

func (c *appCollector) Start() error {
	err := c.client.Connect()
	if err != nil {
		c.ctx.WithError(err).Error("Connect failed")
		return err
	}
	if token := c.client.SubscribeAppUplink(c.eui, c.handleUplink); token.Wait() && token.Error() != nil {
		c.ctx.WithError(token.Error()).Error("Failed to subscribe")
		return token.Error()
	}
	c.ctx.WithField("Broker", c.mqttBroker).Info("Subscribed to app uplink packets")
	return nil
}

func (c *appCollector) Stop() {
	c.client.Disconnect()
}

func (c *appCollector) handleUplink(client mqtt.Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) {
	if req.Fields == nil || len(req.Fields) == 0 {
		return
	}

	ctx := c.ctx.WithField("DevEUI", devEUI)

	t, err := time.Parse(time.RFC3339, req.Metadata[0].ServerTime)
	if err != nil {
		ctx.WithError(err).Warnf("Invalid time: %v", req.Metadata[0].ServerTime)
		return
	}

	err = c.storage.Save(appEUI, devEUI, t, req.Fields)
	if err != nil {
		ctx.WithError(err).Error("Failed to save data")
		return
	}
	ctx.Debug("Saved uplink packet in store")
}
