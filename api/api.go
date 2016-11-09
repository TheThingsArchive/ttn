// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Validator interface is used to validate protos, nil if validation is successful
type Validator interface {
	Validate() error
}

// Validate the given object if it implements the Validator interface
func Validate(in interface{}) error {
	if v, ok := in.(Validator); ok {
		return v.Validate()
	}
	return nil
}

// Backoff indicates how long a client should wait between failed requests
var Backoff = 1 * time.Second

// KeepAlive indicates the keep-alive time for the Dialer
var KeepAlive = 10 * time.Second

// DialOptions to use in TTN gRPC
var DialOptions = []grpc.DialOption{
	WithKeepAliveDialer(),
}

// DialWithCert dials the address using the given TLS root cert
func DialWithCert(address string, cert string) (*grpc.ClientConn, error) {
	var tlsConfig *tls.Config

	if cert != "" {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(cert))
		if !ok {
			panic("failed to parse root certificate")
		}
		tlsConfig = &tls.Config{RootCAs: roots}
	}

	opts := DialOptions
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	return grpc.Dial(
		address,
		opts...,
	)
}

// WithKeepAliveDialer creates a dialer with the configured KeepAlive time
func WithKeepAliveDialer() grpc.DialOption {
	return grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
		d := net.Dialer{Timeout: timeout, KeepAlive: KeepAlive}
		return d.Dial("tcp", addr)
	})
}
