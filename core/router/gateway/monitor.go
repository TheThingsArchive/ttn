package gateway

import (
	"io"
	"sync"

	"google.golang.org/grpc/metadata"

	context "golang.org/x/net/context"

	pb "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_noc "github.com/TheThingsNetwork/ttn/api/noc"
	pb_router "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
)

func (g *Gateway) monitorContext() (ctx context.Context) {
	return metadata.NewContext(context.Background(), metadata.Pairs(
		"id", g.ID,
		"token", g.Token,
	))
}

type monitorConn struct {
	clients map[string]pb_noc.MonitorClient
	sync.RWMutex

	status struct {
		streams map[string]pb_noc.Monitor_GatewayStatusClient
		sync.RWMutex
	}

	uplink struct {
		streams map[string]pb_noc.Monitor_GatewayUplinkClient
		sync.RWMutex
	}

	downlink struct {
		streams map[string]pb_noc.Monitor_GatewayDownlinkClient
		sync.RWMutex
	}
}

func (g *Gateway) SetMonitors(clients map[string]pb_noc.MonitorClient) {
	g.monitor = NewMonitorConn(clients)
}

func NewMonitorConn(clients map[string]pb_noc.MonitorClient) (conn *monitorConn) {
	conn = &monitorConn{clients: clients}
	conn.uplink.streams = map[string]pb_noc.Monitor_GatewayUplinkClient{}
	conn.downlink.streams = map[string]pb_noc.Monitor_GatewayDownlinkClient{}
	conn.status.streams = map[string]pb_noc.Monitor_GatewayStatusClient{}
	return conn
}

func (g *Gateway) pushStatusToMonitor(ctx log.Interface, name string, status *pb.Status) (err error) {
	defer func() {
		switch err {
		case nil:
			ctx.Info("Pushed status to monitor")

		case io.EOF:
			ctx.Info("Stream closed, retrying")
			g.monitor.status.Lock()
			delete(g.monitor.status.streams, name)
			g.monitor.status.Unlock()
			err = g.pushStatusToMonitor(ctx, name, status)

		default:
			ctx.WithError(errors.FromGRPCError(err)).Warn("Monitor status push failed")
		}
	}()

	g.monitor.status.RLock()
	stream, ok := g.monitor.status.streams[name]
	g.monitor.status.RUnlock()

	if !ok {
		g.monitor.status.Lock()
		if _, ok := g.monitor.status.streams[name]; !ok {
			g.monitor.RLock()
			cl, ok := g.monitor.clients[name]
			if !ok {
				// Should not happen
				return errors.New("Monitor not found")
			}

			if stream, err = cl.GatewayStatus(g.monitorContext()); err != nil {
				// TODO check if err returned is GRPC error
				err = errors.FromGRPCError(err)
				ctx.WithError(err).Warn("Failed to open new status stream")
				return err
			}
			ctx.Info("Opened new status stream")
			g.monitor.status.streams[name] = stream

			g.monitor.RUnlock()
		}
		g.monitor.status.Unlock()
	}

	return stream.Send(status)
}

func (g *Gateway) pushUplinkToMonitor(ctx log.Interface, name string, uplink *pb_router.UplinkMessage) (err error) {
	defer func() {
		switch err {
		case nil:
			ctx.Info("Pushed uplink to monitor")

		case io.EOF:
			ctx.Info("Stream closed, retrying")
			g.monitor.uplink.Lock()
			delete(g.monitor.uplink.streams, name)
			g.monitor.uplink.Unlock()
			err = g.pushUplinkToMonitor(ctx, name, uplink)

		default:
			ctx.WithError(errors.FromGRPCError(err)).Warn("Monitor uplink push failed")
		}
	}()

	g.monitor.uplink.RLock()
	stream, ok := g.monitor.uplink.streams[name]
	g.monitor.uplink.RUnlock()

	if !ok {
		g.monitor.uplink.Lock()
		defer g.monitor.uplink.Unlock()

		if _, ok := g.monitor.uplink.streams[name]; !ok {
			g.monitor.RLock()
			defer g.monitor.RUnlock()

			cl, ok := g.monitor.clients[name]
			if !ok {
				// Should not happen
				return errors.New("Monitor not found")
			}

			if stream, err = cl.GatewayUplink(g.monitorContext()); err != nil {
				// TODO check if err returned is GRPC error
				err = errors.FromGRPCError(err)
				ctx.WithError(err).Warn("Failed to open new uplink stream")
				return err
			}
			g.monitor.uplink.streams[name] = stream

		}
	}

	return stream.Send(uplink)
}

func (g *Gateway) pushDownlinkToMonitor(ctx log.Interface, name string, downlink *pb_router.DownlinkMessage) (err error) {
	defer func() {
		switch err {
		case nil:
			ctx.Info("Pushed downlink to monitor")

		case io.EOF:
			ctx.Info("Stream closed, retrying")
			g.monitor.downlink.Lock()
			delete(g.monitor.downlink.streams, name)
			g.monitor.downlink.Unlock()
			err = g.pushDownlinkToMonitor(ctx, name, downlink)

		default:
			ctx.WithError(errors.FromGRPCError(err)).Warn("Monitor downlink push failed")
		}
	}()

	g.monitor.downlink.RLock()
	stream, ok := g.monitor.downlink.streams[name]
	g.monitor.downlink.RUnlock()

	if !ok {
		g.monitor.downlink.Lock()
		if _, ok := g.monitor.downlink.streams[name]; !ok {
			g.monitor.RLock()
			cl, ok := g.monitor.clients[name]
			if !ok {
				// Should not happen
				return errors.New("Monitor not found")
			}

			if stream, err = cl.GatewayDownlink(g.monitorContext()); err != nil {
				// TODO check if err returned is GRPC error
				err = errors.FromGRPCError(err)
				ctx.WithError(err).Warn("Failed to open new downlink stream")
				return err
			}
			ctx.Info("Opened new downlink stream")
			g.monitor.downlink.streams[name] = stream

			g.monitor.RUnlock()
		}
		g.monitor.downlink.Unlock()
	}

	return stream.Send(downlink)
}
