// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
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
	AccessToken      string `json:"access_token,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
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

// Logout deletes the token for the specified server
func Logout(server string) error {
	a, err := loadAuths()
	if err != nil {
		return err
	}
	delete(a.Auths, server)
	if err := saveAuths(a); err != nil {
		return err
	}
	return nil
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
	if time.Now().UTC().After(auth.Expires) {
		return refreshToken(server, auth)
	}
	return auth, nil
}

// NewRequestWithAuth creates a new HTTP request and adds the access token of
// the authenticated user as bearer token
func NewRequestWithAuth(server, method, urlStr string, body io.Reader) (*http.Request, error) {
	auth, err := LoadAuth(server)
	if err != nil {
		return nil, err
	}
	if auth == nil {
		return nil, errors.New("No authentication token found. Please login")
	}
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.AccessToken))
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// RefreshToken refreshes the current token
func RefreshToken(server string) (*Auth, error) {
	auth, err := LoadAuth(server)
	if err != nil {
		return nil, err
	}
	return refreshToken(server, auth)
}

func refreshToken(server string, auth *Auth) (*Auth, error) {
	values := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {auth.RefreshToken},
	}
	return newToken(server, auth.Email, values)
}

func newToken(server, email string, values url.Values) (*Auth, error) {
	uri := fmt.Sprintf("%s/users/token", server)
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

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("Incorrect username or password.")
	}

	if resp.StatusCode != http.StatusOK {
		if t.Error != "" {
			return nil, errors.New(t.ErrorDescription)
		}
		return nil, errors.New(resp.Status)
	}

	expires := time.Now().UTC().Add(time.Duration(t.ExpiresIn) * time.Second)
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
	auth := &Auth{accessToken, refreshToken, email, expires}
	a.Auths[server] = auth
	if err := saveAuths(a); err != nil {
		return nil, err
	}

	return auth, nil
}

// loadAuths loads the authentication tokens. This function returns an empty
// structure if the file does not exist.
func loadAuths() (*auths, error) {
	if _, err := os.Stat(AuthsFileName); os.IsNotExist(err) {
		return &auths{make(map[string]*Auth)}, nil
	}
	buff, err := ioutil.ReadFile(AuthsFileName)
	if err != nil {
		return nil, err
	}
	var a auths
	if err := json.Unmarshal(buff, &a); err != nil {
		return nil, err
	}
	if a.Auths == nil {
		a.Auths = make(map[string]*Auth)
	}
	return &a, nil
}

func saveAuths(a *auths) error {
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
