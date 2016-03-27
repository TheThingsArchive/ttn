// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
)

const authsFilePerm = 0600

// AuthsFileName is where the authentication tokens are stored. Defaults to
// $HOME/.ttnctl/auths.json
var AuthsFileName string

// Auth represents an authentication token
type Auth struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Email        string    `json:"email"`
	Expires      time.Time `json:"expires"`
}

type auths struct {
	Auths map[string]*Auth `json:"auths"`
}

type token struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ExpiresIn        int    `json:"expires_in"`
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

// Login attemps to login using the specified credentials on the server
func Login(server, email, password string) (*Auth, error) {
	values := url.Values{
		"grant_type": {"password"},
		"username":   {email},
		"password":   {password},
	}
	return newToken(server, email, values)
}

// LoadAuth loads the authentication token for the specified server and attempts
// to refresh the token if it has been expired
func LoadAuth(server string) (*Auth, error) {
	a, err := loadAuths()
	if err != nil {
		return nil, err
	}
	auth, ok := a.Auths[server]
	if !ok {
		return nil, nil
	}
	if time.Now().After(auth.Expires) {
		return refreshToken(server, auth)
	}
	return auth, nil
}

func refreshToken(server string, auth *Auth) (*Auth, error) {
	values := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {auth.RefreshToken},
	}
	return newToken(server, auth.Email, values)
}

func newToken(server, email string, values url.Values) (*Auth, error) {
	uri := fmt.Sprintf("%s/token", server)
	req, err := http.NewRequest("POST", uri, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("ttnctl", "")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var t token
	if err := decoder.Decode(&t); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if t.Error != "" {
			return nil, errors.New(t.ErrorDescription)
		}
		return nil, errors.New(resp.Status)
	}

	expires := time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
	auth, err := saveAuth(server, email, t.AccessToken, t.RefreshToken, expires)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

// saveAuth saves the authentication token for the specified server and e-mail
func saveAuth(server, email, accessToken, refreshToken string, expires time.Time) (*Auth, error) {
	a, err := loadAuths()
	// Ignore error - just create new structure
	if err != nil || a == nil {
		a = &auths{}
	}

	// Initialize the map if not exists and add the token
	if a.Auths == nil {
		a.Auths = make(map[string]*Auth)
	}
	auth := &Auth{accessToken, refreshToken, email, expires}
	a.Auths[server] = auth

	// Marshal and write to disk
	buff, err := json.Marshal(&a)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(path.Dir(AuthsFileName), 0755); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(AuthsFileName, buff, authsFilePerm); err != nil {
		return nil, err
	}
	return auth, nil
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
