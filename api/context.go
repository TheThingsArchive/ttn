// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import "context"

func TokenFromContext(ctx context.Context) (token string, err error) {
	md, err := MetadataFromContext(ctx)
	if err != nil {
		return "", err
	}
	return TokenFromMetadata(md)
}

func KeyFromContext(ctx context.Context) (key string, err error) {
	md, err := MetadataFromContext(ctx)
	if err != nil {
		return "", err
	}
	return KeyFromMetadata(md)
}

func IDFromContext(ctx context.Context) (token string, err error) {
	md, err := MetadataFromContext(ctx)
	if err != nil {
		return "", err
	}
	return IDFromMetadata(md)
}
