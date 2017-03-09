package auth

import (
	"github.com/TheThingsNetwork/ttn/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const tokenKey = "token"

// TokenCredentials RPC Credentials
type TokenCredentials struct {
	token     string
	tokenFunc func(id string) string
}

// WithStaticToken injects a static token on each request
func WithStaticToken(token string) *TokenCredentials {
	return &TokenCredentials{token: token}
}

// WithTokenFunc returns TokenCredentials that execute the tokenFunc on each request
func WithTokenFunc(tokenFunc func(id string) string) *TokenCredentials {
	return &TokenCredentials{tokenFunc: tokenFunc}
}

// RequireTransportSecurity implements credentials.PerRPCCredentials
func (c *TokenCredentials) RequireTransportSecurity() bool { return true }

// GetRequestMetadata implements credentials.PerRPCCredentials
func (c *TokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, _ := api.TokenFromContext(ctx)
	if token != "" {
		return map[string]string{tokenKey: token}, nil
	}
	if c.tokenFunc != nil {
		id, _ := api.IDFromContext(ctx)
		return map[string]string{tokenKey: c.tokenFunc(id)}, nil
	}
	if c.token != "" {
		return map[string]string{tokenKey: c.token}, nil
	}
	return map[string]string{tokenKey: ""}, nil
}

// DialOption returns a DialOption for the TokenCredentials
func (c *TokenCredentials) DialOption() grpc.DialOption {
	return grpc.WithPerRPCCredentials(c)
}
