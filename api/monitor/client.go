package monitor

import (
	"io"
	"sync"

	"golang.org/x/net/context"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// Client is a wrapper around MonitorClient
type Client struct {
	Ctx log.Interface

	client MonitorClient
	conn   *grpc.ClientConn
	addr   string

	once *sync.Once

	gateways map[string]GatewayClient
	mutex    sync.RWMutex
}

// NewClient is a wrapper for NewMonitorClient, initializes
// connection to MonitorServer on monitorAddr with default gRPC options
func NewClient(ctx log.Interface, monitorAddr string) (cl *Client, err error) {
	cl = &Client{
		Ctx:      ctx,
		addr:     monitorAddr,
		gateways: make(map[string]GatewayClient),

		once: &sync.Once{},
	}
	return cl, cl.Open()
}

// Open opens connection to the monitor
func (cl *Client) Open() (err error) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	return cl.open()
}
func (cl *Client) open() (err error) {
	addr := cl.addr
	ctx := cl.Ctx.WithField("addr", addr)

	defer func() {
		if err != nil {
			ctx.Warn("Failed to open monitor connection")
		} else {
			ctx.Info("Monitor connection opened")
		}
	}()

	ctx.Debug("Opening monitor connection...")

	cl.conn, err = grpc.Dial(addr, append(api.DialOptions, grpc.WithInsecure())...)
	if err != nil {
		ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to establish connection to gRPC service")
		return err
	}

	cl.client = NewMonitorClient(cl.conn)
	return nil
}

// Close closes connection to the monitor
func (cl *Client) Close() (err error) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	return cl.close()
}
func (cl *Client) close() (err error) {
	defer func() {
		if err != nil {
			cl.Ctx.Warn("Failed to close monitor connection")
		} else {
			cl.Ctx.Info("Monitor connection closed")
		}
	}()

	for _, gtw := range cl.gateways {
		ctx := cl.Ctx.WithField("GatewayID", gtw.(*gatewayClient).id)

		ctx.Debug("Closing gateway streams...")
		err = gtw.Close()
		if err != nil {
			ctx.Warn("Failed to close gateway streams")
		}
	}

	cl.Ctx.Debug("Closing monitor connection...")
	err = cl.conn.Close()
	if err != nil {
		return err
	}

	cl.conn = nil
	return nil
}

// Reopen reopens connection to the monitor. It first attempts to close already opened connection
// and then opens a new one. If closing already opened connection fails, Reopen fails too.
func (cl *Client) Reopen() (err error) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	return cl.reopen()
}
func (cl *Client) reopen() (err error) {
	defer func() {
		if err != nil {
			cl.Ctx.Warn("Failed to reopen monitor connection")
		} else {
			cl.Ctx.Info("Monitor connection reopened")
		}
	}()

	cl.Ctx.Debug("Reopening monitor connection...")

	err = cl.close()
	if err != nil {
		return err
	}
	return cl.open()
}

// IsConnected returns whether connection to the monitor had been established or not
func (cl *Client) IsConnected() bool {
	return cl.client != nil && cl.conn != nil
}

// GatewayClient returns monitor GatewayClient for id and token specified
func (cl *Client) GatewayClient(id string) (gtwCl GatewayClient) {
	cl.mutex.RLock()
	gtwCl, ok := cl.gateways[id]
	cl.mutex.RUnlock()
	if !ok {
		cl.mutex.Lock()
		gtwCl = &gatewayClient{
			Ctx:    cl.Ctx.WithField("GatewayID", id),
			client: cl,
			id:     id,
		}
		cl.gateways[id] = gtwCl
		cl.mutex.Unlock()
	}
	return gtwCl
}

type gatewayClient struct {
	sync.RWMutex

	client *Client

	Ctx log.Interface

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

// GatewayClient is used as the main client for Gateways to communicate with the monitor
type GatewayClient interface {
	SetToken(token string)
	SendStatus(status *gateway.Status) (err error)
	SendUplink(msg *router.UplinkMessage) (err error)
	SendDownlink(msg *router.DownlinkMessage) (err error)
	Close() (err error)
}

func (cl *gatewayClient) SetToken(token string) {
	cl.Lock()
	defer cl.Unlock()
	cl.token = token
}

func (cl *gatewayClient) IsConfigured() bool {
	cl.RLock()
	defer cl.RUnlock()
	return cl.token != ""
}

// SendStatus sends status to the monitor
func (cl *gatewayClient) SendStatus(status *gateway.Status) (err error) {
	if !cl.IsConfigured() {
		return nil
	}

	cl.status.RLock()
	cl.client.mutex.RLock()

	once := cl.client.once
	stream := cl.status.stream

	cl.status.RUnlock()
	cl.client.mutex.RUnlock()

	defer func() {
		if err != nil {
			cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to send status to monitor")

			if code := grpc.Code(err); code == codes.Unavailable || code == codes.Internal {
				once.Do(func() {
					err = cl.client.Reopen()

					cl.client.mutex.Lock()
					cl.client.once = &sync.Once{}
					cl.client.mutex.Unlock()
				})
			}
		} else {
			cl.Ctx.Debug("Sent status to monitor")
		}
	}()

	if stream == nil {
		cl.status.Lock()
		if stream = cl.status.stream; stream == nil {
			stream, err = cl.setupStatus()
			if err != nil {
				cl.status.Unlock()
				return err
			}
		}
		go func() {
			var msg []byte
			if err := stream.RecvMsg(&msg); err != nil {
				cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Received error on monitor status stream, closing...")
				cl.status.Lock()
				cl.status.stream.CloseSend()
				if cl.status.stream == stream {
					cl.status.stream = nil
				}
				cl.status.Unlock()
			}
		}()
		cl.status.Unlock()
	}

	if err = stream.Send(status); err == io.EOF {
		cl.Ctx.Warn("Monitor status stream closed")
		cl.status.Lock()
		if cl.status.stream == stream {
			cl.status.stream = nil
		}
		cl.status.Unlock()
		return nil
	}
	return err
}

// SendUplink sends uplink to the monitor
func (cl *gatewayClient) SendUplink(uplink *router.UplinkMessage) (err error) {
	if !cl.IsConfigured() {
		return nil
	}

	cl.uplink.RLock()
	cl.client.mutex.RLock()

	once := cl.client.once
	stream := cl.uplink.stream

	cl.uplink.RUnlock()
	cl.client.mutex.RUnlock()

	defer func() {
		if err != nil {
			cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to send uplink to monitor")

			if code := grpc.Code(err); code == codes.Unavailable || code == codes.Internal {
				once.Do(func() {
					err = cl.client.Reopen()

					cl.client.mutex.Lock()
					cl.client.once = &sync.Once{}
					cl.client.mutex.Unlock()
				})
			}
		} else {
			cl.Ctx.Debug("Sent uplink to monitor")
		}
	}()

	if stream == nil {
		cl.uplink.Lock()
		if stream = cl.uplink.stream; stream == nil {
			stream, err = cl.setupUplink()
			if err != nil {
				cl.uplink.Unlock()
				return err
			}
		}
		go func() {
			var msg []byte
			if err := stream.RecvMsg(&msg); err != nil {
				cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Received error on monitor uplink stream, closing...")
				cl.uplink.Lock()
				cl.uplink.stream.CloseSend()
				if cl.uplink.stream == stream {
					cl.uplink.stream = nil
				}
				cl.uplink.Unlock()
			}
		}()
		cl.uplink.Unlock()
	}

	if err = stream.Send(uplink); err == io.EOF {
		cl.Ctx.Warn("Monitor uplink stream closed")
		cl.uplink.Lock()
		if cl.uplink.stream == stream {
			cl.uplink.stream = nil
		}
		cl.uplink.Unlock()
		return nil
	}
	return err
}

// SendUplink sends downlink to the monitor
func (cl *gatewayClient) SendDownlink(downlink *router.DownlinkMessage) (err error) {
	if !cl.IsConfigured() {
		return nil
	}

	cl.downlink.RLock()
	cl.client.mutex.RLock()

	once := cl.client.once
	stream := cl.downlink.stream

	cl.downlink.RUnlock()
	cl.client.mutex.RUnlock()

	defer func() {
		if err != nil {
			cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to send downlink to monitor")

			if code := grpc.Code(err); code == codes.Unavailable || code == codes.Internal {
				once.Do(func() {
					err = cl.client.Reopen()

					cl.client.mutex.Lock()
					cl.client.once = &sync.Once{}
					cl.client.mutex.Unlock()
				})
			}
		} else {
			cl.Ctx.Debug("Sent downlink to monitor")
		}
	}()

	if stream == nil {
		cl.downlink.Lock()
		if stream = cl.downlink.stream; stream == nil {
			stream, err = cl.setupDownlink()
			if err != nil {
				cl.downlink.Unlock()
				return err
			}
		}
		go func() {
			var msg []byte
			if err := stream.RecvMsg(&msg); err != nil {
				cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Received error on monitor downlink stream, closing...")
				cl.downlink.Lock()
				cl.downlink.stream.CloseSend()
				if cl.downlink.stream == stream {
					cl.downlink.stream = nil
				}
				cl.downlink.Unlock()
			}
		}()
		cl.downlink.Unlock()
	}

	if err = stream.Send(downlink); err == io.EOF {
		cl.Ctx.Warn("Monitor downlink stream closed")
		cl.downlink.Lock()
		if cl.downlink.stream == stream {
			cl.downlink.stream = nil
		}
		cl.downlink.Unlock()
		return nil
	}
	return err
}

// Close closes all opened monitor streams for the gateway
func (cl *gatewayClient) Close() (err error) {
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		cl.status.Lock()

		if cl.status.stream != nil {
			if cerr := cl.closeStatus(); cerr != nil {
				err = cerr
			}
			cl.status.stream = nil
		}
	}()
	defer cl.status.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		cl.uplink.Lock()

		if cl.uplink.stream != nil {
			if cerr := cl.closeUplink(); cerr != nil {
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
			cerr := cl.closeDownlink()
			if cerr != nil {
				err = cerr
			}
			cl.downlink.stream = nil
		}
	}()
	defer cl.downlink.Unlock()

	wg.Wait()
	return err
}

// Context returns monitor connection context for gateway
func (cl *gatewayClient) Context() (monitorContext context.Context) {
	cl.RLock()
	defer cl.RUnlock()
	return metadata.NewContext(context.Background(), metadata.Pairs(
		"id", cl.id,
		"token", cl.token,
	))
}

func (cl *gatewayClient) setupStatus() (stream Monitor_GatewayStatusClient, err error) {
	stream, err = cl.client.client.GatewayStatus(cl.Context())
	if err != nil {
		cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor status stream")
		return nil, err
	}
	cl.Ctx.Debug("Opened new monitor status stream")

	cl.status.stream = stream
	return stream, nil
}
func (cl *gatewayClient) setupUplink() (stream Monitor_GatewayUplinkClient, err error) {
	stream, err = cl.client.client.GatewayUplink(cl.Context())
	if err != nil {
		cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor uplink stream")
		return nil, err
	}
	cl.Ctx.Debug("Opened new monitor uplink stream")

	cl.uplink.stream = stream
	return stream, nil
}
func (cl *gatewayClient) setupDownlink() (stream Monitor_GatewayDownlinkClient, err error) {
	stream, err = cl.client.client.GatewayDownlink(cl.Context())
	if err != nil {
		cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor downlink stream")
		return nil, err
	}
	cl.Ctx.Debug("Opened new monitor downlink stream")

	cl.downlink.stream = stream
	return stream, nil
}

func (cl *gatewayClient) closeStatus() (err error) {
	err = cl.status.stream.CloseSend()
	if err != nil {
		cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to close status stream")
	}
	cl.Ctx.Debug("Closed status stream")

	return err
}
func (cl *gatewayClient) closeUplink() (err error) {
	err = cl.uplink.stream.CloseSend()
	if err != nil {
		cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to close uplink stream")
	}
	cl.Ctx.Debug("Closed uplink stream")

	return err
}
func (cl *gatewayClient) closeDownlink() (err error) {
	err = cl.downlink.stream.CloseSend()
	if err != nil {
		cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to close downlink stream")
	}
	cl.Ctx.Debug("Closed downlink stream")

	return err
}
