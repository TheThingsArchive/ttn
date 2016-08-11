// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"time"

	"github.com/TheThingsNetwork/ttn/core/account"
	accountUtil "github.com/TheThingsNetwork/ttn/core/account/util"
	"github.com/apex/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// AccountClaims are extracted from the access token
type AccountClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Name     struct {
		First string `json:"first"`
		Last  string `json:"last"`
	} `json:"name"`
	Email  string              `json:"email"`
	Client string              `json:"client"`
	Scopes []string            `json:"scope"`
	Apps   map[string][]string `json:"apps,omitempty"`
}

var tokenSource oauth2.TokenSource

func getAccountServerTokenSource(token *oauth2.Token) oauth2.TokenSource {
	config := accountUtil.MakeConfig(viper.GetString("ttn-account-server"), "ttnctl", "", "")
	return config.TokenSource(context.Background(), token)
}

func getStoredToken(ctx log.Interface) *oauth2.Token {
	tokenString := viper.GetString("oauth2-token")
	if tokenString == "" {
		ctx.Fatal("No account information found. Please login with ttnctl user login [e-mail]")
	}
	token := &oauth2.Token{}
	err := json.Unmarshal([]byte(tokenString), token)
	if err != nil {
		ctx.Fatal("Account information invalid. Please login with ttnctl user login [e-mail]")
	}
	return token
}

func saveToken(ctx log.Interface, token *oauth2.Token) {
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		ctx.WithError(err).Fatal("Could not save access token")
	}
	if viper.GetString("oauth2-token") != string(tokenBytes) {
		config, _ := ReadConfig()
		if config == nil {
			config = map[string]interface{}{}
		}
		config["oauth2-token"] = string(tokenBytes)
		err = WriteConfigFile(config)
		if err != nil {
			ctx.WithError(err).Fatal("Could not save access token")
		}
	}
}

type ttnctlTokenSource struct {
	ctx    log.Interface
	source oauth2.TokenSource
}

func (s *ttnctlTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.source.Token()
	if err != nil {
		return nil, err
	}
	saveToken(s.ctx, token)
	return token, nil
}

// ForceRefreshToken forces a refresh of the access token
func ForceRefreshToken(ctx log.Interface) {
	tokenSource := GetTokenSource(ctx).(*ttnctlTokenSource)
	token, err := tokenSource.Token()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get access token")
	}
	token.Expiry = time.Now().Add(-1 * time.Second)
	tokenSource.source = oauth2.ReuseTokenSource(token, getAccountServerTokenSource(token))
	tokenSource.Token()
}

// GetTokenSource builds a new oauth2.TokenSource that uses the ttnctl config to store the token
func GetTokenSource(ctx log.Interface) oauth2.TokenSource {
	if tokenSource != nil {
		return tokenSource
	}
	token := getStoredToken(ctx)
	source := oauth2.ReuseTokenSource(token, getAccountServerTokenSource(token))
	tokenSource = &ttnctlTokenSource{ctx, source}
	return tokenSource
}

// GetAccount gets a new Account server client for ttnctl
func GetAccount(ctx log.Interface) *account.Account {
	token, err := GetTokenSource(ctx).Token()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get access token")
	}
	return account.New(viper.GetString("ttn-account-server"), token.AccessToken)
}

// Login does a login to the Account server with the given username and password
func Login(ctx log.Interface, code string) {
	config := accountUtil.MakeConfig(viper.GetString("ttn-account-server"), "ttnctl", "", "")
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		ctx.WithError(err).Fatal("Login failed")
	}
	saveToken(ctx, token)
	ctx.Info("Login successful")
}
