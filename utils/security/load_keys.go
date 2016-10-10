package security

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"path/filepath"
)

// LoadKeypair loads the keypair in the given location
func LoadKeypair(location string) (*ecdsa.PrivateKey, error) {
	priv, err := ioutil.ReadFile(filepath.Clean(location + "/server.key"))
	if err != nil {
		return nil, err
	}
	privBlock, _ := pem.Decode(priv)
	if privBlock == nil {
		return nil, errors.New("No private key data found")
	}
	privKey, err := x509.ParseECPrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

// LoadCert loads the certificate in the given location
func LoadCert(location string) (cert []byte, err error) {
	cert, err = ioutil.ReadFile(filepath.Clean(location + "/server.cert"))
	if err != nil {
		return
	}
	return
}
