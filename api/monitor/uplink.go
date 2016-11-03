// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"time"

	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/backoff"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context"
)

func (cl *gatewayClient) initUplink() {
	cl.uplink.ch = make(chan *router.UplinkMessage, BufferSize)
	go cl.monitorUplink()
}

func (cl *gatewayClient) monitorUplink() {
	var retries int
newStream:
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

		var msg []byte
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
