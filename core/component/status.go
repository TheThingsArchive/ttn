// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Status indicates the health status of this component
type Status int

const (
	// StatusHealthy indicates a healthy component
	StatusHealthy Status = iota
	// StatusUnhealthy indicates an unhealthy component
	StatusUnhealthy
)

var statusName = "ttn"

var statusMu sync.RWMutex

var status = make(map[*Component]Status)

func setStatus(c *Component, s Status) {
	statusMu.Lock()
	defer statusMu.Unlock()
	status[c] = s
	if c.healthServer != nil {
		switch s {
		case StatusHealthy:
			c.healthServer.SetServingStatus(statusName, healthpb.HealthCheckResponse_SERVING)
		case StatusUnhealthy:
			c.healthServer.SetServingStatus(statusName, healthpb.HealthCheckResponse_NOT_SERVING)
		}
	}
}

func getStatus(c *Component) Status {
	statusMu.RLock()
	defer statusMu.RUnlock()
	if s, ok := status[c]; ok {
		return s
	}
	return StatusUnhealthy
}

func getStatusPage(c *Component) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		switch getStatus(c) {
		case StatusHealthy:
			w.WriteHeader(200)
			w.Write([]byte("Status is HEALTHY"))
			return
		case StatusUnhealthy:
			w.WriteHeader(503)
			w.Write([]byte("Status is UNHEALTHY"))
			return
		}
	}
}

// GetStatus gets the health status of the component
func (c *Component) GetStatus() Status {
	return getStatus(c)
}

// SetStatus sets the health status of the component
func (c *Component) SetStatus(s Status) {
	setStatus(c, s)
}

// RegisterHealthServer registers the component's health status to the gRPC server
func (c *Component) RegisterHealthServer(srv *grpc.Server) {
	c.healthServer = health.NewServer()
	healthpb.RegisterHealthServer(srv, c.healthServer)
	grpc_prometheus.Register(srv)
}

func initStatus(c *Component) error {
	setStatus(c, StatusUnhealthy)
	if healthPort := viper.GetInt("health-port"); healthPort > 0 {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/healthz", getStatusPage(c))
		go func() {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", healthPort), nil); err != nil {
				c.Ctx.WithError(err).Error("Status server exited")
			}
		}()
	}
	return nil
}

func init() {
	OnInitialize(initStatus)
}
