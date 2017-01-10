// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package security

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// BuildJWT builds a JSON Web Token for the given subject and ttl, and signs it with the given private key
func BuildJWT(subject string, ttl time.Duration, privateKey []byte) (token string, err error) {
	claims := jwt.StandardClaims{
		Issuer:    subject,
		Subject:   subject,
		IssuedAt:  time.Now().Add(-20 * time.Second).Unix(),
		NotBefore: time.Now().Add(-20 * time.Second).Unix(),
	}
	if ttl > 0 {
		claims.ExpiresAt = time.Now().Add(ttl).Unix()
	}
	tokenBuilder := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	var key *ecdsa.PrivateKey
	key, err = jwt.ParseECPrivateKeyFromPEM(privateKey)
	if err != nil {
		return
	}
	token, err = tokenBuilder.SignedString(key)
	if err != nil {
		return
	}
	return
}

// ValidateJWT validates a JSON Web Token with the given public key
func ValidateJWT(token string, publicKey []byte) (*jwt.StandardClaims, error) {
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected JWT signing method: %v", token.Header["alg"])
		}
		key, err := jwt.ParseECPublicKeyFromPEM(publicKey)
		if err != nil {
			return nil, err
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to verify JWT: %s", err.Error())
	}
	return claims, nil
}
