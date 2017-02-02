// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/roots"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// NewTLSClient creates a new DefaultClient with TLS enabled
func NewTLSClient(ctx log.Interface, id, username, password string, tlsConfig *tls.Config, brokers ...string) Client {
	ttnClient := NewClient(ctx, id, username, password, brokers...).(*DefaultClient)
	if tlsConfig == nil {
		ttnClient.opts.SetTLSConfig(&tls.Config{
			RootCAs: RootCAs,
		})
	} else {
		ttnClient.opts.SetTLSConfig(tlsConfig)
	}
	ttnClient.mqtt = MQTT.NewClient(ttnClient.opts)
	return ttnClient
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
