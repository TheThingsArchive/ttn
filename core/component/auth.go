package component

import (
	"time"

	"github.com/TheThingsNetwork/go-account-lib/claims"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/security"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

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
		return nil, errors.NewErrPermissionDenied(err.Error())
	}

	return claims, nil
}
