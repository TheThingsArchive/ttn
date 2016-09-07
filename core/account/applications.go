// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// ListApplications list all applications
func (a *Account) ListApplications() (apps []Application, err error) {
	err = a.get("/applications", &apps)
	return apps, err
}

// FindApplication gets a specific application from the account server
func (a *Account) FindApplication(appID string) (app Application, err error) {
	err = a.get(fmt.Sprintf("/applications/%s", appID), &app)
	return app, err
}

type createApplicationReq struct {
	Name  string         `json:"name" valid:"required"`
	AppID string         `json:"id"   valid:"required"`
	EUIs  []types.AppEUI `json:"euis"`
}

// CreateApplication creates a new application on the account server
func (a *Account) CreateApplication(appID string, name string, EUIs []types.AppEUI) (app Application, err error) {
	body := createApplicationReq{
		Name:  name,
		AppID: appID,
		EUIs:  EUIs,
	}

	err = a.post("/applications", &body, &app)
	return app, err
}

// DeleteApplication deletes an application
func (a *Account) DeleteApplication(appID string) error {
	return a.del(fmt.Sprintf("/applications/%s", appID))
}

type grantReq struct {
	Rights []types.Right `json:"rights"`
}

// Grant adds a collaborator to the application
func (a *Account) Grant(appID string, username string, rights []types.Right) error {
	req := grantReq{
		Rights: rights,
	}
	return a.put(fmt.Sprintf("/applications/%s/collaborators/%s", appID, username), req, nil)
}

// Retract removes rights from a collaborator of the application
func (a *Account) Retract(appID string, username string) error {
	return a.del(fmt.Sprintf("/applications/%s/collaborators/%s", appID, username))
}

type addAccessKeyReq struct {
	Name   string        `json:"name"   valid:"required"`
	Rights []types.Right `json:"rights" valid:"required"`
}

// AddAccessKey adds an access key to the application with the specified name
// and rights
func (a *Account) AddAccessKey(appID string, name string, rights []types.Right) (key types.AccessKey, err error) {
	body := addAccessKeyReq{
		Name:   name,
		Rights: rights,
	}
	err = a.post(fmt.Sprintf("/applications/%s/access-keys", appID), body, &key)
	return key, err
}

// RemoveAccessKey removes the specified access key from the application
func (a *Account) RemoveAccessKey(appID string, name string) error {
	return a.del(fmt.Sprintf("/applications/%s/access-keys/%s", appID, name))
}

type editAppReq struct {
	Name string `json:"name,omitempty"`
}

// ChangeName changes the application name
func (a *Account) ChangeName(appID string, name string) (app Application, err error) {
	body := editAppReq{
		Name: name,
	}
	err = a.patch(fmt.Sprintf("/applications/%s", appID), body, &app)
	return app, err
}

// AddEUI adds an EUI to the applications list of EUIs
func (a *Account) AddEUI(appID string, eui types.AppEUI) error {
	return a.put(fmt.Sprintf("/applications/%s/euis/%s", appID, eui.String()), nil, nil)
}

type genEUIRes struct {
	EUI types.AppEUI `json:"eui"`
}

// GenerateEUI creates a new EUI for the application
func (a *Account) GenerateEUI(appID string) (*types.AppEUI, error) {
	var res genEUIRes
	err := a.post(fmt.Sprintf("/applications/%s/euis", appID), nil, &res)
	if err != nil {
		return nil, err
	}
	return &res.EUI, nil
}

// RemoveEUI removes the specified EUI from the application
func (a *Account) RemoveEUI(appID string, eui types.AppEUI) error {
	return a.del(fmt.Sprintf("/applications/%s/euis/%s", appID, eui.String()))
}
