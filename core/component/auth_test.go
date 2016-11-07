package component

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/security"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/assertions"
	"google.golang.org/grpc/metadata"
)

func TestAuthServerRegex(t *testing.T) {
	a := assertions.New(t)
	var matches []string
	matches = AuthServerRegex.FindStringSubmatch("https://user:pass@account.thethingsnetwork.org/")
	a.So(matches, assertions.ShouldResemble, []string{"https://user:pass@account.thethingsnetwork.org/", "https://", "user", "pass", "account.thethingsnetwork.org"})
	matches = AuthServerRegex.FindStringSubmatch("https://user@account.thethingsnetwork.org/")
	a.So(matches, assertions.ShouldResemble, []string{"https://user@account.thethingsnetwork.org/", "https://", "user", "", "account.thethingsnetwork.org"})
	matches = AuthServerRegex.FindStringSubmatch("https://account.thethingsnetwork.org/")
	a.So(matches, assertions.ShouldResemble, []string{"https://account.thethingsnetwork.org/", "https://", "", "", "account.thethingsnetwork.org"})
}

func TestInitAuthServers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	a := assertions.New(t)
	c := new(Component)
	c.Config.KeyDir = os.TempDir()
	c.Ctx = GetLogger(t, "TestInitAuthServers")
	c.Config.AuthServers = map[string]string{
		"ttn":           "https://account.thethingsnetwork.org/",
		"ttn-user":      "https://user@account.thethingsnetwork.org/",
		"ttn-user-pass": "https://user:pass@account.thethingsnetwork.org/",
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
	a := assertions.New(t)
	c := new(Component)
	c.Config.KeyDir = os.TempDir()
	c.Config.AuthServers = map[string]string{
		"ttn-account-preview": "https://preview.account.thethingsnetwork.org/",
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
		ctx := metadata.NewContext(context.Background(), md)
		_, err = c.ValidateTTNAuthContext(ctx)
		a.So(err, assertions.ShouldNotBeNil)
	}

	{
		md := metadata.Pairs(
			"id", "dev",
		)
		ctx := metadata.NewContext(context.Background(), md)
		_, err = c.ValidateTTNAuthContext(ctx)
		a.So(err, assertions.ShouldNotBeNil)
	}

	{
		md := metadata.Pairs(
			"id", "dev",
			"token", "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJ0dG4tYWNjb3VudC1wcmV2aWV3Iiwic3ViIjoiZGV2IiwidHlwZSI6InJvdXRlciIsImlhdCI6MTQ3NjQzOTQzOH0.Duz-E5aMYEPY_Nf5Pky7Qmjbs1dMp9PN9nMqbSzoU079b8TPL4DH2SKcRHrrMqieB3yhJb3YaQBfY6dKWfgVz8BmTeKlGXfFrqEj91y30J7r9_VsHRzgDMJedlqXryvf0S_yD27TsJ7TMbGYyE00T4tAX3Uf6wQZDhdyHNGtdf4jtoAjzOxVAodNtXZp26LR7fFk56UstBxOxztBMzyzmAdiTG4lSyEqq7zsuJcFjmHB9MfEoD4ZT-iTRL1ohFjGuj2HN49oPyYlZAVPP7QajLyNsLnv-nDqXE_QecOjAcEq4PLNJ3DpXtX-lo8I_F1eV9yQnDdQQi4EUvxmxZWeBA",
		)
		ctx := metadata.NewContext(context.Background(), md)
		_, err = c.ValidateTTNAuthContext(ctx)
		a.So(err, assertions.ShouldBeNil)
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

func TestInit(t *testing.T) {
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
		_, err := c.ValidateNetworkContext(ctx)
		a.So(err, assertions.ShouldNotBeNil)
	}

	c.Identity.Id = "test-context"
	{
		ctx := c.GetContext("")
		_, err := c.ValidateNetworkContext(ctx)
		a.So(err, assertions.ShouldNotBeNil)
	}

	c.Identity.ServiceName = "test-service"

	ctrl := gomock.NewController(t)
	discoveryClient := discovery.NewMockClient(ctrl)
	c.Discovery = discoveryClient

	discoveryClient.EXPECT().Get("test-service", "test-context").Return(c.Identity, nil)

	ctx := c.GetContext("")
	_, err := c.ValidateNetworkContext(ctx)
	a.So(err, assertions.ShouldBeNil)

}
