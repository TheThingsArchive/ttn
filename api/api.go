package api

import (
	"time"

	"google.golang.org/grpc"
)

// DialOptions are the gRPC dial options for discovery calls
// TODO: disable insecure connections
var DialOptions = []grpc.DialOption{
	grpc.WithInsecure(),
	grpc.WithTimeout(2 * time.Second),
}
