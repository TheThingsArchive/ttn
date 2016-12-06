// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/backoff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// KeepAlive indicates the keep-alive time for the Dialer
var KeepAlive = 10 * time.Second

// MaxRetries indicates how often clients should retry dialing a component
var MaxRetries = 100

// Timeout for connections
var Timeout = 2 * time.Second

// DialOptions to use in TTN gRPC
var DialOptions = []grpc.DialOption{
	WithKeepAliveDialer(),
	grpc.WithBlock(),
	grpc.FailOnNonTempDialError(true),
	grpc.WithTimeout(Timeout),
}

func dial(address string, tlsConfig *tls.Config, fallback bool) (conn *grpc.ClientConn, err error) {
	ctx := GetLogger().WithField("Address", address)
	retries := 0
	retriesLeft := MaxRetries
	opts := DialOptions
	if tlsConfig != nil {
		tlsConfig.ServerName = strings.SplitN(address, ":", 2)[0] // trim the port
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	for retriesLeft > 0 {
		conn, err = grpc.Dial(
			address,
			opts...,
		)
		if err == nil {
			ctx.Debug("Connected")
			return
		}

		switch err := err.(type) {
		case *net.OpError:
			// Dial problem
			if err.Op == "dial" {
				ctx.WithError(err).Debug("Could not connect, reconnecting...")
			}
		case x509.CertificateInvalidError,
			x509.ConstraintViolationError,
			x509.HostnameError,
			x509.InsecureAlgorithmError,
			x509.SystemRootsError,
			x509.UnhandledCriticalExtension,
			x509.UnknownAuthorityError:
			// Non-temporary error while connecting to a TLS-enabled server
			return nil, err
		case tls.RecordHeaderError:
			if fallback {
				ctx.WithError(err).Warn("Could not connect with TLS, reconnecting without it...")
				return dial(address, nil, fallback)
			}
			return nil, err
		default:
			GetLogger().WithField("ErrType", fmt.Sprintf("%T", err)).WithError(err).Error("Unhandled dial error [please create issue on Github]")
			return nil, err
		}

		// Backoff
		time.Sleep(backoff.Backoff(retries))
		retries++
		retriesLeft--
	}
	return
}

// RootCAs to use in API connections
var RootCAs *x509.CertPool

func init() {
	RootCAs, _ = x509.SystemCertPool()
}

// Dial an address
func Dial(address string) (*grpc.ClientConn, error) {
	tlsConfig := &tls.Config{RootCAs: RootCAs}
	return dial(address, tlsConfig, true)
}

// DialWithCert dials the address using the given TLS cert
func DialWithCert(address string, cert string) (*grpc.ClientConn, error) {
	rootCAs := x509.NewCertPool()
	ok := rootCAs.AppendCertsFromPEM([]byte(cert))
	if !ok {
		panic("failed to parse root certificate")
	}
	tlsConfig := &tls.Config{RootCAs: rootCAs}
	return dial(address, tlsConfig, false)
}

// WithKeepAliveDialer creates a dialer with the configured KeepAlive time
func WithKeepAliveDialer() grpc.DialOption {
	return grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
		d := net.Dialer{Timeout: timeout, KeepAlive: KeepAlive}
		return d.Dial("tcp", addr)
	})
}
