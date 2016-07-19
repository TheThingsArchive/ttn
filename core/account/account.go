// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

// Account is a proxy to an account on the account server
type Account struct {
	// server is the server where the account lives
	server string

	// accessToken is the accessToken that gives this client the
	// right to act on behalf of the account
	accessToken string
}

// New creates a new Account for the given server and accessToken
func New(server string, accessToken string) *Account {
	return &Account{
		server:      server,
		accessToken: accessToken,
	}
}
