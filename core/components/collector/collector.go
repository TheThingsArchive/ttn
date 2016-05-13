// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collector

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"

	"github.com/TheThingsNetwork/ttn/core/collection"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/apex/log"
)

// Collector collects data from the Handler and stores it
type Collector interface {
	Start() (map[types.AppEUI]collection.AppCollector, error)
	Stop()
	StopApp(eui types.AppEUI) error
}

type collector struct {
	ctx         log.Interface
	appStorage  AppStorage
	broker      string
	apps        map[types.AppEUI]collection.AppCollector
	dataStorage collection.DataStorage
	netAddr     string
	server      *grpc.Server
}

// NewCollector creates a new collector
func NewCollector(ctx log.Interface, appStorage AppStorage, broker string, dataStorage collection.DataStorage, netAddr string) Collector {
	c := &collector{
		ctx:         ctx,
		appStorage:  appStorage,
		broker:      broker,
		apps:        map[types.AppEUI]collection.AppCollector{},
		dataStorage: dataStorage,
		netAddr:     netAddr,
		server:      grpc.NewServer(),
	}
	c.RegisterServer(c.server)
	return c
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

func (c *collector) Start() (map[types.AppEUI]collection.AppCollector, error) {
	// Start management service
	lis, err := net.Listen("tcp", c.netAddr)
	if err != nil {
		c.ctx.WithError(err).Fatalf("Listen on %s failed", c.netAddr)
	}
	go c.server.Serve(lis)

	// Get applications
	apps, err := c.appStorage.List()
	if err != nil {
		c.ctx.WithError(err).Error("Failed to get applications")
		return nil, err
	}

	// Start app collectors
	startErrors := make(map[types.AppEUI]error)
	for _, appEUI := range apps {
		if err := c.startApp(appEUI); err != nil {
			c.ctx.WithField("AppEUI", appEUI).WithError(err).Warn("Failed to start app collector; ignoring app")
			startErrors[appEUI] = err
		}
	}
	if len(startErrors) > 0 {
		return c.apps, StartError{startErrors}
	}
	return c.apps, nil
}

func (c *collector) Stop() {
	c.server.Stop()

	for _, app := range c.apps {
		app.Stop()
	}
	c.apps = nil
}

func (c *collector) startApp(eui types.AppEUI) error {
	ctx := c.ctx.WithField("AppEUI", eui)

	key, err := c.appStorage.GetAccessKey(eui)
	if err != nil {
		return err
	} else if key == "" {
		return errors.New("Not found")
	}

	ac := collection.NewMqttAppCollector(ctx, c.broker, eui, key, c.dataStorage)
	ctx.Info("Starting app collector")
	if err := ac.Start(); err != nil {
		return err
	}
	c.apps[eui] = ac
	return nil
}

func (c *collector) StopApp(eui types.AppEUI) error {
	app, ok := c.apps[eui]
	if !ok {
		return errors.New("Not found")
	}
	app.Stop()
	delete(c.apps, eui)
	return nil
}
