// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import "github.com/TheThingsNetwork/ttn/core/account/util"

// Account is a client to an account server
type Account struct {
	server      string
	accessToken string
}

// New creates a new accoun client that will use the
// accessToken to make requests to the specified account server
func New(server, accessToken string) *Account {
	return &Account{
		server:      server,
		accessToken: accessToken,
	}
}

// SetToken changes the accessToken the account client uses to
// makes requests.
func (a *Account) SetToken(accessToken string) {
	a.accessToken = accessToken
}

func (a *Account) get(URI string, res interface{}) error {
	return util.GET(a.server, a.accessToken, URI, res)
}

func (a *Account) put(URI string, body, res interface{}) error {
	return util.PUT(a.server, a.accessToken, URI, body, res)
}

func (a *Account) post(URI string, body, res interface{}) error {
	return util.POST(a.server, a.accessToken, URI, body, res)
}

func (a *Account) patch(URI string, body, res interface{}) error {
	return util.PATCH(a.server, a.accessToken, URI, body, res)
}

func (a *Account) del(URI string) error {
	return util.DELETE(a.server, a.accessToken, URI)
}
