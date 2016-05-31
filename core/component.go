package core

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/TheThingsNetwork/ttn/api"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/tokenkey"
	"github.com/apex/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

type ComponentInterface interface {
	RegisterRPC(s *grpc.Server)
	Init(c *Component) error
}

// NewComponent creates a new Component
func NewComponent(ctx log.Interface, serviceName string, announcedAddress string) *Component {
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
	_, err = client.Announce(context.Background(), c.Identity)
	if err != nil {
		return fmt.Errorf("ttn: Failed to announce this component to TTN discovery: %s", err.Error())
	}
	c.Ctx.Info("ttn: Announced to TTN discovery")

	return nil
}

// UpdateTokenKey updates the OAuth Bearer token key
func (c *Component) UpdateTokenKey() error {
	if c.TokenKeyProvider == nil {
		return errors.New("No token provider configured")
	}

	// Set up Auth Server Token Validation
	tokenKey, err := c.TokenKeyProvider.Get(true)
	if err != nil {
		return fmt.Errorf("ttn: Failed to refresh token key: %s", err.Error())
	}
	c.Ctx.Infof("ttn: Got token key for algorithm %v", tokenKey.Algorithm)

	return nil

}

// ValidateToken verifies an OAuth Bearer token
func (c *Component) ValidateToken(token string) (claims map[string]interface{}, err error) {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
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
		return []byte(k.Key), nil
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to parse token: %s", err.Error())
	}
	if !parsed.Valid {
		return nil, errors.New("The token is not valid or is expired")
	}
	return parsed.Claims, nil
}

// GetContext returns a context for outgoing RPC requests
func (c *Component) GetContext() context.Context {
	var id, token string
	if c.Identity != nil {
		id = c.Identity.Id
		token = c.Identity.Token
	}
	md := metadata.Pairs(
		"token", token,
		"id", id,
	)
	ctx := metadata.NewContext(context.Background(), md)
	return ctx
}
