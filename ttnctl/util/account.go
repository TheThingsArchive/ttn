// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/go-account-lib/cache"
	"github.com/TheThingsNetwork/go-account-lib/tokens"
	accountUtil "github.com/TheThingsNetwork/go-account-lib/util"
	"github.com/apex/log"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var tokenSource oauth2.TokenSource

func tokenName() string {
	return viper.GetString("ttn-account-server")
}

func serverKey() string {
	replacer := strings.NewReplacer("https:", "", "http:", "", "/", "", ".", "")
	return replacer.Replace(viper.GetString("ttn-account-server"))
}

func tokenFilename(name string) string {
	return serverKey() + ".token"
}

func derivedTokenFilename(name string) string {
	return serverKey() + "." + name + ".token"
}

// GetCache get's the cache that will store our tokens
func GetTokenCache() cache.Cache {
	return cache.FileCacheWithNameFn(viper.GetString("token-dir"), tokenFilename)
}

func getAccountServerTokenSource(token *oauth2.Token) oauth2.TokenSource {
	config := accountUtil.MakeConfig(viper.GetString("ttn-account-server"), "ttnctl", "", "")
	return config.TokenSource(context.Background(), token)
}

func getStoredToken(ctx log.Interface) *oauth2.Token {
	tokenCache := GetTokenCache()
	data, err := tokenCache.Get(tokenName())
	if err != nil {
		ctx.WithError(err).Fatal("Could not read stored token")
	}
	if data == nil {
		ctx.Fatal("No account information found. Please login with ttnctl user login [access code]")
	}

	token := &oauth2.Token{}
	err = json.Unmarshal(data, token)
	if err != nil {
		ctx.Fatal("Account information invalid. Please login with ttnctl user login [access code]")
	}
	return token
}

func saveToken(ctx log.Interface, token *oauth2.Token) {
	data, err := json.Marshal(token)
	if err != nil {
		ctx.WithError(err).Fatal("Could not save access token")
	}
	err = GetTokenCache().Set(tokenName(), data)
	if err != nil {
		ctx.WithError(err).Fatal("Could not save access token")
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

func GetTokenManager(accessToken string) tokens.Manager {
	server := viper.GetString("ttn-account-server")
	return tokens.HTTPManager(server, accessToken, tokens.FileStoreWithNameFn(viper.GetString("token-dir"), derivedTokenFilename))
}

// GetAccount gets a new Account server client for ttnctl
func GetAccount(ctx log.Interface) *account.Account {
	token, err := GetTokenSource(ctx).Token()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get access token")
	}

	server := viper.GetString("ttn-account-server")
	manager := GetTokenManager(token.AccessToken)

	return account.NewWithManager(server, token.AccessToken, manager)
}

// Login does a login to the Account server with the given username and password
func Login(ctx log.Interface, code string) (*oauth2.Token, error) {
	config := accountUtil.MakeConfig(viper.GetString("ttn-account-server"), "ttnctl", "", "")
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	saveToken(ctx, token)
	return token, nil
}

func TokenForScope(ctx log.Interface, scope string) string {
	token, err := GetTokenSource(ctx).Token()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get token")
	}

	restricted, err := GetTokenManager(token.AccessToken).TokenForScope(scope)
	if err != nil {
		ctx.WithError(err).Fatal("Could not get correct rights")
	}

	return restricted
}
