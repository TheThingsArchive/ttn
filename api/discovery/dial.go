// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"strings"

	"github.com/TheThingsNetwork/ttn/api/pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TLSConfig for securely connecting with this component
func (a *Announcement) TLSConfig() (*tls.Config, error) {
	if a.NetAddress == "" {
		return nil, errors.New("No address known for this component")
	}
	netAddress := strings.Split(a.NetAddress, ",")[0]
	netHost, _, _ := net.SplitHostPort(netAddress)
	if a.Certificate == "" {
		return nil, nil
	}
	rootCAs := x509.NewCertPool()
	ok := rootCAs.AppendCertsFromPEM([]byte(a.Certificate))
	if !ok {
		return nil, errors.New("Could not read component certificate")
	}
	return &tls.Config{ServerName: netHost, RootCAs: rootCAs}, nil
}

// WithSecure returns a gRPC DialOption with TLS
func (a *Announcement) WithSecure() grpc.DialOption {
	tlsConfig, _ := a.TLSConfig()
	return grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
}

// Dial the component represented by this Announcement. We use the global connection pool here
func (a *Announcement) Dial(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if a.NetAddress == "" {
		return nil, errors.New("No address known for this component")
	}
	netAddress := strings.Split(a.NetAddress, ",")[0]
	if a.Certificate == "" {
		return pool.Global.Dial(netAddress, append(opts, grpc.WithInsecure())...)
	}
	return pool.Global.Dial(netAddress, append(opts, a.WithSecure())...)
}
