// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"path/filepath"
	"time"
)

var (
	validFor = 365 * 24 * time.Hour
)

// GenerateKeypair generates a new keypair in the given location
func GenerateKeypair(location string) error {
	// Generate private key
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	privPEM, err := PrivatePEM(key)
	if err != nil {
		return err
	}
	pubPEM, err := PublicPEM(key)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Clean(location+"/server.pub"), pubPEM, 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Clean(location+"/server.key"), privPEM, 0600)
	if err != nil {
		return err
	}

	return nil
}

// GenerateCert generates a certificate for the given hostnames in the given location
func GenerateCert(location string, commonName string, hostnames ...string) error {
	privKey, err := LoadKeypair(location)
	if err != nil {
		return err
	}

	// Build Certificate
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"The Things Network"},
		},
		IsCA:                  true,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	for _, h := range hostnames {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// Generate certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, privKey.Public(), privKey)
	if err != nil {
		return err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	err = ioutil.WriteFile(filepath.Clean(location+"/server.cert"), certPEM, 0644)
	if err != nil {
		return err
	}

	return nil
}
