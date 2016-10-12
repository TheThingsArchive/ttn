package gateway

import (
	"io"
	"sync"

	pb "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_monitor "github.com/TheThingsNetwork/ttn/api/monitor"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	context "golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func (g *Gateway) monitorContext() (ctx context.Context) {
	return metadata.NewContext(context.Background(), metadata.Pairs(
		"id", g.ID,
		"token", g.Token,
	))
}

type monitorConn struct {
	clients map[string]pb_monitor.MonitorClient

	status struct {
		streams map[string]pb_monitor.Monitor_GatewayStatusClient
		sync.RWMutex
	}

	uplink struct {
		streams map[string]pb_monitor.Monitor_GatewayUplinkClient
		sync.RWMutex
	}

	downlink struct {
		streams map[string]pb_monitor.Monitor_GatewayDownlinkClient
		sync.RWMutex
	}
}

func (g *Gateway) SetMonitors(clients map[string]pb_monitor.MonitorClient) {
	g.monitor = NewMonitorConn(clients)
}

func NewMonitorConn(clients map[string]pb_monitor.MonitorClient) (conn *monitorConn) {
	conn = &monitorConn{clients: clients}
	conn.uplink.streams = make(map[string]pb_monitor.Monitor_GatewayUplinkClient)
	conn.downlink.streams = make(map[string]pb_monitor.Monitor_GatewayDownlinkClient)
	conn.status.streams = make(map[string]pb_monitor.Monitor_GatewayStatusClient)
	return conn
}

func (g *Gateway) pushStatusToMonitor(ctx log.Interface, name string, status *pb.Status) (err error) {
	defer func() {
		if err != nil {
			ctx.WithError(errors.FromGRPCError(err)).Warn("Monitor status push failed")
		} else {
			ctx.Debug("Pushed status to monitor")
		}
	}()

	g.monitor.status.RLock()
	stream, ok := g.monitor.status.streams[name]
	g.monitor.status.RUnlock()

	if !ok {
		g.monitor.status.Lock()
		if _, ok := g.monitor.status.streams[name]; !ok {
			cl, ok := g.monitor.clients[name]
			if !ok {
				// Should not happen
				return errors.New("Monitor not found")
			}

			stream, err = cl.GatewayStatus(g.monitorContext())
			if err != nil {
				ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor status stream")
				return err
			}
			ctx.Debug("Opened new monitor status stream")

			g.monitor.status.streams[name] = stream
		}
		g.monitor.status.Unlock()
	}

	if err = stream.Send(status); err == io.EOF {
		ctx.Warn("Monitor status stream closed")
		g.monitor.status.Lock()
		if g.monitor.status.streams[name] == stream {
			delete(g.monitor.status.streams, name)
		}
		g.monitor.status.Unlock()
	}
	return err
}

func (g *Gateway) pushUplinkToMonitor(ctx log.Interface, name string, uplink *pb_router.UplinkMessage) (err error) {
	defer func() {
		if err != nil {
			ctx.WithError(errors.FromGRPCError(err)).Warn("Monitor uplink push failed")
		} else {
			ctx.Debug("Pushed uplink to monitor")
		}
	}()

	g.monitor.uplink.RLock()
	stream, ok := g.monitor.uplink.streams[name]
	g.monitor.uplink.RUnlock()

	if !ok {
		g.monitor.uplink.Lock()
		if _, ok := g.monitor.uplink.streams[name]; !ok {
			cl, ok := g.monitor.clients[name]
			if !ok {
				// Should not happen
				return errors.New("Monitor not found")
			}

			stream, err = cl.GatewayUplink(g.monitorContext())
			if err != nil {
				ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor uplink stream")
				return err
			}
			ctx.Debug("Opened new monitor uplink stream")

			g.monitor.uplink.streams[name] = stream
		}
		g.monitor.uplink.Unlock()
	}

	if err = stream.Send(uplink); err == io.EOF {
		ctx.Warn("Monitor uplink stream closed")
		g.monitor.uplink.Lock()
		if g.monitor.uplink.streams[name] == stream {
			delete(g.monitor.uplink.streams, name)
		}
		g.monitor.uplink.Unlock()
	}
	return err
}

func (g *Gateway) pushDownlinkToMonitor(ctx log.Interface, name string, downlink *pb_router.DownlinkMessage) (err error) {
	defer func() {
		if err != nil {
			ctx.WithError(errors.FromGRPCError(err)).Warn("Monitor downlink push failed")
		} else {
			ctx.Debug("Pushed downlink to monitor")
		}
	}()

	g.monitor.downlink.RLock()
	stream, ok := g.monitor.downlink.streams[name]
	g.monitor.downlink.RUnlock()

	if !ok {
		g.monitor.downlink.Lock()
		if _, ok := g.monitor.downlink.streams[name]; !ok {
			cl, ok := g.monitor.clients[name]
			if !ok {
				// Should not happen
				return errors.New("Monitor not found")
			}

			stream, err = cl.GatewayDownlink(g.monitorContext())
			if err != nil {
				ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor downlink stream")
				return err
			}
			ctx.Debug("Opened new monitor downlink stream")

			g.monitor.downlink.streams[name] = stream
		}
		g.monitor.downlink.Unlock()
	}

	if err = stream.Send(downlink); err == io.EOF {
		ctx.Warn("Monitor downlink stream closed")
		g.monitor.downlink.Lock()
		if g.monitor.downlink.streams[name] == stream {
			delete(g.monitor.downlink.streams, name)
		}
		g.monitor.downlink.Unlock()
	}
	return err
}
