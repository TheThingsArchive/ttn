// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"time"

	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/utils/backoff"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
)

func (cl *gatewayClient) initStatus() {
	cl.status.ch = make(chan *gateway.Status, BufferSize)
	go cl.monitorStatus()
}

func (cl *gatewayClient) monitorStatus() {
	var retries int
newStream:
	for {
		ctx, cancel := context.WithCancel(cl.Context())
		cl.status.Lock()
		cl.status.cancel = cancel
		cl.status.Unlock()

		stream, err := cl.client.client.GatewayStatus(ctx)
		if err != nil {
			cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to open new monitor status stream")

			retries++
			time.Sleep(backoff.Backoff(retries))

			continue
		}
		retries = 0
		cl.Ctx.Debug("Opened new monitor status stream")

		// The actual stream
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case status, ok := <-cl.status.ch:
					if ok {
						stream.Send(status)
						cl.Ctx.Debug("Sent status to monitor")
					}
				}
			}
		}()

		msg := new(empty.Empty)
		for {
			if err := stream.RecvMsg(&msg); err != nil {
				cl.Ctx.WithError(errors.FromGRPCError(err)).Warn("Received error on monitor status stream, closing...")
				stream.CloseSend()
				cl.Ctx.Debug("Closed monitor status stream")

				cl.status.Lock()
				cl.status.cancel()
				cl.status.cancel = nil
				cl.status.Unlock()

				retries++
				time.Sleep(backoff.Backoff(retries))

				continue newStream
			}
		}
	}
}

func (cl *gatewayClient) closeStatus() {
	cl.status.Lock()
	defer cl.status.Unlock()
	if cl.status.cancel != nil {
		cl.status.cancel()
	}
}

// SendStatus sends status to the monitor
func (cl *gatewayClient) SendStatus(status *gateway.Status) (err error) {
	if !cl.IsConfigured() {
		return nil
	}

	cl.status.init.Do(cl.initStatus)

	select {
	case cl.status.ch <- status:
	default:
		cl.Ctx.Warn("Not sending status to monitor, buffer full")
		return errors.New("Not sending status to monitor, buffer full")
	}
	return
}
