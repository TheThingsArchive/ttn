package gateway

import (
	"context"
	"sync"

	pb_noc "github.com/TheThingsNetwork/ttn/api/noc"
)

func (g Gateway) monitorContext() (ctx context.Context) {
	ctx = context.WithValue(context.Background(), "id", g.ID)
	ctx = context.WithValue(ctx, "token", g.Token)
	return ctx
}

type monitorConn struct {
	client pb_noc.MonitorClient

	uplink struct {
		client pb_noc.Monitor_GatewayUplinkClient
		sync.Mutex
	}
	downlink struct {
		client pb_noc.Monitor_GatewayDownlinkClient
		sync.Mutex
	}
	status struct {
		client pb_noc.Monitor_GatewayStatusClient
		sync.Mutex
	}
}

func (g *Gateway) SetMonitor(client pb_noc.MonitorClient) {
	g.monitor = &monitorConn{client: client}
}

func (g *Gateway) uplinkMonitor() (client pb_noc.Monitor_GatewayUplinkClient, err error) {
	g.monitor.uplink.Lock()
	defer g.monitor.uplink.Unlock()

	if g.monitor.uplink.client != nil {
		return g.monitor.uplink.client, nil
	}
	return g.connectUplinkMonitor()
}

func (g *Gateway) downlinkMonitor() (client pb_noc.Monitor_GatewayDownlinkClient, err error) {
	g.monitor.downlink.Lock()
	defer g.monitor.downlink.Unlock()

	if g.monitor.downlink.client != nil {
		return g.monitor.downlink.client, nil
	}
	return g.connectDownlinkMonitor()
}

func (g *Gateway) statusMonitor() (client pb_noc.Monitor_GatewayStatusClient, err error) {
	g.monitor.status.Lock()
	defer g.monitor.status.Unlock()

	if g.monitor.status.client != nil {
		return g.monitor.status.client, nil
	}
	return g.connectStatusMonitor()
}

func (g *Gateway) connectUplinkMonitor() (client pb_noc.Monitor_GatewayUplinkClient, err error) {
	if client, err = connectUplinkMonitor(g.monitorContext(), g.monitor.client); err != nil {
		return nil, err
	}
	g.monitor.uplink.client = client
	return client, nil
}

func (g *Gateway) connectDownlinkMonitor() (client pb_noc.Monitor_GatewayDownlinkClient, err error) {
	if client, err = connectDownlinkMonitor(g.monitorContext(), g.monitor.client); err != nil {
		return nil, err
	}
	g.monitor.downlink.client = client
	return client, nil
}

func (g *Gateway) connectStatusMonitor() (client pb_noc.Monitor_GatewayStatusClient, err error) {
	if client, err = connectStatusMonitor(g.monitorContext(), g.monitor.client); err != nil {
		return nil, err
	}
	g.monitor.status.client = client
	return client, nil
}

func connectDownlinkMonitor(ctx context.Context, client pb_noc.MonitorClient) (pb_noc.Monitor_GatewayDownlinkClient, error) {
	return client.GatewayDownlink(ctx)
}

func connectUplinkMonitor(ctx context.Context, client pb_noc.MonitorClient) (pb_noc.Monitor_GatewayUplinkClient, error) {
	return client.GatewayUplink(ctx)
}

func connectStatusMonitor(ctx context.Context, client pb_noc.MonitorClient) (pb_noc.Monitor_GatewayStatusClient, error) {
	return client.GatewayStatus(ctx)
}
