// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/account/util"
	"golang.org/x/oauth2"
)

// Account is a proxy to an account on the account server
type Account struct {
	// server is the server where the account lives
	server string

	// login is the login strategy used by the account to log in
	tokenSource oauth2.TokenSource
}

// New creates a new Account for the given server and accessToken
func New(server string, source oauth2.TokenSource) *Account {
	return &Account{
		server:      server,
		tokenSource: source,
	}
}

// WithAccessToken creates a new Account that just has an accessToken,
// which it cannot refresh itself. This is useful if you need to manage
// the token refreshes outside of the Account.
func WithAccessToken(server, accessToken string) *Account {
	source := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	})
	return &Account{
		server:      server,
		tokenSource: source,
	}
}

// Token returns the last valid accessToken/refreshToken pair
// and should be used to store the latest token after doing things
// with an Account. If you fail to do this, the token might have refreshed
// and restoring your last copy will return an invalid token
func (a *Account) Token() (*oauth2.Token, error) {
	if a.tokenSource == nil {
		return nil, fmt.Errorf("Could not get credentials for account")
	}
	return a.tokenSource.Token()
}

// AccessToken returns a valid access token for the account,
// refreshing it if necessary
func (a *Account) AccessToken() (accessToken string, err error) {
	token, err := a.Token()
	if err != nil {
		return accessToken, err
	}
	return token.AccessToken, nil
}

func (a *Account) get(URI string, res interface{}) error {
	accessToken, err := a.AccessToken()
	if err != nil {
		return err
	}

	return util.GET(a.server, accessToken, URI, res)
}

func (a *Account) put(URI string, body, res interface{}) error {
	accessToken, err := a.AccessToken()
	if err != nil {
		return err
	}

	return util.PUT(a.server, accessToken, URI, body, res)
}

func (a *Account) post(URI string, body, res interface{}) error {
	accessToken, err := a.AccessToken()
	if err != nil {
		return err
	}

	return util.POST(a.server, accessToken, URI, body, res)
}

func (a *Account) patch(URI string, body, res interface{}) error {
	accessToken, err := a.AccessToken()
	if err != nil {
		return err
	}

	return util.PATCH(a.server, accessToken, URI, body, res)
}

func (a *Account) del(URI string) error {
	accessToken, err := a.AccessToken()
	if err != nil {
		return err
	}

	return util.DELETE(a.server, accessToken, URI)
}
