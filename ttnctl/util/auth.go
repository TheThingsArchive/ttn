// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"time"

	homedir "github.com/mitchellh/go-homedir"
)

const authsFilePerm = 0600

// AuthsFileName is where the authentication tokens are stored. Defaults to
// $HOME/.ttnctl/auths.json
var AuthsFileName string

// Auth represents an authentication token
type Auth struct {
	Token   string    `json:"token"`
	Email   string    `json:"email"`
	Expires time.Time `json:"expires"`
}

type auths struct {
	Auths map[string]*Auth `json:"auths"`
}

func init() {
	dir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	expanded, err := homedir.Expand(dir)
	if err != nil {
		panic(err)
	}
	AuthsFileName = path.Join(expanded, ".ttnctl/auths.json")
}

// LoadAuth loads the authentication token for the specified server
func LoadAuth(server string) (*Auth, error) {
	a, err := loadAuths()
	if err != nil {
		return nil, err
	}
	t, ok := a.Auths[server]
	if !ok || time.Now().After(t.Expires) {
		return nil, nil
	}
	return t, nil
}

// SaveAuth saves the authentication token for the specified server and e-mail
func SaveAuth(server, email, token string, expires time.Time) error {
	a, err := loadAuths()
	// Ignore error - just create new structure
	if err != nil || a == nil {
		a = &auths{}
	}

	// Initialize the map if not exists and add the token
	if a.Auths == nil {
		a.Auths = make(map[string]*Auth)
	}
	a.Auths[server] = &Auth{token, email, expires}

	// Marshal and write to disk
	buff, err := json.Marshal(&a)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path.Dir(AuthsFileName), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(AuthsFileName, buff, authsFilePerm); err != nil {
		return err
	}
	return nil
}

// loadAuths loads the authentication tokens. This function always returns an
// empty structure if the file does not exist.
func loadAuths() (*auths, error) {
	if _, err := os.Stat(AuthsFileName); os.IsNotExist(err) {
		return &auths{}, nil
	}
	buff, err := ioutil.ReadFile(AuthsFileName)
	if err != nil {
		return nil, err
	}
	var a auths
	if err := json.Unmarshal(buff, &a); err != nil {
		return nil, err
	}
	return &a, nil
}
