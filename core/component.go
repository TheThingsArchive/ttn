// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"path"
	"runtime"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/TheThingsNetwork/ttn/api"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/security"
	"github.com/TheThingsNetwork/ttn/utils/tokenkey"
	"github.com/apex/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/spf13/viper"
)

type ComponentInterface interface {
	RegisterRPC(s *grpc.Server)
	Init(c *Component) error
	ValidateNetworkContext(ctx context.Context) (string, error)
	ValidateTTNAuthContext(ctx context.Context) (*TTNClaims, error)
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

	var discovery pb_discovery.DiscoveryClient
	if serviceName != "discovery" {
		discoveryConn, err := grpc.Dial(viper.GetString("discovery-server"), append(api.DialOptions, grpc.WithBlock(), grpc.WithInsecure())...)
		if err != nil {
			return nil, err
		}
		discovery = pb_discovery.NewDiscoveryClient(discoveryConn)
	}

	component := &Component{
		Ctx: ctx,
		Identity: &pb_discovery.Announcement{
			Id:          viper.GetString("id"),
			Description: viper.GetString("description"),
			ServiceName: serviceName,
			NetAddress:  announcedAddress,
		},
		AccessToken: viper.GetString("auth-token"),
		Discovery:   discovery,
		TokenKeyProvider: tokenkey.NewHTTPProvider(
			fmt.Sprintf("%s/key", viper.GetString("auth-server")),
			path.Join(viper.GetString("key-dir"), "/auth-server.pub"),
		),
	}

	if pub, priv, cert, err := security.LoadKeys(viper.GetString("key-dir")); err == nil {
		component.Identity.PublicKey = string(pub)
		component.privateKey = string(priv)

		if viper.GetBool("tls") {
			component.Identity.Certificate = string(cert)
			cer, err := tls.X509KeyPair(cert, priv)
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
		http.ListenAndServe(fmt.Sprintf(":%d", healthPort), nil)
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
	Discovery        pb_discovery.DiscoveryClient
	Ctx              log.Interface
	AccessToken      string
	privateKey       string
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
	return c.Discovery.Get(c.GetContext(""), &pb_discovery.GetRequest{
		ServiceName: serviceName,
		Id:          id,
	})
}

// Announce the component to TTN discovery
func (c *Component) Announce() error {
	if c.Identity.Id == "" {
		return errors.New("ttn: No ID configured")
	}
	_, err := c.Discovery.Announce(c.GetContext(c.AccessToken), c.Identity)
	if err != nil {
		return fmt.Errorf("ttn: Failed to announce this component to TTN discovery: %s", err.Error())
	}
	c.Ctx.Info("ttn: Announced to TTN discovery")

	return nil
}

// UpdateTokenKey updates the OAuth Bearer token key
func (c *Component) UpdateTokenKey() error {
	if c.TokenKeyProvider == nil {
		return errors.New("ttn: No public key provider configured for token validation")
	}

	// Set up Auth Server Token Validation
	tokenKey, err := c.TokenKeyProvider.Get(true)
	if err != nil {
		c.Ctx.Warnf("ttn: Failed to refresh public key for token validation: %s", err.Error())
	} else {
		c.Ctx.Infof("ttn: Got public key for token validation (%v)", tokenKey.Algorithm)
	}

	return nil

}

// TTNClaims contains the claims that are set by the TTN Token Issuer
type TTNClaims struct {
	jwt.StandardClaims
	Type   string              `json:"type"`
	Client string              `json:"client"`
	Scopes []string            `json:"scope"`
	Apps   map[string][]string `json:"apps,omitempty"`
}

// CanEditApp indicates wheter someone with the claims can manage the given app
func (c *TTNClaims) CanEditApp(appID string) bool {
	for id, rights := range c.Apps {
		if appID == id {
			for _, right := range rights {
				if right == "settings" {
					return true
				}
			}
		}
	}
	return false
}

// ValidateNetworkContext validates the context of a network request (router-broker, broker-handler, etc)
func (c *Component) ValidateNetworkContext(ctx context.Context) (componentID string, err error) {
	defer func() {
		if err != nil {
			time.Sleep(time.Second)
		}
	}()

	md, ok := metadata.FromContext(ctx)
	if !ok {
		err = errors.New("ttn: Could not get metadata")
		return
	}
	var id, serviceName, token string
	if ids, ok := md["id"]; ok && len(ids) == 1 {
		id = ids[0]
	}
	if id == "" {
		err = errors.New("ttn: Could not get id")
		return
	}
	if serviceNames, ok := md["service-name"]; ok && len(serviceNames) == 1 {
		serviceName = serviceNames[0]
	}
	if serviceName == "" {
		err = errors.New("ttn: Could not get service name")
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
		return id, nil
	}

	if token == "" {
		err = errors.New("ttn: Could not get token")
		return
	}

	var claims *jwt.StandardClaims
	claims, err = security.ValidateJWT(token, []byte(announcement.PublicKey))
	if err != nil {
		return
	}
	if claims.Subject != id {
		err = errors.New("The token was issued for a different component ID")
		return
	}

	return id, nil
}

// ValidateTTNAuthContext gets a token from the context and validates it
func (c *Component) ValidateTTNAuthContext(ctx context.Context) (*TTNClaims, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, errors.New("ttn: Could not get metadata")
	}
	token, ok := md["token"]
	if !ok || len(token) < 1 {
		return nil, errors.New("ttn: Could not get token")
	}
	ttnClaims := &TTNClaims{}
	parsed, err := jwt.ParseWithClaims(token[0], ttnClaims, func(token *jwt.Token) (interface{}, error) {
		if c.TokenKeyProvider == nil {
			return nil, errors.New("No token provider configured")
		}
		k, err := c.TokenKeyProvider.Get(false)
		if err != nil {
			return nil, err
		}
		if k.Algorithm != token.Header["alg"] {
			return nil, fmt.Errorf("Expected algorithm %v but got %v", k.Algorithm, token.Header["alg"])
		}
		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(k.Key))
		if err != nil {
			return nil, err
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to parse token: %s", err.Error())
	}
	if !parsed.Valid {
		return nil, errors.New("The token is not valid or is expired")
	}
	return ttnClaims, nil
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
			logCtx.WithError(err).Warn("Could not handle Request")
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
	if c.privateKey != "" {
		return security.BuildJWT(c.Identity.Id, 10*time.Second, []byte(c.privateKey))
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
