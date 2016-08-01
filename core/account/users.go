// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core/account/util"
)

type registerUserReq struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterUser registers a new user with the specified username, email and
// password on the specified account server
func RegisterUser(server, username, email, password string) error {
	user := registerUserReq{
		Username: username,
		Email:    email,
		Password: password,
	}

	err := util.POST(server, "", "/api/users", user, nil)
	if err != nil {
		return fmt.Errorf("Could not register user: %s", err)
	}
	return nil
}

// Profile gets the user profile of the user that is logged in
func (a *Account) Profile() (user Profile, err error) {
	err = a.get("/api/users/me", &user)
	if err != nil {
		return user, fmt.Errorf("Could not get user profile: %s", err)
	}

	return user, nil
}

type nameReq struct {
	First string `json:"first,omitemtpy"`
	Last  string `json:"last,omitempty"`
}

type editProfileReq struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Name     *Name  `json:"name,omitempty"`
}

// EditProfile edits the users profile. You can change only
// part of the profile (only the name, for instance) by
// omitting the other fields from the passed in Profile struct.
func (a *Account) EditProfile(profile Profile) error {
	var edits editProfileReq

	if profile.Username != "" {
		edits.Username = profile.Username
	}

	if profile.Email != "" {
		edits.Email = profile.Email
	}

	if profile.Name != nil {
		edits.Name = profile.Name
	}

	err := a.patch("/api/users/me", edits, nil)
	if err != nil {
		return fmt.Errorf("Could not update profile: %s", err)
	}
	return err
}

type editPasswordReq struct {
	OldPassword string `json:"old_password"`
	Password    string `json:"password"`
}

// EditPassword edits the users password, it requires the old password
// to be given.
func (a *Account) EditPassword(oldPassword, newPassword string) error {
	edits := editPasswordReq{
		OldPassword: oldPassword,
		Password:    newPassword,
	}

	err := a.patch("/api/users/me", edits, nil)
	if err != nil {
		return fmt.Errorf("Could change not password: %s", err)
	}
	return nil
}
