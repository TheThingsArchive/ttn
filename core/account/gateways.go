// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"errors"
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// ListApplications list all applications
func (a *Account) ListGateways() (gateways []Gateway, err error) {
	err = a.get("/gateways", &gateways)
	return gateways, err
}

// FindGateway returns the information about a specific gateay
func (a *Account) FindGateway(gatewayID string) (gateway Gateway, err error) {
	err = a.get(fmt.Sprintf("/gateways/%s", gatewayID), &gateway)
	return gateway, err
}

// NewGateway is used as a paramater to CreateGateway to allow for optional
// arguments
type NewGateway struct {
	// ID is the ID of the new gateway (required)
	ID string `json:"id"`

	// Country is the country code of the new gateway (required)
	FrequencyPlan string `json:"frequency_plan"`

	// EUI is the EUI of the new gateway
	EUI string `json:"eui,omitemtpy"`

	// Location is the location of the new gateway
	Location *Location `json:"location,omitempty"`
}

// CreateGateway registers a new gateway on the account server
func (a *Account) CreateGateway(opts *NewGateway) (gateway Gateway, err error) {
	if opts.ID == "" {
		return gateway, errors.New("Cannot create gateway: no ID given")
	}

	if opts.FrequencyPlan == "" {
		return gateway, errors.New("Cannot create gateway: no FrequencyPlan given")
	}

	err = a.post("/gateways", &opts, &gateway)
	return gateway, err
}

// DeleteGateway removes a gateway from the account server
func (a *Account) DeleteGateway(gatewayID string) error {
	return a.del(fmt.Sprintf("/gateways/%s", gatewayID))
}

// Grant grants rights to a collaborator of the gateway
func (a *Account) GrantGatewayRights(gatewayID string, username string, rights []types.Right) error {
	req := grantReq{
		Rights: rights,
	}
	return a.put(fmt.Sprintf("/gateways/%s/collaborators/%s", gatewayID, username), req, nil)
}

// Retract removes rights from a collaborator of the gateway
func (a *Account) RetractGatewayRights(gatewayID string, username string) error {
	return a.del(fmt.Sprintf("/gateways/%s/collaborators/%s", gatewayID, username))
}

type editGatewayReq struct {
	Owner        string        `json:"owner,omitempty"`
	PublicRights []types.Right `json:"public_rights,omitempty"`
}

// TransferOwnership transfers the owenership of the gateway to another user
func (a *Account) TransferOwnership(gatewayID, username string) error {
	req := &editGatewayReq{
		Owner: username,
	}

	return a.patch(fmt.Sprintf("/gateways/%s", gatewayID), req, nil)
}

// SetPublicRights changes the publicily visible rights of the gateway
func (a *Account) SetPublicRights(gatewayID string, rights []types.Right) error {
	req := &editGatewayReq{
		PublicRights: rights,
	}

	return a.patch(fmt.Sprintf("/gateways/%s", gatewayID), req, nil)
}
