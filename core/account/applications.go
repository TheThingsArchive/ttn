// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TheThingsNetwork/ttn/core/account/util"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// ListApplications list all applications
func (a *Account) ListApplications() (apps []Application, err error) {
	resp, err := util.GET(a.server, a.accessToken, "/applications")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&apps); err != nil {
		return nil, err
	}

	return apps, nil
}

// GetApplication gets a specific application from the account server
func (a *Account) FindApplication(appID string) (app Application, err error) {
	resp, err := util.GET(a.server, a.accessToken, fmt.Printf("/applications/%s", appID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return app, fmt.Errorf("Application with id '%s' does not exist", appID)
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&apps); err != nil {
		return app, err
	}

	return app, nil
}

type createApplicationReq struct {
	Name  string         `json:"name"`
	AppID string         `json:"id"`
	EUIS  []types.AppEUI `json:"euis"`
}

// CreateApplication creates a new application on the account server
func (a *Account) CreateApplication(appID string, name string, EUIs []types.AppEUI) (app Application, err error) {
	body := createApplicationReq{
		Name:  name,
		AppID: appID,
		EUIs:  EUIs,
	}

	resp, err := util.POST(a.server, a.accessToken, "/applications", body)
	if resp.StatusCode != http.StatusCreated {
		return app, fmt.Errorf("Could not create application: %s", resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&apps); err != nil {
		return app, err
	}

	return app, nil
}

// DeleteApplication deletes an application
func (a *Account) DeleteAppliction(appID string) error {
	panic("DeleteApplication not implemented")
}

// Grant adds a collaborator to the application
func (a *Account) Grant(appID string, username string, rights []Right) error {
	panic("Grant not implemented")
}

// Retract removes rights from a collaborator of the application
func (a *Account) Retract(appID string, username string, rights []Right) error {
	panic("Retract not implemented")
}

// AddAccessKey
func (a *Account) AddAccessKey(appID string, key AccessKey) error {
	panic("AddAccessKey not implemented")
}

// RemoveAccessKey
func (a *Account) RemoveAccessKey(appID string, key AccessKey) error {
	panic("RemoveAccessKey not implemented")
}

// ChangeName
func (a *Account) ChangeName(appID string, name string) error {
	panic("ChangeName not implemented")
}

// AddEUI
func (a *Account) AddEUI(appID string, eui types.AppEUI) error {
	panic("AddEUI not implemented")
}

// RemoveEUI
func (a *Account) RemoveEUI(appID string, eui types.AppEUI) error {
	panic("RemoveEUI not implemented")
}
