// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// Application represents an application on The Things Network
type Application struct {
	ID            string            `json:"id" validate:"nonzero"`
	Name          string            `json:"name" validate:"nonzero"`
	EUIs          []types.AppEUI    `json:"euis,omitempty"`
	AccessKeys    []types.AccessKey `json:"access_keys,omitempty"`
	Created       time.Time         `json:"created,omitempty"`
	Collaborators []Collaborator    `json:"collaborators,omitempty"`
}

// Collaborator is a user that has rights to a certain application
type Collaborator struct {
	Username string        `json:"username" validate:"nonzero"`
	Rights   []types.Right `json:"rights" validate:"nonzero"`
}
