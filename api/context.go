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
