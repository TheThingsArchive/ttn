// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package health

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// RegisterServer registers and returns a new Health server
func RegisterServer(s *grpc.Server) *health.Server {
	srv := health.NewServer()
	healthpb.RegisterHealthServer(s, srv)
	return srv
}
