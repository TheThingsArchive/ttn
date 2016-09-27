// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"errors"
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// ListGateways list all gateways
func (a *Account) ListGateways() (gateways []Gateway, err error) {
	err = a.get("/gateways", &gateways)
	return gateways, err
}

// FindGateway returns the information about a specific gateay
func (a *Account) FindGateway(gatewayID string) (gateway Gateway, err error) {
	err = a.get(fmt.Sprintf("/gateways/%s", gatewayID), &gateway)
	return gateway, err
}

type registerGatewayReq struct {
	// ID is the ID of the new gateway (required)
	ID string `json:"id"`

	// Country is the country code of the new gateway (required)
	FrequencyPlan string `json:"frequency_plan"`

	// Location is the location of the new gateway
	Location *Location `json:"location,omitempty"`
}

// RegisterGateway registers a new gateway on the account server
func (a *Account) RegisterGateway(gatewayID string, frequencyPlan string, location *Location) (gateway Gateway, err error) {
	if gatewayID == "" {
		return gateway, errors.New("Cannot create gateway: no ID given")
	}

	if frequencyPlan == "" {
		return gateway, errors.New("Cannot create gateway: no FrequencyPlan given")
	}

	req := registerGatewayReq{
		ID:            gatewayID,
		FrequencyPlan: frequencyPlan,
		Location:      location,
	}

	err = a.post("/gateways", req, &gateway)
	return gateway, err
}

// DeleteGateway removes a gateway from the account server
func (a *Account) DeleteGateway(gatewayID string) error {
	return a.del(fmt.Sprintf("/gateways/%s", gatewayID))
}

// GrantGatewayRights grants rights to a collaborator of the gateway
func (a *Account) GrantGatewayRights(gatewayID string, username string, rights []types.Right) error {
	req := grantReq{
		Rights: rights,
	}
	return a.put(fmt.Sprintf("/gateways/%s/collaborators/%s", gatewayID, username), req, nil)
}

// RetractGatewayRights removes rights from a collaborator of the gateway
func (a *Account) RetractGatewayRights(gatewayID string, username string) error {
	return a.del(fmt.Sprintf("/gateways/%s/collaborators/%s", gatewayID, username))
}

// GatewayEdits contains editable fields of gateways
type GatewayEdits struct {
	Owner         string        `json:"owner,omitempty"`
	PublicRights  []types.Right `json:"public_rights,omitempty"`
	FrequencyPlan string        `json:"frequency_plan,omitempty"`
	Location      *Location     `json:"location,omitempty"`
}

// EditGateway edits the fields of a gateway
func (a *Account) EditGateway(gatewayID string, edits GatewayEdits) (gateway Gateway, err error) {
	err = a.patch(fmt.Sprintf("/gateways/%s", gatewayID), edits, &gateway)
	return gateway, err
}

// TransferOwnership transfers the owenership of the gateway to another user
func (a *Account) TransferOwnership(gatewayID, username string) (Gateway, error) {
	return a.EditGateway(gatewayID, GatewayEdits{
		Owner: username,
	})
}

// SetPublicRights changes the publicily visible rights of the gateway
func (a *Account) SetPublicRights(gatewayID string, rights []types.Right) (Gateway, error) {
	return a.EditGateway(gatewayID, GatewayEdits{
		PublicRights: rights,
	})
}

// ChangeFrequencyPlan changes the requency plan of a gateway
func (a *Account) ChangeFrequencyPlan(gatewayID, plan string) (Gateway, error) {
	return a.EditGateway(gatewayID, GatewayEdits{
		FrequencyPlan: plan,
	})
}

// ChangeLocation changes the location of the gateway
func (a *Account) ChangeLocation(gatewayID string, latitude, longitude float64) (Gateway, error) {
	return a.EditGateway(gatewayID, GatewayEdits{
		Location: &Location{
			Longitude: longitude,
			Latitude:  latitude,
		},
	})
}
