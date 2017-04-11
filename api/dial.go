// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// AllowInsecureFallback can be set to true if you need to connect with a server that does not use TLS
var AllowInsecureFallback = false

// Dial an address with default TLS config
func Dial(target string) (*grpc.ClientConn, error) {
	conn, err := pool.Global.DialSecure(target, nil)
	if err == nil {
		return conn, nil
	}
	if _, ok := err.(tls.RecordHeaderError); ok && AllowInsecureFallback {
		pool.Global.Close(target)
		log.Get().Warn("Could not connect to gRPC server with TLS, will reconnect without TLS")
		log.Get().Warnf("This is a security risk, you should enable TLS on %s", target)
		conn, err = pool.Global.DialInsecure(target)
	}
	return conn, err
}

// DialWithCert dials the target using the given TLS cert
func DialWithCert(target string, cert string) (*grpc.ClientConn, error) {
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(cert))
	if !ok {
		panic("failed to parse root certificate")
	}
	return pool.Global.DialSecure(target, credentials.NewClientTLSFromCert(certPool, ""))
}
