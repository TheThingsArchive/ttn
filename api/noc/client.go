package noc

import (
	"github.com/TheThingsNetwork/ttn/api"
	"google.golang.org/grpc"
)

// NewClient is a wrapper for NewMonitorClient, initializes
// connection to MonitorServer on monitorAddr with default gRPC options
func NewClient(monitorAddr string) (cl MonitorClient, err error) {
	var conn *grpc.ClientConn
	if conn, err = grpc.Dial(monitorAddr, append(api.DialOptions, grpc.WithInsecure())...); err != nil {
		return nil, err
	}

	return NewMonitorClient(conn), nil
}
