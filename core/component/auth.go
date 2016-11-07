package component

import (
	"crypto/tls"
	"time"

	"regexp"

	"github.com/TheThingsNetwork/go-account-lib/cache"
	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-account-lib/tokenkey"
	"github.com/TheThingsNetwork/ttn/api"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/security"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// InitAuth initializes Auth functionality
func (c *Component) InitAuth() error {
	inits := []func() error{
		c.initAuthServers,
		c.initKeyPair,
	}
	if c.Config.UseTLS {
		inits = append(inits, c.initTLS)
	}

	for _, init := range inits {
		if err := init(); err != nil {
			return err
		}
	}

	return nil
}

// AuthServerRegex gives the format of auth server configuration.
// Format: [username[:password]@]domain
// - usernames can contain lowercase letters, numbers, underscores and dashes
// - passwords can contain uppercase and lowercase letters, numbers, and special characters
// - domains can be http/https and can contain lowercase letters, numbers, dashes and dots
var AuthServerRegex = regexp.MustCompile(`^(http[s]?://)(?:([0-9a-z_-]+)(?::([0-9A-Za-z-!"#$%&'()*+,.:;<=>?@[\]^_{|}~]+))?@)?([0-9a-z.-]+)/?$`)

// ErrNoAuthServerRegexMatch is returned when an auth server
var ErrNoAuthServerRegexMatch = errors.New("Account server did not match AuthServerRegex")

func (c *Component) initAuthServers() error {
	urlMap := make(map[string]string)
	for id, url := range c.Config.AuthServers {
		matches := AuthServerRegex.FindStringSubmatch(url)
		if len(matches) == 0 {
			return ErrNoAuthServerRegexMatch
		}
		urlMap[id] = matches[1] + matches[4]
	}
	c.TokenKeyProvider = tokenkey.HTTPProvider(
		urlMap,
		cache.WriteTroughCacheWithFormat(c.Config.KeyDir, "auth-%s.pub"),
	)
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

func (c *Component) initKeyPair() error {
	priv, err := security.LoadKeypair(c.Config.KeyDir)
	if err != nil {
		return err
	}
	c.privateKey = priv

	pubPEM, _ := security.PublicPEM(priv)
	c.Identity.PublicKey = string(pubPEM)

	return nil
}

func (c *Component) initTLS() error {
	cert, err := security.LoadCert(c.Config.KeyDir)
	if err != nil {
		return err
	}
	c.Identity.Certificate = string(cert)

	privPEM, _ := security.PrivatePEM(c.privateKey)
	cer, err := tls.X509KeyPair(cert, privPEM)
	if err != nil {
		return err
	}

	c.tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	return nil
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
	token, err := api.TokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if c.TokenKeyProvider == nil {
		return nil, errors.NewErrInternal("No token provider configured")
	}

	claims, err := claims.FromToken(c.TokenKeyProvider, token)
	if err != nil {
		return nil, errors.NewErrPermissionDenied(err.Error())
	}

	return claims, nil
}
