// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/account/util"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// ListApplications list all applications
func (a *Account) ListApplications() (apps []Application, err error) {
	err = util.GET(a.server, a.accessToken, "/applications", &apps)
	return apps, err
}

// GetApplication gets a specific application from the account server
func (a *Account) FindApplication(appID string) (app Application, err error) {
	err = util.GET(a.server, a.accessToken, fmt.Sprintf("/applications/%s", appID), &app)
	return app, err
}

type createApplicationReq struct {
	Name  string         `json:"name"`
	AppID string         `json:"id"`
	EUIs  []types.AppEUI `json:"euis"`
}

// CreateApplication creates a new application on the account server
func (a *Account) CreateApplication(appID string, name string, EUIs []types.AppEUI) (app Application, err error) {
	body := createApplicationReq{
		Name:  name,
		AppID: appID,
		EUIs:  EUIs,
	}

	err = util.POST(a.server, a.accessToken, "/applications", &body, &app)
	return app, err
}

// DeleteApplication deletes an application
func (a *Account) DeleteAppliction(appID string) error {
	return util.DELETE(a.server, a.accessToken, fmt.Sprintf("/applications/%s", appID))
}

// Grant adds a collaborator to the application
func (a *Account) Grant(appID string, username string, rights []types.Right) error {
	return util.PUT(a.server, a.accessToken, fmt.Sprintf("/applications/%s/collaborators/%s", appID, username), rights, nil)
}

// Retract removes rights from a collaborator of the application
func (a *Account) Retract(appID string, username string) error {
	return util.DELETE(a.server, a.accessToken, fmt.Sprintf("/applications/%s/collaborators/%s", appID, username))
}

type addAccessKeyReq struct {
	Name   string        `json:"name" valid:"required"`
	Rights []types.Right `json:"rights" valid:"required"`
}

// AddAccessKey
func (a *Account) AddAccessKey(appID string, name string, rights []types.Right) (key types.AccessKey, err error) {
	body := addAccessKeyReq{
		Name:   name,
		Rights: rights,
	}
	util.POST(a.server, a.accessToken, fmt.Sprintf("/applications/%s/access-keys", appID), body, &key)
	return key, err
}

// RemoveAccessKey
func (a *Account) RemoveAccessKey(appID string, name string) error {
	return util.DELETE(a.server, a.accessToken, fmt.Sprintf("/applications/%s/access-keys/%s", appID, name))
}

type editAppReq struct {
	Name string `json:"name,omitempty"`
}

// ChangeName
func (a *Account) ChangeName(appID string, name string) (app Application, err error) {
	body := editAppReq{
		Name: name,
	}
	err = util.PATCH(a.server, a.accessToken, fmt.Sprintf("/applications/%s", appID), body, &app)
	return app, err
}

// AddEUI
func (a *Account) AddEUI(appID string, eui types.AppEUI) error {
	return util.POST(a.server, a.accessToken, fmt.Sprintf("/applications/%s/euis/%s", appID, eui.String()), nil, nil)
}

// RemoveEUI
func (a *Account) RemoveEUI(appID string, eui types.AppEUI) error {
	return util.DELETE(a.server, a.accessToken, fmt.Sprintf("/applications/%s/euis/%s", appID, eui.String()))
}
