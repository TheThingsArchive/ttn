package api

import (
	"context"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"google.golang.org/grpc/metadata"
)

var ErrContext = errors.NewErrInternal("Could not get metadata from context")

var ErrNoToken = errors.NewErrInvalidArgument("Metadata", "token missing")
var ErrNoID = errors.NewErrInvalidArgument("Metadata", "id missing")

func MetadataFromContext(ctx context.Context) (md metadata.MD, err error) {
	var ok bool
	if md, ok = metadata.FromContext(ctx); !ok {
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
