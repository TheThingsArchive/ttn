// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitorclient

import (
	"context"

	"github.com/TheThingsNetwork/go-utils/grpc/sendbuffer"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// HandlerClient creates a new client for Handler monitoring
func (m *MonitorClient) HandlerClient(ctx context.Context, opts ...grpc.CallOption) Stream {
	c := new(componentClient)
	c.log = m.log.WithField("Component", "Handler")
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if id, err := ttnctx.IDFromMetadata(md); err == nil {
			c.log = c.log.WithField("ID", id)
		}
	}
	c.setup = func() {
		ctx, c.cancel = context.WithCancel(ctx)
		for name, cli := range m.clients {
			uplink := sendbuffer.New(m.bufferSize, func() (grpc.ClientStream, error) {
				return cli.HandlerUplink(ctx, opts...)
			})
			c.uplink = append(c.uplink, uplink)
			go c.run(name, "Uplink", uplink)

			downlink := sendbuffer.New(m.bufferSize, func() (grpc.ClientStream, error) {
				return cli.HandlerDownlink(ctx, opts...)
			})
			c.downlink = append(c.downlink, downlink)
			go c.run(name, "Downlink", downlink)

			status := sendbuffer.New(m.bufferSize, func() (grpc.ClientStream, error) {
				return cli.HandlerStatus(ctx, opts...)
			})
			c.status = append(c.status, status)
			go c.run(name, "Status", status)
		}
	}
	c.Open()
	return c
}
