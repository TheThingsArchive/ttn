// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package security

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
)

// PublicPEM returns the PEM-encoded public key
func PublicPEM(key *ecdsa.PrivateKey) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return nil, err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	return pubPEM, nil
}

// PrivatePEM returns the PEM-encoded private key
func PrivatePEM(key *ecdsa.PrivateKey) ([]byte, error) {
	privBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})
	return privPEM, nil
}
