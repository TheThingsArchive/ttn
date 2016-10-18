// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"crypto/ecdsa"
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/cache"
	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-account-lib/tokenkey"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	pb_monitor "github.com/TheThingsNetwork/ttn/api/monitor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/logging"
	"github.com/TheThingsNetwork/ttn/utils/security"
	"github.com/apex/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type ComponentInterface interface {
	RegisterRPC(s *grpc.Server)
	Init(c *Component) error
	ValidateNetworkContext(ctx context.Context) (*pb_discovery.Announcement, error)
	ValidateTTNAuthContext(ctx context.Context) (*claims.Claims, error)
}

type ManagementInterface interface {
	RegisterManager(s *grpc.Server)
}

// NewComponent creates a new Component
func NewComponent(ctx log.Interface, serviceName string, announcedAddress string) (*Component, error) {
	go func() {
		memstats := new(runtime.MemStats)
		for range time.Tick(time.Minute) {
			runtime.ReadMemStats(memstats)
			ctx.WithFields(log.Fields{
				"Goroutines": runtime.NumGoroutine(),
				"Memory":     float64(memstats.Alloc) / 1000000,
			}).Debugf("Stats")
		}
	}()

	// Disable gRPC tracing
	// SEE: https://github.com/grpc/grpc-go/issues/695
	grpc.EnableTracing = false

	grpclog.SetLogger(logging.NewGRPCLogger(ctx))

	component := &Component{
		Ctx: ctx,
		Identity: &pb_discovery.Announcement{
			Id:             viper.GetString("id"),
			Description:    viper.GetString("description"),
			ServiceName:    serviceName,
			ServiceVersion: fmt.Sprintf("%s-%s (%s)", viper.GetString("version"), viper.GetString("gitCommit"), viper.GetString("buildDate")),
			NetAddress:     announcedAddress,
		},
		AccessToken: viper.GetString("auth-token"),
		TokenKeyProvider: tokenkey.HTTPProvider(
			viper.GetStringMapString("auth-servers"),
			cache.WriteTroughCacheWithFormat(viper.GetString("key-dir"), "auth-%s.pub"),
		),
	}

	if serviceName != "discovery" {
		var err error
		component.Discovery, err = pb_discovery.NewClient(
			viper.GetString("discovery-server"),
			component.Identity,
			func() string {
				token, _ := component.BuildJWT()
				return token
			},
		)
		if err != nil {
			return nil, err
		}
	}

	if priv, err := security.LoadKeypair(viper.GetString("key-dir")); err == nil {
		component.privateKey = priv

		pubPEM, _ := security.PublicPEM(priv)
		component.Identity.PublicKey = string(pubPEM)

		privPEM, _ := security.PrivatePEM(priv)

		if viper.GetBool("tls") {
			cert, err := security.LoadCert(viper.GetString("key-dir"))
			if err != nil {
				return nil, err
			}
			component.Identity.Certificate = string(cert)

			cer, err := tls.X509KeyPair(cert, privPEM)
			if err != nil {
				return nil, err
			}
			component.tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
		}
	}

	if healthPort := viper.GetInt("health-port"); healthPort > 0 {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
			switch component.GetStatus() {
			case StatusHealthy:
				w.WriteHeader(200)
				w.Write([]byte("Status is HEALTHY"))
				return
			case StatusUnhealthy:
				w.WriteHeader(503)
				w.Write([]byte("Status is UNHEALTHY"))
				return
			}
		})
		go http.ListenAndServe(fmt.Sprintf(":%d", healthPort), nil)
	}

	if monitors := viper.GetStringMapString("monitor-servers"); len(monitors) != 0 {
		component.Monitors = make(map[string]pb_monitor.MonitorClient)
		for name, addr := range monitors {
			var err error
			component.Monitors[name], err = pb_monitor.NewClient(addr)
			if err != nil {
				// Assuming grpc.WithBlock() is not set
				return nil, err
			}
		}
	}

	return component, nil
}

// Status indicates the health status of this component
type Status int

const (
	// StatusHealthy indicates a healthy component
	StatusHealthy Status = iota
	// StatusUnhealthy indicates an unhealthy component
	StatusUnhealthy
)

// Component contains the common attributes for all TTN components
type Component struct {
	Identity         *pb_discovery.Announcement
	Discovery        pb_discovery.Client
	Monitors         map[string]pb_monitor.MonitorClient
	Ctx              log.Interface
	AccessToken      string
	privateKey       *ecdsa.PrivateKey
	tlsConfig        *tls.Config
	TokenKeyProvider tokenkey.Provider
	status           int64
}

// GetStatus gets the health status of the component
func (c *Component) GetStatus() Status {
	return Status(atomic.LoadInt64(&c.status))
}

// SetStatus sets the health status of the component
func (c *Component) SetStatus(status Status) {
	atomic.StoreInt64(&c.status, int64(status))
}

// Discover is used to discover another component
func (c *Component) Discover(serviceName, id string) (*pb_discovery.Announcement, error) {
	res, err := c.Discovery.Get(serviceName, id)
	if err != nil {
		return nil, errors.Wrapf(errors.FromGRPCError(err), "Failed to discover %s/%s", serviceName, id)
	}
	return res, nil
}

// Announce the component to TTN discovery
func (c *Component) Announce() error {
	if c.Identity.Id == "" {
		return errors.NewErrInvalidArgument("Component ID", "can not be empty")
	}
	err := c.Discovery.Announce(c.AccessToken)
	if err != nil {
		return errors.Wrapf(errors.FromGRPCError(err), "Failed to announce this component to TTN discovery: %s", err.Error())
	}
	c.Ctx.Info("ttn: Announced to TTN discovery")

	return nil
}

// UpdateTokenKey updates the OAuth Bearer token key
func (c *Component) UpdateTokenKey() error {
	if c.TokenKeyProvider == nil {
		return errors.NewErrInternal("No public key provider configured for token validation")
	}

	// Set up Auth Server Token Validation
	err := c.TokenKeyProvider.Update()
	if err != nil {
		c.Ctx.Warnf("ttn: Failed to refresh public keys for token validation: %s", err.Error())
	} else {
		c.Ctx.Info("ttn: Got public keys for token validation")
	}

	return nil

}

// ValidateNetworkContext validates the context of a network request (router-broker, broker-handler, etc)
func (c *Component) ValidateNetworkContext(ctx context.Context) (component *pb_discovery.Announcement, err error) {
	defer func() {
		if err != nil {
			time.Sleep(time.Second)
		}
	}()

	md, ok := metadata.FromContext(ctx)
	if !ok {
		err = errors.NewErrInternal("Could not get metadata from context")
		return
	}
	var id, serviceName, token string
	if ids, ok := md["id"]; ok && len(ids) == 1 {
		id = ids[0]
	}
	if id == "" {
		err = errors.NewErrInvalidArgument("Metadata", "id missing")
		return
	}
	if serviceNames, ok := md["service-name"]; ok && len(serviceNames) == 1 {
		serviceName = serviceNames[0]
	}
	if serviceName == "" {
		err = errors.NewErrInvalidArgument("Metadata", "service-name missing")
		return
	}
	if tokens, ok := md["token"]; ok && len(tokens) == 1 {
		token = tokens[0]
	}

	var announcement *pb_discovery.Announcement
	announcement, err = c.Discover(serviceName, id)
	if err != nil {
		return
	}

	if announcement.PublicKey == "" {
		return announcement, nil
	}

	if token == "" {
		err = errors.NewErrInvalidArgument("Metadata", "token missing")
		return
	}

	var claims *jwt.StandardClaims
	claims, err = security.ValidateJWT(token, []byte(announcement.PublicKey))
	if err != nil {
		return
	}
	if claims.Issuer != id {
		err = errors.NewErrInvalidArgument("Metadata", "token was issued by different component id")
		return
	}

	return announcement, nil
}

// ValidateTTNAuthContext gets a token from the context and validates it
func (c *Component) ValidateTTNAuthContext(ctx context.Context) (*claims.Claims, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, errors.NewErrInternal("Could not get metadata from context")
	}
	token, ok := md["token"]
	if !ok || len(token) < 1 {
		return nil, errors.NewErrInvalidArgument("Metadata", "token missing")
	}

	if c.TokenKeyProvider == nil {
		return nil, errors.NewErrInternal("No token provider configured")
	}

	if token[0] == "" {
		return nil, errors.NewErrInvalidArgument("Metadata", "token is empty")
	}

	claims, err := claims.FromToken(c.TokenKeyProvider, token[0])
	if err != nil {
		return nil, err
	}

	return claims, nil
}

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

// BuildJWT builds a short-lived JSON Web Token for this component
func (c *Component) BuildJWT() (string, error) {
	if c.privateKey != nil {
		privPEM, err := security.PrivatePEM(c.privateKey)
		if err != nil {
			return "", err
		}
		return security.BuildJWT(c.Identity.Id, 20*time.Second, privPEM)
	}
	return "", nil
}

// GetContext returns a context for outgoing RPC request. If token is "", this function will generate a short lived token from the component
func (c *Component) GetContext(token string) context.Context {
	var serviceName, id, netAddress string
	if c.Identity != nil {
		serviceName = c.Identity.ServiceName
		id = c.Identity.Id
		if token == "" {
			token, _ = c.BuildJWT()
		}
		netAddress = c.Identity.NetAddress
	}
	md := metadata.Pairs(
		"service-name", serviceName,
		"id", id,
		"token", token,
		"net-address", netAddress,
	)
	ctx := metadata.NewContext(context.Background(), md)
	return ctx
}
