// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitorclient

import (
	"context"

	"github.com/TheThingsNetwork/go-utils/grpc/sendbuffer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// BrokerClient creates a new client for broker monitoring
func (m *MonitorClient) BrokerClient(ctx context.Context, opts ...grpc.CallOption) Stream {
	c := new(componentClient)
	c.log = m.log.WithField("Component", "Broker")
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if id, ok := md["id"]; ok && len(id) > 0 {
			c.log = c.log.WithField("ID", id[0])
		}
	}
	c.setup = func() {
		ctx, c.cancel = context.WithCancel(ctx)
		for name, cli := range m.clients {
			name, cli := name, cli // shadow vars

			uplink := sendbuffer.New(m.bufferSize, func() (grpc.ClientStream, error) {
				return cli.BrokerUplink(ctx, opts...)
			})
			c.uplink = append(c.uplink, uplink)
			go c.run(name, "Uplink", uplink)

			downlink := sendbuffer.New(m.bufferSize, func() (grpc.ClientStream, error) {
				return cli.BrokerDownlink(ctx, opts...)
			})
			c.downlink = append(c.downlink, downlink)
			go c.run(name, "Downlink", downlink)

			status := sendbuffer.New(m.bufferSize, func() (grpc.ClientStream, error) {
				return cli.BrokerStatus(ctx, opts...)
			})
			c.status = append(c.status, status)
			go c.run(name, "Status", status)
		}
	}
	c.Open()
	return c
}
