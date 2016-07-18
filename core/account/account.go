// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/apex/log"
)

// Account is a proxy to an account on the account server
type Account struct {
	accessToken string
}

// New creates a new Account with the default Client
func New(accessToken string) *Account {
	return &Account{
		accessToken: accessToken,
	}
}

// ListApplications list all applications
func (a *Account) ListApplications(ctx log.Interface) (apps []Application, err error) {
	panic("ListApplications not implemented")
}

// GetApplication gets a specific application from the account server
func (a *Account) FindApplication(ctx log.Interface, appID string) (Application, error) {
	panic("FindApplication not implemented")
}

// CreateApplication creates a new application on the account server
func (a *Account) CreateApplication(ctx log.Interface, appID string, name string, EUIs []types.AppEUI) (Application, error) {
	panic("CreateApplication not implemented")
}

// DeleteApplication deletes an application
func (a *Account) DeleteAppliction(ctx log.Interface, appID string) error {
	panic("DeleteApplication not implemented")
}

// Grant adds a collaborator to the application
func (a *Account) Grant(ctx log.Interface, appID string, username string, rights []Right) error {
	panic("Grant not implemented")
}

// Retract removes rights from a collaborator of the application
func (a *Account) Retract(ctx log.Interface, appID string, username string, rights []Right) error {
	panic("Retract not implemented")
}

// AddAccessKey
func (a *Account) AddAccessKey(ctx log.Interface, appID string, key AccessKey) error {
	panic("AddAccessKey not implemented")
}

// RemoveAccessKey
func (a *Account) RemoveAccessKey(ctx log.Interface, appID string, key AccessKey) error {
	panic("RemoveAccessKey not implemented")
}

// ChangeName
func (a *Account) ChangeName(ctx log.Interface, appID string, name string) error {
	panic("ChangeName not implemented")
}

// AddEUI
func (a *Account) AddEUI(ctx log.Interface, appID string, eui types.AppEUI) error {
	panic("AddEUI not implemented")
}

// RemoveEUI
func (a *Account) RemoveEUI(ctx log.Interface, appID string, eui types.AppEUI) error {
	panic("RemoveEUI not implemented")
}
