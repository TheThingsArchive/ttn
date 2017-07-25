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

	pb_discovery "github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/go-account-lib/auth"
	"github.com/TheThingsNetwork/go-account-lib/cache"
	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-account-lib/keys"
	"github.com/TheThingsNetwork/go-account-lib/tokenkey"
	api_auth "github.com/TheThingsNetwork/go-utils/grpc/auth"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/security"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

// InitAuth initializes Auth functionality
func (c *Component) InitAuth() error {
	inits := []func() error{
		c.initAuthServers,
		c.initKeyPair,
		c.initRoots,
		c.initBgCtx,
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

	if c.Pool != nil {
		c.Pool.AddDialOption(api_auth.WithTokenFunc("target-id", func(_ string) string {
			token, _ := c.BuildJWT()
			return token
		}).DialOption())
	}

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
	if !pool.RootCAs.AppendCertsFromPEM(cert) {
		return fmt.Errorf("Could not add root certificates from %s", path)
	}
	return nil
}

func (c *Component) initBgCtx() error {
	ctx := context.Background()
	if c.Identity != nil {
		ctx = ttnctx.OutgoingContextWithID(ctx, c.Identity.ID)
		ctx = ttnctx.OutgoingContextWithServiceInfo(ctx, c.Identity.ServiceName, c.Identity.ServiceVersion, c.Identity.NetAddress)
	}
	c.Context = ctx
	if c.Pool != nil {
		c.Pool.SetContext(c.Context)
	}
	return nil
}

// BuildJWT builds a short-lived JSON Web Token for this component
func (c *Component) BuildJWT() (string, error) {
	if c.privateKey == nil {
		return "", nil
	}
	if c.Identity == nil {
		return "", nil
	}
	privPEM, err := security.PrivatePEM(c.privateKey)
	if err != nil {
		return "", err
	}
	return security.BuildJWT(c.Identity.ID, 20*time.Second, privPEM)
}

// GetContext returns a context for outgoing RPC request. If token is "", this function will generate a short lived token from the component
func (c *Component) GetContext(token string) context.Context {
	if c.Context == nil {
		c.initBgCtx()
	}
	ctx := c.Context
	if token == "" && c.Identity != nil {
		token, _ = c.BuildJWT()
	}
	ctx = ttnctx.OutgoingContextWithToken(ctx, token)
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
		return "", fmt.Errorf("Auth server \"%s\" not registered", issuerID)
	}

	token, err := getTokenFromCache(oauthCache, appID, key)
	if err != nil {
		return "", err
	}

	if token != nil {
		return token.AccessToken, nil
	}

	srv, _ := parseAuthServer(issuer)
	acc := account.New(srv.url)

	if srv.username != "" {
		acc = acc.WithAuth(auth.BasicAuth(srv.username, srv.password))
	} else {
		acc = acc.WithAuth(auth.AccessToken(c.AccessToken))
	}

	token, err = acc.ExchangeAppKeyForToken(appID, key)
	if err != nil {
		return "", err
	}

	saveTokenToCache(oauthCache, appID, key, token)

	return token.AccessToken, nil
}

// ValidateNetworkContext validates the context of a network request (router-broker, broker-handler, etc)
func (c *Component) ValidateNetworkContext(ctx context.Context) (component *pb_discovery.Announcement, err error) {
	defer func() {
		if err != nil {
			time.Sleep(time.Second)
		}
	}()

	id, err := ttnctx.IDFromIncomingContext(ctx)
	if err != nil {
		return nil, err
	}

	serviceName, _, _, _ := ttnctx.ServiceInfoFromIncomingContext(ctx)
	if serviceName == "" {
		return nil, errors.NewErrInvalidArgument("Metadata", "service-name missing")
	}

	announcement, err := c.Discover(serviceName, id)
	if err != nil {
		return nil, err
	}

	if announcement.PublicKey == "" {
		return announcement, nil
	}

	token, err := ttnctx.TokenFromIncomingContext(ctx)
	if err != nil {
		return nil, err
	}

	var claims *jwt.StandardClaims
	claims, err = security.ValidateJWT(token, []byte(announcement.PublicKey))
	if err != nil {
		return
	}
	if claims.Issuer != id {
		err = errors.NewErrPermissionDenied(fmt.Sprintf("Token was issued by %s, not by %s", claims.Issuer, id))
		return
	}
	if claims.Subject != "" && claims.Subject != claims.Issuer && claims.Subject != c.Identity.ID {
		err = errors.NewErrPermissionDenied(fmt.Sprintf("Token was issued to connect with %s, not with %s", claims.Subject, c.Identity.ID))
		return
	}

	return announcement, nil
}

// ValidateTTNAuthContext gets a token from the context and validates it
func (c *Component) ValidateTTNAuthContext(ctx context.Context) (*claims.Claims, error) {
	token, err := ttnctx.TokenFromIncomingContext(ctx)
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
