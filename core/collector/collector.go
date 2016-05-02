// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collector

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/core/collection"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/apex/log"
)

// Collector collects data from the Handler and stores it
type Collector interface {
	Start() ([]collection.AppCollector, error)
	Stop()
}

type collector struct {
	ctx         log.Interface
	appStorage  AppStorage
	broker      string
	apps        []collection.AppCollector
	dataStorage collection.DataStorage
}

// NewCollector creates a new collector
func NewCollector(ctx log.Interface, appStorage AppStorage, broker string, dataStorage collection.DataStorage) Collector {
	return &collector{
		ctx:         ctx,
		appStorage:  appStorage,
		broker:      broker,
		dataStorage: dataStorage,
	}
}

// StartError contains errors of starting applications
type StartError struct {
	errors map[types.AppEUI]error
}

func (e StartError) Error() string {
	var s string
	for eui, err := range e.errors {
		s += fmt.Sprintf("%v: %v\n", eui, err)
	}
	return strings.TrimRight(s, "\n")
}

func (c *collector) Start() ([]collection.AppCollector, error) {
	apps, err := c.appStorage.GetAll()
	if err != nil {
		c.ctx.WithError(err).Error("Failed to get applications")
		return nil, err
	}

	startErrors := make(map[types.AppEUI]error)
	for _, app := range apps {
		ctx := c.ctx.WithField("appEUI", app.EUI)
		ac := collection.NewMqttAppCollector(ctx, c.broker, app.EUI, app.Key, &app.Functions, c.dataStorage)
		ctx.Info("Starting app collector")
		if err := ac.Start(); err != nil {
			ctx.WithError(err).Warn("Failed to start app collector; ignoring app")
			startErrors[app.EUI] = err
			continue
		}
		c.apps = append(c.apps, ac)
	}

	if len(startErrors) > 0 {
		return c.apps, StartError{startErrors}
	}
	return c.apps, nil
}

func (c *collector) Stop() {
	for _, app := range c.apps {
		app.Stop()
	}
	c.apps = nil
}
