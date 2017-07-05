// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"sync"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_monitor "github.com/TheThingsNetwork/ttn/api/monitor"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
)

// NewGateway creates a new in-memory Gateway structure
func NewGateway(ctx ttnlog.Interface, id string) *Gateway {
	ctx = ctx.WithField("GatewayID", id)
	gtw := &Gateway{
		ID:          id,
		Status:      NewStatusStore(),
		Utilization: NewUtilization(),
		Schedule:    NewSchedule(ctx),
		Ctx:         ctx,
	}
	gtw.Schedule.(*schedule).gateway = gtw // FIXME: Issue #420
	return gtw
}

// Gateway contains the state of a gateway
type Gateway struct {
	ID          string
	Status      StatusStore
	Utilization Utilization
	Schedule    Schedule
	LastSeen    time.Time

	mu            sync.RWMutex // Protect token and authenticated
	token         string
	authenticated bool

	Monitor       *pb_monitor.Client
	MonitorStream pb_monitor.GenericStream

	Ctx ttnlog.Interface
}

func (g *Gateway) SetAuth(token string, authenticated bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.authenticated = authenticated
	if g.MonitorStream != nil {
		if token == g.token {
			return
		}
		g.Ctx.Debug("Stopping Monitor stream (token changed)")
		g.MonitorStream.Close()
	}
	g.token = token
	if g.Monitor != nil {
		g.Ctx.Debug("Starting Gateway Monitor Stream")
		g.MonitorStream = g.Monitor.NewGatewayStreams(g.ID, g.token)
	}
}

func (g *Gateway) updateLastSeen() {
	g.LastSeen = time.Now()
}

func (g *Gateway) HandleStatus(status *pb.Status) (err error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	status.GatewayTrusted = g.authenticated
	if err = g.Status.Update(status); err != nil {
		return err
	}
	g.updateLastSeen()
	return nil
}

func (g *Gateway) HandleUplink(uplink *pb_router.UplinkMessage) (err error) {
	if err = g.Utilization.AddRx(uplink); err != nil {
		return err
	}
	g.Schedule.Sync(uplink.GatewayMetadata.Timestamp)
	g.updateLastSeen()

	status, err := g.Status.Get()
	if err == nil {
		// Inject Gateway location
		if uplink.GatewayMetadata.Location == nil {
			uplink.GatewayMetadata.Location = status.GetLocation()
		}
		// Inject Gateway frequency plan
		if frequencyPlan, ok := pb_lorawan.FrequencyPlan_value[status.FrequencyPlan]; ok {
			if lorawan := uplink.GetProtocolMetadata().GetLorawan(); lorawan != nil {
				lorawan.FrequencyPlan = pb_lorawan.FrequencyPlan(frequencyPlan)
			}
		}
	}

	// Inject authenticated as GatewayTrusted
	g.mu.RLock()
	defer g.mu.RUnlock()
	uplink.GatewayMetadata.GatewayTrusted = g.authenticated
	uplink.GatewayMetadata.GatewayId = g.ID
	return nil
}

func (g *Gateway) HandleDownlink(identifier string, downlink *pb_router.DownlinkMessage) (err error) {
	ctx := g.Ctx.WithField("Identifier", identifier).WithFields(fields.Get(downlink))
	if err = g.Schedule.Schedule(identifier, downlink); err != nil {
		ctx.WithError(err).Warn("Could not schedule downlink")
		return err
	}
	ctx.Debug("Scheduled downlink")
	return nil
}
