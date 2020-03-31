// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"math"
	"time"

	"github.com/TheThingsNetwork/api/trace"
	"github.com/TheThingsNetwork/go-utils/grpc/rpcerror"
	"github.com/TheThingsNetwork/go-utils/grpc/rpclog"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func (c *Component) ServerOptions() []grpc.ServerOption {
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(math.MaxUint16),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			rpcerror.UnaryServerInterceptor(errors.BuildGRPCError),
			rpclog.UnaryServerInterceptor(c.Ctx),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_prometheus.StreamServerInterceptor,
			rpcerror.StreamServerInterceptor(errors.BuildGRPCError),
			rpclog.StreamServerInterceptor(c.Ctx),
		)),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              30 * time.Second,
			Timeout:           10 * time.Second,
		}),
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
