// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

const (
	authsFileName = ".ttnctl/auths.json"
	authsFilePerm = 0600
)

// Auth represents an authentication token
type Auth struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

type auths struct {
	Auths map[string]*Auth `json:"auths"`
}

// LoadAuth loads the authentication token for the specified server
func LoadAuth(server string) (*Auth, error) {
	a, err := loadAuths()
	if err != nil {
		return nil, err
	}
	return a.Auths[server], nil
}

// SaveAuth saves the authentication token for the specified server and e-mail
func SaveAuth(server, email, token string) error {
	a, err := loadAuths()
	// Ignore error - just create new structure
	if err != nil || a == nil {
		a = &auths{}
	}

	// Initialize the map if not exists and add the token
	if a.Auths == nil {
		a.Auths = make(map[string]*Auth)
	}
	a.Auths[server] = &Auth{token, email}

	// Marshal and write to disk
	buff, err := json.Marshal(&a)
	if err != nil {
		return err
	}
	filename, err := getAuthsFilename()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, buff, authsFilePerm); err != nil {
		return err
	}
	return nil
}

// loadAuths loads the authentication tokens. This function always returns an
// empty structure if the file does not exist.
func loadAuths() (*auths, error) {
	filename, err := getAuthsFilename()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return &auths{}, nil
	}
	buff, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var a auths
	if err := json.Unmarshal(buff, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func getAuthsFilename() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(u.HomeDir, authsFileName), nil
}
