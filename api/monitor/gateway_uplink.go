// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"time"

	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/backoff"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
)

func (cl *gatewayClient) initUplink() {
	cl.uplink.ch = make(chan *router.UplinkMessage, BufferSize)
	go cl.monitorUplink()
}

func (cl *gatewayClient) monitorUplink() {
	var retries int

	for {
		ctx, cancel := context.WithCancel(cl.Context())
		cl.uplink.Lock()
		cl.uplink.cancel = cancel
		cl.uplink.Unlock()

		stream, err := cl.client.client.GatewayUplink(ctx)
		if err != nil {
			cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor uplink stream")

			retries++
			time.Sleep(backoff.Backoff(retries))

			continue
		}
		retries = 0
		cl.Ctx.Debug("Opened new monitor uplink stream")

		var tokenChanged bool

		go func() {
			if err := stream.RecvMsg(&empty.Empty{}); err != nil {
				if !tokenChanged {
					cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Received error on monitor uplink stream, closing...")
					stream.CloseSend()
				}
				cl.Ctx.Debug("Closed monitor uplink stream")

				cl.uplink.Lock()
				cl.uplink.cancel()
				cl.uplink.cancel = nil
				cl.uplink.Unlock()
			}
		}()

		for {
			select {
			case <-ctx.Done():
				break
			case <-cl.TokenChanged():
				cl.Ctx.Debug("Restarting uplink stream with new token")
				tokenChanged = true
				stream.CloseSend()
				break
			case uplink, ok := <-cl.uplink.ch:
				if ok {
					if err := stream.Send(uplink); err == nil {
						cl.Ctx.Debug("Sent uplink to monitor")
					}
				}
			}
		}

		retries++
		time.Sleep(backoff.Backoff(retries))
	}
}

func (cl *gatewayClient) closeUplink() {
	cl.uplink.Lock()
	defer cl.uplink.Unlock()
	if cl.uplink.cancel != nil {
		cl.uplink.cancel()
	}
}

// SendUplink sends uplink to the monitor
func (cl *gatewayClient) SendUplink(uplink *router.UplinkMessage) (err error) {
	if !cl.IsConfigured() {
		return nil
	}

	cl.uplink.init.Do(cl.initUplink)

	select {
	case cl.uplink.ch <- uplink:
	default:
		cl.Ctx.Warn("Not sending uplink to monitor, buffer full")
		return errors.New("Not sending uplink to monitor, buffer full")
	}
	return
}
