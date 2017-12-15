// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package component contains code that is shared by all components (discovery, router, broker, networkserver, handler)
package component

import (
	"crypto/ecdsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	pb_discovery "github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/api/discovery/discoveryclient"
	"github.com/TheThingsNetwork/api/monitor/monitorclient"
	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-account-lib/tokenkey"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"github.com/spf13/viper"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
)

// Component contains the common attributes for all TTN components
type Component struct {
	Config           Config
	Pool             *pool.Pool
	Identity         *pb_discovery.Announcement
	Discovery        discoveryclient.Client
	Monitor          *monitorclient.MonitorClient
	Ctx              ttnlog.Interface
	Context          context.Context
	AccessToken      string
	privateKey       *ecdsa.PrivateKey
	tlsConfig        *tls.Config
	TokenKeyProvider tokenkey.Provider
	status           int32
	healthServer     *health.Server
}

type Interface interface {
	RegisterRPC(s *grpc.Server)
	Init(c *Component) error
	Shutdown()
	ValidateNetworkContext(ctx context.Context) (*pb_discovery.Announcement, error)
	ValidateTTNAuthContext(ctx context.Context) (*claims.Claims, error)
}

type ManagementInterface interface {
	RegisterManager(s *grpc.Server)
}

// New creates a new Component
func New(ctx ttnlog.Interface, serviceName string, announcedAddress string) (*Component, error) {
	component := &Component{
		Config: ConfigFromViper(),
		Ctx:    ctx,
		Identity: &pb_discovery.Announcement{
			ID:             viper.GetString("id"),
			Description:    viper.GetString("description"),
			ServiceName:    serviceName,
			ServiceVersion: fmt.Sprintf("%s-%s (%s)", viper.GetString("version"), viper.GetString("gitCommit"), viper.GetString("buildDate")),
			NetAddress:     announcedAddress,
			Public:         viper.GetBool("public"),
		},
		AccessToken: viper.GetString("auth-token"),
		Pool:        pool.NewPool(context.Background(), pool.DefaultDialOptions...),
	}

	if err := component.initialize(); err != nil {
		return nil, err
	}

	if err := component.InitAuth(); err != nil {
		return nil, err
	}

	if claims, err := claims.FromToken(component.TokenKeyProvider, component.AccessToken); err == nil {
		tokenExpiry.WithLabelValues(component.Identity.ServiceName, component.Identity.ID).Set(float64(claims.ExpiresAt))
	}

	if p, _ := pem.Decode([]byte(component.Identity.Certificate)); p != nil && p.Type == "CERTIFICATE" {
		if cert, err := x509.ParseCertificate(p.Bytes); err == nil {
			sum := sha1.Sum(cert.Raw)
			certificateExpiry.WithLabelValues(hex.EncodeToString(sum[:])).Set(float64(cert.NotAfter.Unix()))
		}
	}

	if serviceName != "discovery" && serviceName != "networkserver" {
		var err error
		component.Discovery, err = discoveryclient.NewClient(
			viper.GetString("discovery-address"),
			component.Identity,
			func() string {
				token, _ := component.BuildJWT()
				return token
			},
		)
		if err != nil {
			return nil, err
		}
	}

	var monitorOpts []monitorclient.MonitorOption
	for name, addr := range viper.GetStringMapString("monitor-servers") {
		monitorOpts = append(monitorOpts, monitorclient.WithServer(name, addr))
	}
	component.Monitor = monitorclient.NewMonitorClient(monitorOpts...)

	return component, nil
}
