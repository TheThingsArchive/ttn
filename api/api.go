package api

import "google.golang.org/grpc"

// DialOptions are the gRPC dial options for discovery calls
// TODO: disable insecure connections
var DialOptions = []grpc.DialOption{
	grpc.WithInsecure(),
}
