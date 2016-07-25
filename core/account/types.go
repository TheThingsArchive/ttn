// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
)

// Application represents an application on The Things Network
type Application struct {
	ID            string            `json:"id"   valid:"required"`
	Name          string            `json:"name" valid:"required"`
	EUIs          []types.AppEUI    `json:"euis,omitempty"`
	AccessKeys    []types.AccessKey `json:"access_keys,omitempty"`
	Created       time.Time         `json:"created,omitempty"`
	Collaborators []Collaborator    `json:"collaborators,omitempty"`
}

// Collaborator is a user that has rights to a certain application
type Collaborator struct {
	Username string        `json:"username" valid:"required"`
	Rights   []types.Right `json:"rights"   valid:"required"`
}

// HasRight checks if the collaborator has a specific right
func (c *Collaborator) HasRight(right types.Right) bool {
	for _, r := range c.Rights {
		if r == right {
			return true
		}
	}
	return false
}

// Profile represents the profile of a user
type Profile struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     *Name  `json:"name"`
}

// Name represents the full name of a user
type Name struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

// Component represents a component on the newtork
type Component struct {
	Type    string    `json:"type"`
	ID      string    `json:"id"`
	Created time.Time `json:"created,omitempty"`
}

// String implements the Stringer interface for Name
func (n *Name) String() string {
	return n.First + " " + n.Last
}
