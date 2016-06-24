// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/TheThingsNetwork/ttn/api"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/tokenkey"
	"github.com/apex/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/spf13/viper"
)

type ComponentInterface interface {
	RegisterRPC(s *grpc.Server)
	Init(c *Component) error
}

type ManagementInterface interface {
	RegisterManager(s *grpc.Server)
}

// NewComponent creates a new Component
func NewComponent(ctx log.Interface, serviceName string, announcedAddress string) *Component {
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

	return &Component{
		Ctx: ctx,
		Identity: &pb_discovery.Announcement{
			Id:          viper.GetString("id"),
			Token:       viper.GetString("token"),
			Description: viper.GetString("description"),
			ServiceName: serviceName,
			NetAddress:  announcedAddress,
		},
		DiscoveryServer: viper.GetString("discovery-server"),
		TokenKeyProvider: tokenkey.NewHTTPProvider(
			fmt.Sprintf("%s/key", viper.GetString("auth-server")),
			viper.GetString("oauth2-keyfile"),
		),
	}
}

// Component contains the common attributes for all TTN components
type Component struct {
	Identity         *pb_discovery.Announcement
	DiscoveryServer  string
	Ctx              log.Interface
	TokenKeyProvider tokenkey.Provider
}

// Announce the component to TTN discovery
func (c *Component) Announce() error {
	if c.DiscoveryServer == "" {
		return errors.New("ttn: No discovery server configured")
	}

	if c.Identity.Id == "" {
		return errors.New("ttn: No ID configured")
	}

	conn, err := grpc.Dial(c.DiscoveryServer, append(api.DialOptions, grpc.WithBlock())...)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb_discovery.NewDiscoveryClient(conn)
	_, err = client.Announce(c.GetContext(), c.Identity)
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

// ValidateContext gets a token from the context and validates it
func (c *Component) ValidateContext(ctx context.Context) (*TTNClaims, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, errors.New("ttn: Could not get metadata")
	}
	token, ok := md["token"]
	if !ok || len(token) < 1 {
		return nil, errors.New("ttn: Could not get token")
	}
	return c.ValidateToken(token[0])
}

// ValidateToken verifies an OAuth Bearer token
func (c *Component) ValidateToken(token string) (*TTNClaims, error) {
	ttnClaims := &TTNClaims{}
	parsed, err := jwt.ParseWithClaims(token, ttnClaims, func(token *jwt.Token) (interface{}, error) {
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
		c.Ctx.WithFields(log.Fields{
			"CallerID": peerID,
			"CallerIP": peerAddr,
			"Method":   info.FullMethod,
		}).Debug("Handle Request")
		return handler(ctx, req)
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
		}).Debug("Start Stream")
		return handler(srv, stream)
	}

	return []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unary)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(stream)),
	}
}

// GetContext returns a context for outgoing RPC requests
func (c *Component) GetContext() context.Context {
	var id, token, netAddress string
	if c.Identity != nil {
		id = c.Identity.Id
		token = c.Identity.Token
		netAddress = c.Identity.NetAddress
	}
	md := metadata.Pairs(
		"id", id,
		"token", token,
		"net-address", netAddress,
	)
	ctx := metadata.NewContext(context.Background(), md)
	return ctx
}
