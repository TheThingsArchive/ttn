// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"strconv"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc/metadata"
)

// Errors that are returned when an item could not be retrieved
var (
	ErrContext = errors.NewErrInternal("Could not get metadata from context")
	ErrNoToken = errors.NewErrInvalidArgument("Metadata", "token missing")
	ErrNoKey   = errors.NewErrInvalidArgument("Metadata", "key missing")
	ErrNoID    = errors.NewErrInvalidArgument("Metadata", "id missing")
)

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

func KeyFromMetadata(md metadata.MD) (string, error) {
	key, ok := md["key"]
	if !ok || len(key) == 0 {
		return "", ErrNoKey
	}
	return key[0], nil
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 0, 64)
}
func OffsetFromMetadata(md metadata.MD) (int64, error) {
	offset, ok := md["offset"]
	if !ok || len(offset) == 0 {
		return 0, nil
	}
	return parseInt64(offset[0])
}

func LimitFromMetadata(md metadata.MD) (int64, error) {
	limit, ok := md["limit"]
	if !ok || len(limit) == 0 {
		return 0, nil
	}
	return parseInt64(limit[0])
}
