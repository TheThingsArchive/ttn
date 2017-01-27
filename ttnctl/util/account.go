// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/go-account-lib/cache"
	"github.com/TheThingsNetwork/go-account-lib/oauth"
	"github.com/TheThingsNetwork/go-account-lib/tokens"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var tokenSource oauth2.TokenSource

func tokenFile() string {
	return getServerKey() + ".token"
}
func derivedTokenFile() string {
	return getServerKey() + ".derived-tokens"
}

func tokenName() string {
	return viper.GetString("auth-server")
}

func getServerKey() string {
	replacer := strings.NewReplacer("https:", "", "http:", "", "/", "", ".", "")
	return replacer.Replace(viper.GetString("auth-server"))
}

func tokenFilename(name string) string {
	return tokenFile()
}

// GetCache get's the cache that will store our tokens
func GetTokenCache() cache.Cache {
	return cache.FileCacheWithNameFn(GetDataDir(), tokenFilename)
}

func GetUserAgent() string {
	return fmt.Sprintf(
		"ttnctl/%s-%s (%s-%s) (%s)",
		viper.GetString("version"),
		viper.GetString("gitCommit"),
		runtime.GOOS,
		runtime.GOARCH,
		GetID(),
	)
}

// getOAuth gets the OAuth client
func getOAuth() *oauth.Config {
	return oauth.OAuth(viper.GetString("auth-server"), &oauth.Client{
		ID:     "ttnctl",
		Secret: "ttnctl",
		ExtraHeaders: map[string]string{
			"User-Agent": GetUserAgent(),
		},
	})
}

func getAccountServerTokenSource(token *oauth2.Token) oauth2.TokenSource {
	return getOAuth().TokenSource(token)
}

func getStoredToken(ctx ttnlog.Interface) *oauth2.Token {
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

func saveToken(ctx ttnlog.Interface, token *oauth2.Token) {
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
	ctx    ttnlog.Interface
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
func ForceRefreshToken(ctx ttnlog.Interface) {
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
func GetTokenSource(ctx ttnlog.Interface) oauth2.TokenSource {
	if tokenSource != nil {
		return tokenSource
	}
	token := getStoredToken(ctx)
	source := oauth2.ReuseTokenSource(token, getAccountServerTokenSource(token))
	tokenSource = &ttnctlTokenSource{ctx, source}
	return tokenSource
}

func GetTokenManager(accessToken string) tokens.Manager {
	server := viper.GetString("auth-server")
	return tokens.HTTPManager(server, accessToken, tokens.FileStore(path.Join(GetDataDir(), derivedTokenFile())))
}

// GetAccount gets a new Account server client for ttnctl
func GetAccount(ctx ttnlog.Interface) *account.Account {
	token, err := GetTokenSource(ctx).Token()
	if err != nil {
		ctx.WithError(err).Fatal("Could not get access token")
	}

	server := viper.GetString("auth-server")
	manager := GetTokenManager(token.AccessToken)

	return account.NewWithManager(server, token.AccessToken, manager).WithHeader("User-Agent", GetUserAgent())
}

// Login does a login to the Account server with the given username and password
func Login(ctx ttnlog.Interface, code string) (*oauth2.Token, error) {
	config := getOAuth()
	token, err := config.Exchange(code)
	if err != nil {
		return nil, err
	}
	saveToken(ctx, token)
	return token, nil
}

func TokenForScope(ctx ttnlog.Interface, scope string) string {
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

func Logout() error {
	err := os.Remove(path.Join(GetDataDir(), tokenFile()))
	if err != nil {
		return err
	}

	err = os.Remove(path.Join(GetDataDir(), derivedTokenFile()))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
