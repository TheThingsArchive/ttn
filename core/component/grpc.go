// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"math"

	"github.com/TheThingsNetwork/go-utils/grpc/interceptor"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/fields"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/mwitkow/go-grpc-middleware"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (c *Component) ServerOptions() []grpc.ServerOption {

	unaryLog := interceptor.Unary(func(req interface{}, info *grpc.UnaryServerInfo) (log.Interface, string) {
		return c.Ctx.WithFields(fields.Get(req)), "Request"
	})

	unaryErr := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		iface, err := handler(ctx, req)
		err = errors.BuildGRPCError(err)
		return iface, err
	}

	streamLog := interceptor.Stream(func(srv interface{}, info *grpc.StreamServerInfo) (log.Interface, string) {
		return c.Ctx, "Stream"
	})

	streamErr := func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		err = errors.BuildGRPCError(err)
		return err
	}

	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(math.MaxUint16),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryErr, unaryLog)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamErr, streamLog)),
	}

	if c.tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(c.tlsConfig)))
	}

	return opts
}
