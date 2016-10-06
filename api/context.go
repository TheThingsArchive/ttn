package api

import (
	context "golang.org/x/net/context" //TODO change to "context", when protoc supports it
	"google.golang.org/grpc/metadata"
)

func TokenFromContext(ctx context.Context) (token string, err error) {
	var md metadata.MD
	if md, err = MetadataFromContext(ctx); err != nil {
		return "", err
	}

	return TokenFromMetadata(md)
}

func IDFromContext(ctx context.Context) (token string, err error) {
	var md metadata.MD
	if md, err = MetadataFromContext(ctx); err != nil {
		return "", err
	}

	return IDFromMetadata(md)
}
