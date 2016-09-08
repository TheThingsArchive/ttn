// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"github.com/TheThingsNetwork/ttn/core/account/auth"
	"github.com/TheThingsNetwork/ttn/core/account/util"
)

// Account is a client to an account server
type Account struct {
	server string
	auth   auth.Strategy
}

// New creates a new account client that will use the
// accessToken to make requests to the specified account server
func New(server, accessToken string) *Account {
	return &Account{
		server: server,
		auth:   auth.AccessToken(accessToken),
	}
}

// NewWithKey creates an account client that uses an accessKey to
// authenticate
func NewWithKey(server, accessKey string) *Account {
	return &Account{
		server: server,
		auth:   auth.AccessKey(accessKey),
	}
}

func NewWithPublic(server string) *Account {
	return &Account{
		server: server,
		auth:   auth.Public,
	}
}

func (a *Account) get(URI string, res interface{}) error {
	return util.GET(a.server, a.auth, URI, res)
}

func (a *Account) put(URI string, body, res interface{}) error {
	return util.PUT(a.server, a.auth, URI, body, res)
}

func (a *Account) post(URI string, body, res interface{}) error {
	return util.POST(a.server, a.auth, URI, body, res)
}

func (a *Account) patch(URI string, body, res interface{}) error {
	return util.PATCH(a.server, a.auth, URI, body, res)
}

func (a *Account) del(URI string) error {
	return util.DELETE(a.server, a.auth, URI)
}
