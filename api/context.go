// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"strconv"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// MetadataFromContext gets the metadata from the given context
func MetadataFromContext(ctx context.Context) metadata.MD {
	md, _ := metadata.FromContext(ctx)
	return md
}

func contextWithMergedMetadata(ctx context.Context, kv ...string) context.Context {
	md := MetadataFromContext(ctx)
	md = metadata.Join(metadata.Pairs(kv...), md)
	return metadata.NewContext(ctx, md)
}

// TokenFromMetadata gets the token from the metadata or returns ErrNoToken
func TokenFromMetadata(md metadata.MD) (string, error) {
	token, ok := md["token"]
	if !ok || len(token) == 0 {
		return "", ErrNoToken
	}
	return token[0], nil
}

// TokenFromContext gets the token from the context or returns ErrNoToken
func TokenFromContext(ctx context.Context) (string, error) {
	md := MetadataFromContext(ctx)
	return TokenFromMetadata(md)
}

// ContextWithToken returns a context with the token
func ContextWithToken(ctx context.Context, token string) context.Context {
	return contextWithMergedMetadata(ctx, "token", token)
}

// KeyFromMetadata gets the key from the metadata or returns ErrNoKey
func KeyFromMetadata(md metadata.MD) (string, error) {
	key, ok := md["key"]
	if !ok || len(key) == 0 {
		return "", ErrNoKey
	}
	return key[0], nil
}

// KeyFromContext gets the key from the context or returns ErrNoKey
func KeyFromContext(ctx context.Context) (string, error) {
	md := MetadataFromContext(ctx)
	return KeyFromMetadata(md)
}

// ContextWithKey returns a context with the key
func ContextWithKey(ctx context.Context, key string) context.Context {
	return contextWithMergedMetadata(ctx, "key", key)
}

// IDFromMetadata gets the key from the metadata or returns ErrNoID
func IDFromMetadata(md metadata.MD) (string, error) {
	id, ok := md["id"]
	if !ok || len(id) == 0 {
		return "", ErrNoID
	}
	return id[0], nil
}

// IDFromContext gets the key from the context or returns ErrNoID
func IDFromContext(ctx context.Context) (string, error) {
	md := MetadataFromContext(ctx)
	return IDFromMetadata(md)
}

// ContextWithID returns a context with the id
func ContextWithID(ctx context.Context, id string) context.Context {
	return contextWithMergedMetadata(ctx, "id", id)
}

// ServiceInfoFromMetadata gets the service information from the metadata or returns empty strings
func ServiceInfoFromMetadata(md metadata.MD) (serviceName, serviceVersion, netAddress string, err error) {
	serviceNameL, ok := md["service-name"]
	if ok && len(serviceNameL) > 0 {
		serviceName = serviceNameL[0]
	}
	serviceVersionL, ok := md["service-version"]
	if ok && len(serviceVersionL) > 0 {
		serviceVersion = serviceVersionL[0]
	}
	netAddressL, ok := md["net-address"]
	if ok && len(netAddressL) > 0 {
		netAddress = netAddressL[0]
	}
	return
}

// ServiceInfoFromContext gets the service information from the context or returns empty strings
func ServiceInfoFromContext(ctx context.Context) (serviceName, serviceVersion, netAddress string, err error) {
	md := MetadataFromContext(ctx)
	return ServiceInfoFromMetadata(md)
}

// ContextWithServiceInfo returns a context with the id
func ContextWithServiceInfo(ctx context.Context, serviceName, serviceVersion, netAddress string) context.Context {
	return contextWithMergedMetadata(ctx, "service-name", serviceName, "service-version", serviceVersion, "net-address", netAddress)
}

// LimitFromMetadata gets the limit from the metadata
func LimitFromMetadata(md metadata.MD) (uint64, error) {
	limit, ok := md["limit"]
	if !ok || len(limit) == 0 {
		return 0, nil
	}
	return strconv.ParseUint(limit[0], 10, 64)
}

// OffsetFromMetadata gets the offset from the metadata
func OffsetFromMetadata(md metadata.MD) (uint64, error) {
	offset, ok := md["offset"]
	if !ok || len(offset) == 0 {
		return 0, nil
	}
	return strconv.ParseUint(offset[0], 10, 64)
}

// LimitAndOffsetFromContext gets the limit and offset from the context
func LimitAndOffsetFromContext(ctx context.Context) (limit, offset uint64, err error) {
	md := MetadataFromContext(ctx)
	limit, err = LimitFromMetadata(md)
	if err != nil {
		return 0, 0, err
	}
	offset, err = OffsetFromMetadata(md)
	if err != nil {
		return 0, 0, err
	}
	return limit, offset, nil
}

// ContextWithLimitAndOffset returns a context with the limit and offset
func ContextWithLimitAndOffset(ctx context.Context, limit, offset uint64) context.Context {
	var pairs []string
	if limit != 0 {
		pairs = append(pairs, "limit", strconv.FormatUint(limit, 10))
	}
	if offset != 0 {
		pairs = append(pairs, "offset", strconv.FormatUint(offset, 10))
	}
	if len(pairs) == 0 {
		return ctx
	}
	return contextWithMergedMetadata(ctx, pairs...)
}

// Errors that are returned when an item could not be retrieved
var (
	ErrNoToken = errors.NewErrInvalidArgument("Metadata", "token missing")
	ErrNoKey   = errors.NewErrInvalidArgument("Metadata", "key missing")
	ErrNoID    = errors.NewErrInvalidArgument("Metadata", "id missing")
)
