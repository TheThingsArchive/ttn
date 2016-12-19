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

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/roots"
	"github.com/TheThingsNetwork/ttn/utils/backoff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// KeepAlive indicates the keep-alive time for the Dialer
var KeepAlive = 10 * time.Second

// MaxRetries indicates how often clients should retry dialing a component
var MaxRetries = 100

// DialOptions to use in TTN gRPC
var DialOptions = []grpc.DialOption{
	WithTTNDialer(),
	grpc.WithBlock(),
	grpc.FailOnNonTempDialError(true),
}

func dial(address string, tlsConfig *tls.Config, fallback bool) (conn *grpc.ClientConn, err error) {
	ctx := log.Get().WithField("Address", address)
	opts := DialOptions
	if tlsConfig != nil {
		tlsConfig.ServerName = strings.SplitN(address, ":", 2)[0] // trim the port
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err = grpc.Dial(
		address,
		opts...,
	)
	if err == nil {
		return
	}

	switch err := err.(type) {
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
			ctx.WithError(err).Warn("Could not connect to gRPC server with TLS, reconnecting without it...")
			return dial(address, nil, fallback)
		}
		return nil, err
	}

	log.Get().WithField("ErrType", fmt.Sprintf("%T", err)).WithError(err).Error("Unhandled dial error [please create issue on Github]")
	return nil, err
}

// RootCAs to use in API connections
var RootCAs *x509.CertPool

func init() {
	var err error
	RootCAs, err = x509.SystemCertPool()
	if err != nil {
		RootCAs = roots.MozillaRootCAs
	}
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

// WithTTNDialer creates a dialer for TTN
func WithTTNDialer() grpc.DialOption {
	return grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
		ctx := log.Get().WithField("Address", addr)
		d := net.Dialer{Timeout: timeout, KeepAlive: KeepAlive}
		var retries int
		for {
			conn, err := d.Dial("tcp", addr)
			if err == nil {
				ctx.Debug("Connected to gRPC server")
				return conn, nil
			}
			if err, ok := err.(*net.OpError); ok && err.Op == "dial" && retries <= MaxRetries {
				ctx.WithError(err).Debug("Could not connect to gRPC server, reconnecting...")
				time.Sleep(backoff.Backoff(retries))
				retries++
				continue
			}
			return nil, err
		}
	})
}
