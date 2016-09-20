// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"sync"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
)

// NewGateway creates a new in-memory Gateway structure
func NewGateway(ctx log.Interface, id string) *Gateway {
	ctx = ctx.WithField("GatewayID", id)
	return &Gateway{
		ID:          id,
		Status:      NewStatusStore(),
		Utilization: NewUtilization(),
		Schedule:    NewSchedule(ctx),
		Ctx:         ctx,
	}
}

// Gateway contains the state of a gateway
type Gateway struct {
	ID          string
	Status      StatusStore
	Utilization Utilization
	Schedule    Schedule
	LastSeen    time.Time

	Token     string
	tokenLock sync.Mutex

	monitor *monitorConn

	Ctx log.Interface
}

func (g *Gateway) SetToken(token string) {
	g.tokenLock.Lock()
	defer g.tokenLock.Unlock()

	g.Token = token
}

func (g *Gateway) updateLastSeen() {
	g.LastSeen = time.Now()
}

func (g *Gateway) HandleStatus(status *pb.Status) (err error) {
	if err = g.Status.Update(status); err != nil {
		return err
	}
	g.updateLastSeen()

	if g.monitor != nil {
		go func() {
			cl, err := g.statusMonitor()
			if err != nil {
				g.Ctx.WithError(errors.FromGRPCError(err)).Error("Failed to establish status connection to the monitor")
			}

			if err = cl.Send(status); err != nil {
				g.Ctx.WithError(errors.FromGRPCError(err)).Error("Monitor status push failed")
			}
		}()
	}
	return nil
}

func (g *Gateway) HandleUplink(uplink *pb_router.UplinkMessage) (err error) {
	if err = g.Utilization.AddRx(uplink); err != nil {
		return err
	}
	g.Schedule.Sync(uplink.GatewayMetadata.Timestamp)
	g.updateLastSeen()

	if g.monitor != nil {
		go func() {
			cl, err := g.uplinkMonitor()
			if err != nil {
				g.Ctx.WithError(errors.FromGRPCError(err)).Error("Failed to establish uplink connection to the monitor")
			}

			if err = cl.Send(uplink); err != nil {
				g.Ctx.WithError(errors.FromGRPCError(err)).Error("Monitor uplink push failed")
			}
		}()
	}
	return nil
}

func (g *Gateway) HandleDownlink(identifier string, downlink *pb_router.DownlinkMessage) (err error) {
	ctx := g.Ctx.WithField("Identifier", identifier)
	if err = g.Schedule.Schedule(identifier, downlink); err != nil {
		ctx.WithError(err).Warn("Could not schedule downlink")
		return err
	}

	if g.monitor != nil {
		go func() {
			cl, err := g.downlinkMonitor()
			if err != nil {
				g.Ctx.WithError(errors.FromGRPCError(err)).Error("Failed to establish downlink connection to the monitor")
			}

			if err = cl.Send(downlink); err != nil {
				g.Ctx.WithError(errors.FromGRPCError(err)).Error("Monitor downlink push failed")
			}
		}()
	}
	return nil
}
