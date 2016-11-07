// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"crypto/ecdsa"
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/cache"
	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-account-lib/tokenkey"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	pb_monitor "github.com/TheThingsNetwork/ttn/api/monitor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/logging"
	"github.com/TheThingsNetwork/ttn/utils/security"
	"github.com/apex/log"
	"github.com/spf13/viper"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

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
func New(ctx log.Interface, serviceName string, announcedAddress string) (*Component, error) {
	go func() {
		memstats := new(runtime.MemStats)
		for range time.Tick(time.Minute) {
			runtime.ReadMemStats(memstats)
			ctx.WithFields(log.Fields{
				"Goroutines": runtime.NumGoroutine(),
				"Memory":     float64(memstats.Alloc) / 1000000,
			}).Debugf("Stats")
		}
	}()

	// Disable gRPC tracing
	// SEE: https://github.com/grpc/grpc-go/issues/695
	grpc.EnableTracing = false

	grpclog.SetLogger(logging.NewGRPCLogger(ctx))

	component := &Component{
		Ctx: ctx,
		Identity: &pb_discovery.Announcement{
			Id:             viper.GetString("id"),
			Description:    viper.GetString("description"),
			ServiceName:    serviceName,
			ServiceVersion: fmt.Sprintf("%s-%s (%s)", viper.GetString("version"), viper.GetString("gitCommit"), viper.GetString("buildDate")),
			NetAddress:     announcedAddress,
			Public:         viper.GetBool("public"),
		},
		AccessToken: viper.GetString("auth-token"),
		TokenKeyProvider: tokenkey.HTTPProvider(
			viper.GetStringMapString("auth-servers"),
			cache.WriteTroughCacheWithFormat(viper.GetString("key-dir"), "auth-%s.pub"),
		),
	}

	if serviceName != "discovery" {
		var err error
		component.Discovery, err = pb_discovery.NewClient(
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

	if priv, err := security.LoadKeypair(viper.GetString("key-dir")); err == nil {
		component.privateKey = priv

		pubPEM, _ := security.PublicPEM(priv)
		component.Identity.PublicKey = string(pubPEM)

		privPEM, _ := security.PrivatePEM(priv)

		if viper.GetBool("tls") {
			cert, err := security.LoadCert(viper.GetString("key-dir"))
			if err != nil {
				return nil, err
			}
			component.Identity.Certificate = string(cert)

			cer, err := tls.X509KeyPair(cert, privPEM)
			if err != nil {
				return nil, err
			}
			component.tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
		}
	}

	if healthPort := viper.GetInt("health-port"); healthPort > 0 {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
			switch component.GetStatus() {
			case StatusHealthy:
				w.WriteHeader(200)
				w.Write([]byte("Status is HEALTHY"))
				return
			case StatusUnhealthy:
				w.WriteHeader(503)
				w.Write([]byte("Status is UNHEALTHY"))
				return
			}
		})
		go http.ListenAndServe(fmt.Sprintf(":%d", healthPort), nil)
	}

	if monitors := viper.GetStringMapString("monitor-servers"); len(monitors) != 0 {
		component.Monitors = make(map[string]*pb_monitor.Client)
		for name, addr := range monitors {
			var err error
			component.Monitors[name], err = pb_monitor.NewClient(ctx.WithField("Monitor", name), addr)
			if err != nil {
				// Assuming grpc.WithBlock() is not set
				return nil, err
			}
		}
	}

	return component, nil
}

// Component contains the common attributes for all TTN components
type Component struct {
	Identity         *pb_discovery.Announcement
	Discovery        pb_discovery.Client
	Monitors         map[string]*pb_monitor.Client
	Ctx              log.Interface
	AccessToken      string
	privateKey       *ecdsa.PrivateKey
	tlsConfig        *tls.Config
	TokenKeyProvider tokenkey.Provider
	status           int64
}

// UpdateTokenKey updates the OAuth Bearer token key
func (c *Component) UpdateTokenKey() error {
	if c.TokenKeyProvider == nil {
		return errors.NewErrInternal("No public key provider configured for token validation")
	}

	// Set up Auth Server Token Validation
	err := c.TokenKeyProvider.Update()
	if err != nil {
		c.Ctx.Warnf("ttn: Failed to refresh public keys for token validation: %s", err.Error())
	} else {
		c.Ctx.Info("ttn: Got public keys for token validation")
	}

	return nil

}
