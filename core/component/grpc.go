package component

import (
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
	"github.com/mwitkow/go-grpc-middleware"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func (c *Component) ServerOptions() []grpc.ServerOption {
	unary := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var peerAddr string
		peer, ok := peer.FromContext(ctx)
		if ok {
			peerAddr = peer.Addr.String()
		}
		var peerID string
		meta, ok := metadata.FromContext(ctx)
		if ok {
			id, ok := meta["id"]
			if ok && len(id) > 0 {
				peerID = id[0]
			}
		}
		logCtx := c.Ctx.WithFields(log.Fields{
			"CallerID": peerID,
			"CallerIP": peerAddr,
			"Method":   info.FullMethod,
		})
		t := time.Now()
		iface, err := handler(ctx, req)
		logCtx = logCtx.WithField("Duration", time.Now().Sub(t))
		if err != nil {
			err := errors.FromGRPCError(err)
			logCtx.WithField("error", err.Error()).Warn("Could not handle Request")
		} else {
			logCtx.Info("Handled request")
		}
		return iface, err
	}

	stream := func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var peerAddr string
		peer, ok := peer.FromContext(stream.Context())
		if ok {
			peerAddr = peer.Addr.String()
		}
		var peerID string
		meta, ok := metadata.FromContext(stream.Context())
		if ok {
			id, ok := meta["id"]
			if ok && len(id) > 0 {
				peerID = id[0]
			}
		}
		c.Ctx.WithFields(log.Fields{
			"CallerID": peerID,
			"CallerIP": peerAddr,
			"Method":   info.FullMethod,
		}).Info("Start stream")
		return handler(srv, stream)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unary)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(stream)),
	}

	if c.tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(c.tlsConfig)))
	}

	return opts
}
