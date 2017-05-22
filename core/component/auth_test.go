// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/security"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/assertions"
	"google.golang.org/grpc/metadata"
)

func TestParseAuthServer(t *testing.T) {
	a := assertions.New(t)
	{
		srv, err := parseAuthServer("https://user:pass@account.thethingsnetwork.org/")
		a.So(err, assertions.ShouldBeNil)
		a.So(srv.url, assertions.ShouldEqual, "https://account.thethingsnetwork.org")
		a.So(srv.username, assertions.ShouldEqual, "user")
		a.So(srv.password, assertions.ShouldEqual, "pass")
	}
	{
		srv, err := parseAuthServer("https://user@account.thethingsnetwork.org/")
		a.So(err, assertions.ShouldBeNil)
		a.So(srv.url, assertions.ShouldEqual, "https://account.thethingsnetwork.org")
		a.So(srv.username, assertions.ShouldEqual, "user")
	}
	{
		srv, err := parseAuthServer("http://account.thethingsnetwork.org/")
		a.So(err, assertions.ShouldBeNil)
		a.So(srv.url, assertions.ShouldEqual, "http://account.thethingsnetwork.org")
	}
	{
		srv, err := parseAuthServer("http://localhost:9090/")
		a.So(err, assertions.ShouldBeNil)
		a.So(srv.url, assertions.ShouldEqual, "http://localhost:9090")
	}
}

func TestInitAuthServers(t *testing.T) {
	for _, env := range strings.Split("ACCOUNT_SERVER_PROTO ACCOUNT_SERVER_USERNAME ACCOUNT_SERVER_PASSWORD ACCOUNT_SERVER_URL", " ") {
		if os.Getenv(env) == "" {
			t.Skipf("Skipping auth server test: %s not configured", env)
		}
	}

	a := assertions.New(t)
	c := new(Component)
	c.Config.KeyDir = os.TempDir()
	c.Ctx = GetLogger(t, "TestInitAuthServers")
	c.Config.AuthServers = map[string]string{
		"ttn": fmt.Sprintf("%s://%s",
			os.Getenv("ACCOUNT_SERVER_PROTO"),
			os.Getenv("ACCOUNT_SERVER_URL"),
		),
		"ttn-user": fmt.Sprintf("%s://%s@%s",
			os.Getenv("ACCOUNT_SERVER_PROTO"),
			os.Getenv("ACCOUNT_SERVER_USERNAME"),
			os.Getenv("ACCOUNT_SERVER_URL"),
		),
		"ttn-user-pass": fmt.Sprintf("%s://%s:%s@%s",
			os.Getenv("ACCOUNT_SERVER_PROTO"),
			os.Getenv("ACCOUNT_SERVER_USERNAME"),
			os.Getenv("ACCOUNT_SERVER_PASSWORD"),
			os.Getenv("ACCOUNT_SERVER_URL"),
		),
	}
	err := c.initAuthServers()
	a.So(err, assertions.ShouldBeNil)

	{
		k, err := c.TokenKeyProvider.Get("ttn", true)
		a.So(err, assertions.ShouldBeNil)
		a.So(k.Algorithm, assertions.ShouldEqual, "RS256")
	}

	{
		k, err := c.TokenKeyProvider.Get("ttn-user", true)
		a.So(err, assertions.ShouldBeNil)
		a.So(k.Algorithm, assertions.ShouldEqual, "RS256")
	}

	{
		k, err := c.TokenKeyProvider.Get("ttn-user-pass", true)
		a.So(err, assertions.ShouldBeNil)
		a.So(k.Algorithm, assertions.ShouldEqual, "RS256")
	}

	a.So(c.UpdateTokenKey(), assertions.ShouldBeNil)
}

func TestValidateTTNAuthContext(t *testing.T) {
	for _, env := range strings.Split("ACCOUNT_SERVER_PROTO ACCOUNT_SERVER_URL", " ") {
		if os.Getenv(env) == "" {
			t.Skipf("Skipping auth server test: %s not configured", env)
		}
	}
	accountServer := fmt.Sprintf("%s://%s",
		os.Getenv("ACCOUNT_SERVER_PROTO"),
		os.Getenv("ACCOUNT_SERVER_URL"),
	)

	a := assertions.New(t)
	c := new(Component)
	c.Config.KeyDir = os.TempDir()
	c.Config.AuthServers = map[string]string{
		"ttn-account-v2": accountServer,
	}
	err := c.initAuthServers()
	a.So(err, assertions.ShouldBeNil)

	{
		ctx := context.Background()
		_, err = c.ValidateTTNAuthContext(ctx)
		a.So(err, assertions.ShouldNotBeNil)
	}

	{
		md := metadata.Pairs()
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err = c.ValidateTTNAuthContext(ctx)
		a.So(err, assertions.ShouldNotBeNil)
	}

	{
		md := metadata.Pairs(
			"id", "dev",
		)
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err = c.ValidateTTNAuthContext(ctx)
		a.So(err, assertions.ShouldNotBeNil)
	}

	{
		md := metadata.Pairs(
			"id", "dev",
			"token", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ0dG4tYWNjb3VudC12MiIsInN1YiI6ImRldiIsInR5cGUiOiJnYXRld2F5IiwiaWF0IjoxNDgyNDIxMTEyfQ.obhobeREK9bOpi-YO5lZ8rpW4CkXZUSrRBRIjbFThhvAsj_IjkFmCovIVLsGlaDVEKciZmXmWnY-6ZEgUEu6H6_GG4AD6HNHXnT0o0XSPgf5_Bc6dpzuI5FCEpcELihpBMaW3vPUt29NecLo4LvZGAuOllUYKHsZi34GYnR6PFlOgi40drN_iU_8aMCxFxm6ki83QlcyHEmDAh5GAGIym0qnUDh5_L1VE_upmoR72j8_l5lSuUA2_w8CH5_Z9CrXlTKQ2XQXsQXprkhbmOKKC8rfbTjRsB_nxObu0qcTWLH9tMd4KGFkJ20mdMw38fg2Vt7eLrkU1R1kl6a65eo6LZi0JvRSsboVZFWLwI02Azkwsm903K5n1r25Wq2oiwPJpNq5vsYLdYlb-WdAPsEDnfQGLPaqxd5we8tDcHsF4C1JHTwLsKy2Sqj8WNVmLgXiFER0DNfISDgS5SYdOxd9dUf5lTlIYdJU6aG1yYLSEhq80QOcdhCqNMVu1uRIucn_BhHbKo_LCMmD7TGppaXcQ2tCL3qHQaW8GCoun_UPo4C67LIMYUMfwd_h6CaykzlZvDlLa64ZiQ3XPmMcT_gVT7MJS2jGPbtJmcLHAVa5NZLv2d6WZfutPAocl3bYrY-sQmaSwJrzakIb2D-DNsg0qBJAZcm2o021By8U4bKAAFQ",
		)
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err = c.ValidateTTNAuthContext(ctx)
		a.So(err, assertions.ShouldBeNil)
	}
}

func TestExchangeAppKeyForToken(t *testing.T) {
	for _, env := range strings.Split("ACCOUNT_SERVER_PROTO ACCOUNT_SERVER_USERNAME ACCOUNT_SERVER_PASSWORD ACCOUNT_SERVER_URL APP_ID APP_TOKEN", " ") {
		if os.Getenv(env) == "" {
			t.Skipf("Skipping auth server test: %s not configured", env)
		}
	}

	a := assertions.New(t)

	{
		c := new(Component)
		c.Config.KeyDir = os.TempDir()
		c.Config.AuthServers = map[string]string{
			"ttn-account-v2": fmt.Sprintf("%s://%s:%s@%s",
				os.Getenv("ACCOUNT_SERVER_PROTO"),
				os.Getenv("ACCOUNT_SERVER_USERNAME"),
				os.Getenv("ACCOUNT_SERVER_PASSWORD"),
				os.Getenv("ACCOUNT_SERVER_URL"),
			),
		}
		c.initAuthServers()

		{
			token, err := c.ExchangeAppKeyForToken(os.Getenv("APP_ID"), "ttn-account-v2."+os.Getenv("APP_TOKEN"))
			a.So(err, assertions.ShouldBeNil)
			a.So(token, assertions.ShouldNotBeEmpty)
		}

		{
			token, err := c.ExchangeAppKeyForToken(os.Getenv("APP_ID"), os.Getenv("APP_TOKEN"))
			a.So(err, assertions.ShouldBeNil)
			a.So(token, assertions.ShouldNotBeEmpty)
		}
	}

	if componentToken := os.Getenv("COMPONENT_TOKEN"); componentToken != "" {
		c := new(Component)
		c.Config.KeyDir = os.TempDir()
		c.Config.AuthServers = map[string]string{
			"ttn-account-v2": fmt.Sprintf("%s://%s",
				os.Getenv("ACCOUNT_SERVER_PROTO"),
				os.Getenv("ACCOUNT_SERVER_URL"),
			),
		}
		c.AccessToken = componentToken
		c.initAuthServers()

		{
			token, err := c.ExchangeAppKeyForToken(os.Getenv("APP_ID"), "ttn-account-v2."+os.Getenv("APP_TOKEN"))
			a.So(err, assertions.ShouldBeNil)
			a.So(token, assertions.ShouldNotBeEmpty)
		}

		{
			token, err := c.ExchangeAppKeyForToken(os.Getenv("APP_ID"), os.Getenv("APP_TOKEN"))
			a.So(err, assertions.ShouldBeNil)
			a.So(token, assertions.ShouldNotBeEmpty)
		}
	}
}

func TestInitKeyPair(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tmpDir := fmt.Sprintf("%s/%d", os.TempDir(), r.Int63())
	os.Mkdir(tmpDir, 755)
	defer os.Remove(tmpDir)

	a := assertions.New(t)
	c := new(Component)
	c.Identity = new(discovery.Announcement)
	c.Config.KeyDir = tmpDir

	a.So(c.initKeyPair(), assertions.ShouldNotBeNil)

	security.GenerateKeypair(tmpDir)

	a.So(c.initKeyPair(), assertions.ShouldBeNil)

	a.So(c.Identity.PublicKey, assertions.ShouldNotBeEmpty)
	a.So(c.privateKey, assertions.ShouldNotBeNil)
}

func TestInitTLS(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tmpDir := fmt.Sprintf("%s/%d", os.TempDir(), r.Int63())
	os.Mkdir(tmpDir, 755)
	defer os.Remove(tmpDir)

	a := assertions.New(t)
	c := new(Component)
	c.Identity = new(discovery.Announcement)
	c.Config.KeyDir = tmpDir

	security.GenerateKeypair(tmpDir)
	c.initKeyPair()

	a.So(c.initTLS(), assertions.ShouldNotBeNil)

	security.GenerateCert(tmpDir)

	a.So(c.initTLS(), assertions.ShouldBeNil)

	a.So(c.Identity.Certificate, assertions.ShouldNotBeEmpty)
	a.So(c.tlsConfig, assertions.ShouldNotBeNil)
}

func TestInitAuth(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tmpDir := fmt.Sprintf("%s/%d", os.TempDir(), r.Int63())
	os.Mkdir(tmpDir, 755)
	defer os.Remove(tmpDir)

	a := assertions.New(t)
	c := new(Component)
	c.Identity = new(discovery.Announcement)
	c.Config.KeyDir = tmpDir

	security.GenerateKeypair(tmpDir)

	a.So(c.InitAuth(), assertions.ShouldBeNil)
}

func TestGetAndVerifyContext(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tmpDir := fmt.Sprintf("%s/%d", os.TempDir(), r.Int63())
	os.Mkdir(tmpDir, 755)
	defer os.Remove(tmpDir)

	a := assertions.New(t)
	c := new(Component)

	c.Identity = new(discovery.Announcement)

	c.Config.KeyDir = tmpDir
	security.GenerateKeypair(tmpDir)
	c.initKeyPair()

	{
		ctx := c.GetContext("")
		ctx = metadata.NewIncomingContext(ctx, ttnctx.MetadataFromOutgoingContext(ctx)) // Transform outgoing ctx into incoming ctx
		_, err := c.ValidateNetworkContext(ctx)
		a.So(err, assertions.ShouldNotBeNil)
	}

	c.Identity.Id = "test-context"
	{
		ctx := c.GetContext("")
		ctx = metadata.NewIncomingContext(ctx, ttnctx.MetadataFromOutgoingContext(ctx)) // Transform outgoing ctx into incoming ctx
		_, err := c.ValidateNetworkContext(ctx)
		a.So(err, assertions.ShouldNotBeNil)
	}

	c.Identity.ServiceName = "test-service"

	c.initBgCtx()

	ctrl := gomock.NewController(t)
	discoveryClient := discovery.NewMockClient(ctrl)
	c.Discovery = discoveryClient

	discoveryClient.EXPECT().Get("test-service", "test-context").Return(c.Identity, nil)

	ctx := c.GetContext("")
	ctx = metadata.NewIncomingContext(ctx, ttnctx.MetadataFromOutgoingContext(ctx)) // Transform outgoing ctx into incoming ctx
	_, err := c.ValidateNetworkContext(ctx)
	a.So(err, assertions.ShouldBeNil)

}
