// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package health

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Check the health of a connection
func Check(conn *grpc.ClientConn) (bool, error) {
	res, err := healthpb.NewHealthClient(conn).Check(context.Background(), &healthpb.HealthCheckRequest{})
	if err != nil {
		return false, err
	}
	return res.Status == healthpb.HealthCheckResponse_SERVING, nil
}
