// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pool

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"sync"
	"time"

	"github.com/TheThingsNetwork/go-utils/grpc/restartstream"
	"github.com/TheThingsNetwork/go-utils/roots"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// RootCAs to use in API connections
var RootCAs *x509.CertPool

func init() {
	var err error
	RootCAs, err = x509.SystemCertPool()
	if err != nil {
		RootCAs = roots.MozillaRootCAs
	}
}

// TLSConfig that will be used when dialing securely without supplying TransportCredentials
func TLSConfig(serverName string) *tls.Config {
	return &tls.Config{ServerName: serverName, RootCAs: RootCAs}
}

// Pool with connections
type Pool struct {
	dialOptions []grpc.DialOption
	bgCtx       context.Context

	mu    sync.Mutex
	conns map[string]*conn
}

type conn struct {
	sync.WaitGroup
	target string
	opts   []grpc.DialOption
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	err    error
}

func (c *conn) dial(ctx context.Context, opts ...grpc.DialOption) {
	c.Add(1)
	go func() {
		ctx, c.cancel = context.WithCancel(ctx)
		c.conn, c.err = grpc.DialContext(ctx, c.target, opts...)
		c.Done()
	}()
}

// KeepAliveDialer is a dialer that adds a 10 second TCP KeepAlive
func KeepAliveDialer(addr string, timeout time.Duration) (net.Conn, error) {
	return (&net.Dialer{Timeout: timeout, KeepAlive: 10 * time.Second}).Dial("tcp", addr)
}

// DefaultDialOptions for connecting with servers
var DefaultDialOptions = []grpc.DialOption{
	grpc.WithStreamInterceptor(restartstream.Interceptor(restartstream.DefaultSettings)),
	grpc.WithDialer(KeepAliveDialer),
	grpc.WithBlock(),
}

// Global pool with connections
var Global = NewPool(context.Background(), DefaultDialOptions...)

// NewPool returns a new connection pool that uses the given DialOptions
func NewPool(ctx context.Context, dialOptions ...grpc.DialOption) *Pool {
	return &Pool{
		bgCtx:       ctx,
		dialOptions: dialOptions,
		conns:       make(map[string]*conn),
	}
}

// SetContext sets a new background context for the pool. Only new connections will use this new context
func (p *Pool) SetContext(ctx context.Context) {
	p.bgCtx = ctx
}

// AddDialOption adds DialOption for the pool. Only new connections will use these new DialOptions
func (p *Pool) AddDialOption(opts ...grpc.DialOption) {
	p.dialOptions = append(p.dialOptions, opts...)
}

// Close connections. If no target names supplied. just closes all.
func (p *Pool) Close(target ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(target) == 0 {
		// Select all
		for name := range p.conns {
			target = append(target, name)
		}
	}
	for _, target := range target {
		if c, ok := p.conns[target]; ok {
			c.cancel()
			if c.conn != nil {
				c.conn.Close()
			}
			delete(p.conns, target)
		}
	}
}

func (p *Pool) dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	p.mu.Lock()
	if _, ok := p.conns[target]; !ok {
		c := &conn{
			target: target,
			opts:   opts,
		}
		c.dial(p.bgCtx, append(p.dialOptions, c.opts...)...)
		p.conns[target] = c
	}
	c := p.conns[target]
	p.mu.Unlock()
	c.Wait()
	return c.conn, c.err
}

// DialInsecure gets a connection from the pool or creates a new one
// This function is blocking if grpc.WithBlock() is used
func (p *Pool) DialInsecure(target string) (*grpc.ClientConn, error) {
	return p.dial(target, grpc.WithInsecure())
}

// DialSecure gets a connection from the pool or creates a new one
// This function is blocking if grpc.WithBlock() is used
func (p *Pool) DialSecure(target string, creds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	if creds == nil {
		netHost, _, _ := net.SplitHostPort(target)
		creds = credentials.NewTLS(TLSConfig(netHost))
	}
	return p.dial(target, grpc.WithTransportCredentials(creds))
}
