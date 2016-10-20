package monitor

import (
	"context"
	"io"
	"sync"

	"github.com/TheThingsNetwork/ttn/api"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// Client is a wrapper around MonitorClient
type Client struct {
	Log log.Interface

	client MonitorClient
	conn   *grpc.ClientConn
	addr   string

	reopening chan struct{}

	gateways map[string]GatewayClient
	sync.RWMutex
}

// NewClient is a wrapper for NewMonitorClient, initializes
// connection to MonitorServer on monitorAddr with default gRPC options
func NewClient(ctx log.Interface, monitorAddr string) (cl *Client, err error) {
	cl = &Client{
		Log:      ctx,
		addr:     monitorAddr,
		gateways: make(map[string]GatewayClient),
	}
	return cl, cl.Open()
}

func (cl *Client) Open() (err error) {
	cl.Lock()
	defer cl.Unlock()

	return cl.open()
}
func (cl *Client) open() (err error) {
	addr := cl.addr

	ctx := cl.Log.WithField("addr", addr)
	ctx.Debug("Opening monitor connection...")

	cl.conn, err = grpc.Dial(addr, append(api.DialOptions, grpc.WithInsecure())...)
	if err != nil {
		ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to establish connection")
		return err
	}
	ctx.Debug("Connection established")

	cl.client = NewMonitorClient(cl.conn)
	return nil
}

// Close closes connection to the monitor
func (cl *Client) Close() (err error) {
	cl.Lock()
	defer cl.Unlock()

	return cl.close()
}
func (cl *Client) close() (err error) {
	cl.Log.Debug("Closing monitor connection...")
	for _, gtw := range cl.gateways {
		err = gtw.Close()
		if err != nil {
			cl.Log.WithError(err).WithField("GatewayID", gtw.(*gatewayClient).id).Warn("Failed to close streams")
		}
	}

	err = cl.conn.Close()
	if err != nil {
		return err
	}

	cl.conn = nil
	return nil
}

func (cl *Client) Reopen() (err error) {
	cl.Lock()
	defer cl.Unlock()

	return cl.reopen()
}
func (cl *Client) reopen() (err error) {
	cl.Log.Debug("Reopening monitor connection...")

	cl.reopening = make(chan struct{})
	defer func() {
		close(cl.reopening)
		cl.reopening = nil
	}()

	err = cl.close()
	if err != nil {
		return err
	}
	return cl.open()
}

func (cl *Client) IsReopening() bool {
	return cl.reopening != nil
}

func (cl *Client) IsConnected() bool {
	return cl.client != nil && cl.conn != nil
}

func (cl *Client) GatewayClient(id, token string) (gtwCl GatewayClient) {
	cl.RLock()
	gtwCl, ok := cl.gateways[id]
	cl.RUnlock()
	if !ok {
		cl.Lock()
		gtwCl = &gatewayClient{
			Log: cl.Log.WithField("GatewayID", id),

			client: cl,

			id:    id,
			token: token,
		}
		cl.gateways[id] = gtwCl
		cl.Unlock()
	}
	return gtwCl
}

type gatewayClient struct {
	client *Client

	Log log.Interface

	id, token string

	status struct {
		stream Monitor_GatewayStatusClient
		sync.RWMutex
	}

	uplink struct {
		stream Monitor_GatewayUplinkClient
		sync.RWMutex
	}

	downlink struct {
		stream Monitor_GatewayDownlinkClient
		sync.RWMutex
	}
}

// GatewayClient is used as the main client for Gateways to communicate with the Router
type GatewayClient interface {
	SendStatus(status *pb_gateway.Status) (err error)
	SendUplink(msg *router.UplinkMessage) (err error)
	SendDownlink(msg *router.DownlinkMessage) (err error)
	Close() (err error)
}

func (cl *gatewayClient) SendStatus(status *pb_gateway.Status) (err error) {
	defer func() {
		if err != nil {
			cl.Log.WithError(errors.FromGRPCError(err)).Warn("Failed to send status to monitor")

			if code := grpc.Code(err); code == codes.Unavailable || code == codes.Internal {
				cl.client.Lock()
				defer cl.client.Unlock()

				if !cl.client.IsReopening() {
					err = cl.client.reopen()
				}
			}
		} else {
			cl.Log.Debug("Sent status to monitor")
		}
	}()

	cl.status.RLock()
	stream := cl.status.stream
	cl.status.RUnlock()

	if stream == nil {
		cl.status.Lock()
		if stream = cl.status.stream; stream == nil {
			stream, err = cl.setupStatus()
			if err != nil {
				cl.status.Unlock()
				cl.Log.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor status stream")
				return err
			}

			cl.Log.Debug("Opened new monitor status stream")
		}
		cl.status.Unlock()
	}

	if err = stream.Send(status); err == io.EOF {
		cl.Log.Warn("Monitor status stream closed")
		cl.status.Lock()
		if cl.status.stream == stream {
			cl.status.stream = nil
		}
		cl.status.Unlock()
		return nil
	}
	return err
}

func (cl *gatewayClient) SendUplink(uplink *router.UplinkMessage) (err error) {
	defer func() {
		if err != nil {
			cl.Log.WithError(errors.FromGRPCError(err)).Warn("Failed to send uplink to monitor")

			if code := grpc.Code(err); code == codes.Unavailable || code == codes.Internal {
				cl.Log.Debug("error is internal")
				cl.Log.Debug("Locking...")
				cl.client.Lock()
				cl.Log.Debug("Locked...")

				cl.Log.Debug("Check if reopening...")
				if !cl.client.IsReopening() {
					cl.Log.Debug("Not reopening...")
					err = cl.client.reopen()
				}
				cl.Log.Debug("Unlocking...")
				cl.client.Unlock()
				cl.Log.Debug("Unlocked...")
				cl.Log.Debugf("return %s", err)
			}
		} else {
			cl.Log.Debug("Sent uplink to monitor")
		}
	}()

	cl.uplink.RLock()
	stream := cl.uplink.stream
	cl.uplink.RUnlock()

	if stream == nil {
		cl.uplink.Lock()
		if stream = cl.uplink.stream; stream == nil {
			stream, err = cl.setupUplink()
			if err != nil {
				cl.uplink.Unlock()
				cl.Log.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor uplink stream")
				return err
			}

			cl.Log.Debug("Opened new monitor uplink stream")
		}
		cl.uplink.Unlock()
	}

	if err = stream.Send(uplink); err == io.EOF {
		cl.Log.Warn("Monitor uplink stream closed")
		cl.uplink.Lock()
		if cl.uplink.stream == stream {
			cl.uplink.stream = nil
		}
		cl.uplink.Unlock()
		return nil
	}
	return err
}

func (cl *gatewayClient) SendDownlink(downlink *router.DownlinkMessage) (err error) {
	defer func() {
		if err != nil {
			cl.Log.WithError(errors.FromGRPCError(err)).Warn("Failed to send downlink to monitor")

			if code := grpc.Code(err); code == codes.Unavailable || code == codes.Internal {
				cl.client.Lock()
				defer cl.client.Unlock()

				if !cl.client.IsReopening() {
					err = cl.client.reopen()
				}
			}
		} else {
			cl.Log.Debug("Sent downlink to monitor")
		}
	}()

	cl.downlink.RLock()
	stream := cl.downlink.stream
	cl.downlink.RUnlock()

	if stream == nil {
		cl.downlink.Lock()
		if stream = cl.downlink.stream; stream == nil {
			stream, err = cl.setupDownlink()
			if err != nil {
				cl.downlink.Unlock()
				cl.Log.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor downlink stream")
				return err
			}

			cl.Log.Debug("Opened new monitor downlink stream")
		}
		cl.downlink.Unlock()
	}

	if err = stream.Send(downlink); err == io.EOF {
		cl.Log.Warn("Monitor downlink stream closed")
		cl.downlink.Lock()
		if cl.downlink.stream == stream {
			cl.downlink.stream = nil
		}
		cl.downlink.Unlock()
		return nil
	}
	return err
}

func (cl *gatewayClient) Close() (err error) {
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		cl.Log.Debug("Status locking...")
		cl.status.Lock()
		cl.Log.Debug("Status locked")

		if cl.status.stream != nil {
			if cerr := cl.status.stream.CloseSend(); cerr != nil {
				cl.Log.WithError(cerr).Warn("Failed to close status stream")
				err = cerr
			}
			cl.status.stream = nil
		}
	}()
	defer cl.status.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		cl.Log.Debug("Uplink locking...")
		cl.uplink.Lock()
		cl.Log.Debug("Uplink locked")

		if cl.uplink.stream != nil {
			if cerr := cl.uplink.stream.CloseSend(); cerr != nil {
				cl.Log.WithError(cerr).Warn("Failed to close uplink stream")
				err = cerr
			}
			cl.uplink.stream = nil
		}
	}()
	defer cl.uplink.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		cl.downlink.Lock()

		if cl.downlink.stream != nil {
			if cerr := cl.downlink.stream.CloseSend(); cerr != nil {
				cl.Log.WithError(cerr).Warn("Failed to close downlink stream")
				err = cerr
			}
			cl.downlink.stream = nil
		}
	}()
	defer cl.downlink.Unlock()

	wg.Wait()
	return err
}

func (cl *gatewayClient) Context() (monitorContext context.Context) {
	return metadata.NewContext(context.Background(), metadata.Pairs(
		"id", cl.id,
		"token", cl.token,
	))
}

func (cl *gatewayClient) setupStatus() (stream Monitor_GatewayStatusClient, err error) {
	stream, err = cl.client.client.GatewayStatus(cl.Context())
	if err != nil {
		return nil, err
	}

	cl.status.stream = stream
	return stream, nil
}

func (cl *gatewayClient) setupUplink() (stream Monitor_GatewayUplinkClient, err error) {
	stream, err = cl.client.client.GatewayUplink(cl.Context())
	if err != nil {
		return nil, err
	}

	cl.uplink.stream = stream
	return stream, nil
}

func (cl *gatewayClient) setupDownlink() (stream Monitor_GatewayDownlinkClient, err error) {
	stream, err = cl.client.client.GatewayDownlink(cl.Context())
	if err != nil {
		return nil, err
	}

	cl.downlink.stream = stream
	return stream, nil
}
