// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pool

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheThingsNetwork/go-utils/grpc/restartstream"
	"google.golang.org/grpc"
)

// Pool with connections
type Pool struct {
	dialOptions []grpc.DialOption
	bgCtx       context.Context

	mu    sync.Mutex
	conns map[string]*conn
}

type conn struct {
	sync.WaitGroup
	cancel context.CancelFunc

	users int32

	conn *grpc.ClientConn
	err  error
}

// KeepAliveDialer is a dialer that adds a 10 second TCP KeepAlive
func KeepAliveDialer(addr string, timeout time.Duration) (net.Conn, error) {
	return (&net.Dialer{Timeout: timeout, KeepAlive: 10 * time.Second}).Dial("tcp", addr)
}

// DefaultDialOptions for connecting with servers
var DefaultDialOptions = []grpc.DialOption{
	grpc.WithStreamInterceptor(restartstream.Interceptor(restartstream.DefaultSettings)),
	grpc.WithDialer(KeepAliveDialer),
}

// Global pool with connections
var Global = NewPool(DefaultDialOptions)

// NewPool returns a new connection pool that uses the given DialOptions
func NewPool(dialOptions []grpc.DialOption) *Pool {
	return &Pool{
		dialOptions: dialOptions,
		conns:       make(map[string]*conn),
	}
}

// Close connections. If target names supplied, considers the other users. Otherwise just closes all.
func (p *Pool) Close(target ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(target) == 0 {
		// This force-closes all connections
		for _, c := range p.conns {
			c.cancel()
			if c.conn != nil {
				c.conn.Close()
			}
		}
		p.conns = make(map[string]*conn)
	}
	for _, target := range target {
		if c, ok := p.conns[target]; ok {
			new := atomic.AddInt32(&c.users, -1)
			if new < 1 {
				c.cancel()
				if c.conn != nil {
					c.conn.Close()
				}
				delete(p.conns, target)
			}
		}
	}
}

// DialContext gets a connection from the pool or creates a new one
// This function is blocking if grpc.WithBlock() is used
func (p *Pool) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	p.mu.Lock()
	if _, ok := p.conns[target]; !ok {
		c := new(conn)
		c.Add(1)
		p.conns[target] = c
		go func() {
			ctx, c.cancel = context.WithCancel(ctx)
			opts = append(p.dialOptions, opts...)
			c.conn, c.err = grpc.DialContext(ctx, target, opts...)
			c.Done()
		}()
	}
	c := p.conns[target]
	p.mu.Unlock()

	atomic.AddInt32(&c.users, 1)

	c.Wait()
	return c.conn, c.err
}

// Dial gets a connection from the pool or creates a new one
// This function is blocking if grpc.WithBlock() is used
func (p *Pool) Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return p.DialContext(context.Background(), target, opts...)
}
