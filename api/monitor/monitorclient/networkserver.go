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

// NetworkServerClient creates a new client for NetworkServer monitoring
func (m *MonitorClient) NetworkServerClient(ctx context.Context, opts ...grpc.CallOption) Stream {
	c := new(componentClient)
	c.log = m.log.WithField("Component", "NetworkServer")
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if id, err := ttnctx.IDFromMetadata(md); err == nil {
			c.log = c.log.WithField("ID", id)
		}
	}
	c.setup = func() {
		ctx, c.cancel = context.WithCancel(ctx)
		for name, cli := range m.clients {
			status := sendbuffer.New(m.bufferSize, func() (grpc.ClientStream, error) {
				return cli.NetworkServerStatus(ctx, opts...)
			})
			c.status = append(c.status, status)
			go c.run(name, "Status", status)
		}
	}
	c.Open()
	return c
}
