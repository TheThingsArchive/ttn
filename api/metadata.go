package api

import (
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc/metadata"
)

var ErrContext = errors.NewErrInternal("Could not get metadata from context")

var ErrNoToken = errors.NewErrInvalidArgument("Metadata", "token missing")
var ErrNoID = errors.NewErrInvalidArgument("Metadata", "id missing")

func MetadataFromContext(ctx context.Context) (metadata.MD, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return md, ErrContext
	}
	return md, nil
}

func IDFromMetadata(md metadata.MD) (string, error) {
	id, ok := md["id"]
	if !ok || len(id) == 0 {
		return "", ErrNoID
	}
	return id[0], nil
}

func TokenFromMetadata(md metadata.MD) (string, error) {
	token, ok := md["token"]
	if !ok || len(token) == 0 {
		return "", ErrNoToken
	}
	return token[0], nil
}
