// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import (
	"sync/atomic"

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

// GetStatus gets the health status of the component
func (c *Component) GetStatus() Status {
	return Status(atomic.LoadInt32(&c.status))
}

// SetStatus sets the health status of the component
func (c *Component) SetStatus(status Status) {
	atomic.StoreInt32(&c.status, int32(status))
	if c.healthServer != nil {
		switch status {
		case StatusHealthy:
			c.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
		case StatusUnhealthy:
			c.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		}
	}
}

// RegisterHealthServer registers the component's health status to the gRPC server
func (c *Component) RegisterHealthServer(srv *grpc.Server) {
	c.healthServer = health.NewServer()
	healthpb.RegisterHealthServer(srv, c.healthServer)
}
