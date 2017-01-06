// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/cache"
	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-account-lib/keys"
	"github.com/TheThingsNetwork/go-account-lib/oauth"
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
		c.initRoots,
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

type authServer struct {
	url      string
	username string
	password string
}

func parseAuthServer(str string) (srv authServer, err error) {
	url, err := url.Parse(str)
	if err != nil {
		return srv, err
	}
	srv.url = fmt.Sprintf("%s://%s", url.Scheme, url.Host)
	if url.User != nil {
		srv.username = url.User.Username()
		srv.password, _ = url.User.Password()
	}
	return srv, nil
}

func (c *Component) initAuthServers() error {
	urlMap := make(map[string]string)
	funcMap := make(map[string]tokenkey.TokenFunc)
	var httpProvider tokenkey.Provider
	for id, url := range c.Config.AuthServers {
		id, url := id, url // deliberately shadow these
		if strings.HasPrefix(url, "file://") {
			file := strings.TrimPrefix(url, "file://")
			contents, err := ioutil.ReadFile(path.Clean(file))
			if err != nil {
				return err
			}
			funcMap[id] = func(renew bool) (*tokenkey.TokenKey, error) {
				return &tokenkey.TokenKey{Algorithm: "ES256", Key: string(contents)}, nil
			}
			continue
		}
		srv, err := parseAuthServer(url)
		if err != nil {
			return err
		}
		urlMap[id] = srv.url
		funcMap[id] = func(renew bool) (*tokenkey.TokenKey, error) {
			return httpProvider.Get(id, renew)
		}
	}
	httpProvider = tokenkey.HTTPProvider(
		urlMap,
		cache.WriteTroughCacheWithFormat(c.Config.KeyDir, "auth-%s.pub"),
	)
	c.TokenKeyProvider = tokenkey.FuncProvider(funcMap)
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

func (c *Component) initRoots() error {
	path := filepath.Clean(c.Config.KeyDir + "/ca.cert")
	cert, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	if !api.RootCAs.AppendCertsFromPEM(cert) {
		return fmt.Errorf("Could not add root certificates from %s", path)
	}
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
	var serviceName, serviceVersion, id, netAddress string
	if c.Identity != nil {
		serviceName = c.Identity.ServiceName
		id = c.Identity.Id
		if token == "" {
			token, _ = c.BuildJWT()
		}
		serviceVersion = c.Identity.ServiceVersion
		netAddress = c.Identity.NetAddress
	}
	md := metadata.Pairs(
		"service-name", serviceName,
		"service-version", serviceVersion,
		"id", id,
		"token", token,
		"net-address", netAddress,
	)
	ctx := metadata.NewContext(context.Background(), md)
	return ctx
}

var oauthCache = cache.MemoryCache()

// ExchangeAppKeyForToken enables authentication with the App Access Key
func (c *Component) ExchangeAppKeyForToken(appID, key string) (string, error) {
	issuerID := keys.KeyIssuer(key)
	if issuerID == "" {
		// Take the first configured auth server
		for k := range c.Config.AuthServers {
			issuerID = k
			break
		}
		key = fmt.Sprintf("%s.%s", issuerID, key)
	}
	issuer, ok := c.Config.AuthServers[issuerID]
	if !ok {
		return "", fmt.Errorf("Auth server %s not registered", issuer)
	}

	srv, _ := parseAuthServer(issuer)

	oauth := oauth.OAuthWithCache(srv.url, &oauth.Client{
		ID:     srv.username,
		Secret: srv.password,
	}, oauthCache)

	token, err := oauth.ExchangeAppKeyForToken(appID, key)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
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
