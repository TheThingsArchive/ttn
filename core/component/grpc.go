// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"math"

	"github.com/TheThingsNetwork/api/trace"
	"github.com/TheThingsNetwork/go-utils/grpc/rpcerror"
	"github.com/TheThingsNetwork/go-utils/grpc/rpclog"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/mwitkow/go-grpc-middleware" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (c *Component) ServerOptions() []grpc.ServerOption {
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(math.MaxUint16),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			rpcerror.UnaryServerInterceptor(errors.BuildGRPCError),
			rpclog.UnaryServerInterceptor(c.Ctx),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			rpcerror.StreamServerInterceptor(errors.BuildGRPCError),
			rpclog.StreamServerInterceptor(c.Ctx),
		)),
	}
	if c.tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(c.tlsConfig)))
	}
	return opts
}

func init() {
	// Disable gRPC tracing
	// SEE: https://github.com/grpc/grpc-go/issues/695
	grpc.EnableTracing = false

	// Initialize TTN tracing
	OnInitialize(func(c *Component) error {
		trace.SetComponent(c.Identity.ServiceName, c.Identity.ID)
		return nil
	})
}
