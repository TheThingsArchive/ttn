// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"sync"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/rate"
	"github.com/TheThingsNetwork/ttn/api"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_monitor "github.com/TheThingsNetwork/ttn/api/monitor"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/band"
	"github.com/TheThingsNetwork/ttn/utils/toa"
)

// NewGateway creates a new in-memory Gateway structure
func NewGateway(ctx ttnlog.Interface, id string, monitor *pb_monitor.Client) *Gateway {
	ctx = ctx.WithField("GatewayID", id)
	gtw := &Gateway{
		ID:       id,
		Ctx:      ctx,
		uplink:   rate.NewCounter(time.Minute, time.Hour),
		downlink: rate.NewCounter(time.Minute, time.Hour),
		monitor:  monitor,
	}
	gtw.schedule = NewSchedule(ctx, func(msg *pb_router.DownlinkMessage) time.Duration {
		gtw.mu.RLock()
		defer gtw.mu.RUnlock()
		if gtw.frequencyPlan != nil {
			toa, _ := toa.Compute(msg)
			return gtw.frequencyPlan.Limits.TimeOffAir(toa)
		}
		return 0
	})
	return gtw
}

// Gateway contains the state of a gateway
type Gateway struct {
	ID       string
	Ctx      ttnlog.Interface
	uplink   rate.Counter
	downlink rate.Counter
	schedule *Schedule

	mu sync.RWMutex // Protect all fields below

	// downlink
	frequencyPlan *band.FrequencyPlan

	// status
	status   pb_gateway.Status
	lastSeen time.Time

	// monitoring
	authToken     string
	monitor       *pb_monitor.Client
	monitorStream pb_monitor.GenericStream
}

func (g *Gateway) SendToMonitor(msg interface{}) {
	if g == nil {
		return
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	g.sendToMonitor(msg)
}

func (g *Gateway) sendToMonitor(msg interface{}) {
	if g.monitorStream != nil {
		g.monitorStream.Send(msg)
	}
}

func (g *Gateway) SetAuth(token string, authenticated bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.status.GatewayTrusted = authenticated
	if g.monitorStream != nil {
		if token == g.authToken {
			return
		}
		g.Ctx.Debug("Stopping Monitor stream (token changed)")
		g.monitorStream.Close()
	}
	g.authToken = token
	if g.monitor != nil {
		g.Ctx.Debug("Starting Gateway Monitor Stream")
		g.monitorStream = g.monitor.NewGatewayStreams(g.ID, g.authToken)
	}
}

func (g *Gateway) LastSeen() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.lastSeen
}

func (g *Gateway) Status() pb_gateway.Status {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.status
}

func (g *Gateway) Rates() (uplink, downlink *api.Rates) {
	frequencyPlan := g.FrequencyPlan()
	if frequencyPlan == nil {
		return
	}
	now := time.Now()
	uplink = new(api.Rates)
	uplink1, _ := g.uplink.Get(now, time.Minute)
	uplink.Rate1 = float32(uplink1) / float32(time.Minute)
	uplink5, _ := g.uplink.Get(now, 5*time.Minute)
	uplink.Rate5 = float32(uplink5) / float32(5*time.Minute)
	uplink15, _ := g.uplink.Get(now, 15*time.Minute)
	uplink.Rate15 = float32(uplink15) / float32(15*time.Minute)
	downlink = new(api.Rates)
	downlink1, _ := g.downlink.Get(now, time.Minute)
	downlink.Rate1 = float32(downlink1) / float32(time.Minute)
	downlink5, _ := g.downlink.Get(now, 5*time.Minute)
	downlink.Rate5 = float32(downlink5) / float32(5*time.Minute)
	downlink15, _ := g.downlink.Get(now, 15*time.Minute)
	downlink.Rate15 = float32(downlink15) / float32(15*time.Minute)
	return
}

func (g *Gateway) FrequencyPlan() *band.FrequencyPlan {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.frequencyPlan
}

func (g *Gateway) setFrequencyPlan(fpName string) error {
	fp, err := band.Get(fpName)
	if err == nil {
		g.frequencyPlan = &fp
		g.frequencyPlan.Limits = fp.Limits.New()
	}
	return err
}
