// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"net"
	"time"

	"google.golang.org/grpc"
)

// Backoff indicates how long a client should wait between failed requests
var Backoff = 1 * time.Second

// KeepAlive indicates the keep-alive time for the Dialer
var KeepAlive = 10 * time.Second

// DialOptions to use in TTN gRPC
// TODO: disable insecure connections
var DialOptions = []grpc.DialOption{
	grpc.WithInsecure(),
	WithKeepAliveDialer(),
}

// WithKeepAliveDialer creates a dialer with the configured KeepAlive time
func WithKeepAliveDialer() grpc.DialOption {
	return grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
		d := net.Dialer{Timeout: timeout, KeepAlive: KeepAlive}
		return d.Dial("tcp", addr)
	})
}
