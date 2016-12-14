package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/utils/backoff"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/metadata"
)

type brokerClient struct {
	sync.RWMutex

	client *Client

	Ctx log.Interface

	id, token string

	uplink struct {
		init   sync.Once
		ch     chan *broker.DeduplicatedUplinkMessage
		cancel func()
		sync.Mutex
	}

	downlink struct {
		init   sync.Once
		ch     chan *broker.DownlinkMessage
		cancel func()
		sync.RWMutex
	}
}

// BrokerClient is used as the main client for Brokers to communicate with the monitor
type BrokerClient interface {
	SetToken(token string)
	IsConfigured() bool
	SendUplink(msg *broker.DeduplicatedUplinkMessage) (err error)
	SendDownlink(msg *broker.DownlinkMessage) (err error)
	Close() (err error)
}

func (cl *brokerClient) SetToken(token string) {
	cl.Lock()
	defer cl.Unlock()
	cl.token = token
}

func (cl *brokerClient) IsConfigured() bool {
	cl.RLock()
	defer cl.RUnlock()
	return cl.token != "" && cl.token != "token"
}

// Close closes all opened monitor streams for the broker
func (cl *brokerClient) Close() (err error) {
	cl.closeUplink()
	cl.closeDownlink()
	return err
}

// Context returns monitor connection context for broker
func (cl *brokerClient) Context() (monitorContext context.Context) {
	cl.RLock()
	defer cl.RUnlock()
	return metadata.NewContext(context.Background(), metadata.Pairs(
		"id", cl.id,
		"token", cl.token,
	))
}

func (cl *brokerClient) initUplink() {
	cl.uplink.ch = make(chan *broker.DeduplicatedUplinkMessage, BufferSize)
	go cl.monitorUplink()
}

func (cl *brokerClient) monitorUplink() {
	var retries int
newStream:
	for {
		ctx, cancel := context.WithCancel(cl.Context())
		cl.uplink.Lock()
		cl.uplink.cancel = cancel
		cl.uplink.Unlock()

		stream, err := cl.client.client.BrokerUplink(ctx)
		if err != nil {
			cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor uplink stream")

			retries++
			time.Sleep(backoff.Backoff(retries))

			continue
		}
		retries = 0
		cl.Ctx.Debug("Opened new monitor uplink stream")

		// The actual stream
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case uplink, ok := <-cl.uplink.ch:
					if ok {
						stream.Send(uplink)
						cl.Ctx.Debug("Sent uplink to monitor")
					}
				}
			}
		}()

		msg := new(empty.Empty)
		for {
			if err := stream.RecvMsg(&msg); err != nil {
				cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Received error on monitor uplink stream, closing...")
				stream.CloseSend()
				cl.Ctx.Debug("Closed monitor uplink stream")

				cl.uplink.Lock()
				cl.uplink.cancel()
				cl.uplink.cancel = nil
				cl.uplink.Unlock()

				retries++
				time.Sleep(backoff.Backoff(retries))

				continue newStream
			}
		}
	}
}

func (cl *brokerClient) closeUplink() {
	cl.uplink.Lock()
	defer cl.uplink.Unlock()
	if cl.uplink.cancel != nil {
		cl.uplink.cancel()
	}
}

// SendUplink sends uplink to the monitor
func (cl *brokerClient) SendUplink(uplink *broker.DeduplicatedUplinkMessage) (err error) {
	cl.uplink.init.Do(cl.initUplink)

	select {
	case cl.uplink.ch <- uplink:
	default:
		cl.Ctx.Warn("Not sending uplink to monitor, buffer full")
		return errors.New("Not sending uplink to monitor, buffer full")
	}
	return
}

func (cl *brokerClient) initDownlink() {
	cl.downlink.ch = make(chan *broker.DownlinkMessage, BufferSize)
	go cl.monitorDownlink()
}

func (cl *brokerClient) monitorDownlink() {
	var retries int
newStream:
	for {
		ctx, cancel := context.WithCancel(cl.Context())
		cl.downlink.Lock()
		cl.downlink.cancel = cancel
		cl.downlink.Unlock()

		stream, err := cl.client.client.BrokerDownlink(ctx)
		if err != nil {
			cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor downlink stream")

			retries++
			time.Sleep(backoff.Backoff(retries))

			continue
		}
		retries = 0
		cl.Ctx.Debug("Opened new monitor downlink stream")

		// The actual stream
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case downlink, ok := <-cl.downlink.ch:
					if ok {
						stream.Send(downlink)
						cl.Ctx.Debug("Sent downlink to monitor")
					}
				}
			}
		}()

		msg := new(empty.Empty)
		for {
			if err := stream.RecvMsg(&msg); err != nil {
				cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Received error on monitor downlink stream, closing...")
				stream.CloseSend()
				cl.Ctx.Debug("Closed monitor downlink stream")

				cl.downlink.Lock()
				cl.downlink.cancel()
				cl.downlink.cancel = nil
				cl.downlink.Unlock()

				retries++
				time.Sleep(backoff.Backoff(retries))

				continue newStream
			}
		}
	}
}

func (cl *brokerClient) closeDownlink() {
	cl.downlink.Lock()
	defer cl.downlink.Unlock()
	if cl.downlink.cancel != nil {
		cl.downlink.cancel()
	}
}

// SendDownlink sends downlink to the monitor
func (cl *brokerClient) SendDownlink(downlink *broker.DownlinkMessage) (err error) {
	if !cl.IsConfigured() {
		return nil
	}

	cl.downlink.init.Do(cl.initDownlink)

	select {
	case cl.downlink.ch <- downlink:
	default:
		cl.Ctx.Warn("Not sending downlink to monitor, buffer full")
		return errors.New("Not sending downlink to monitor, buffer full")
	}
	return
}
