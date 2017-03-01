// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"

	. "github.com/smartystreets/assertions"
)

func TestContext(t *testing.T) {
	a := New(t)
	var err error

	// Errors if the context doesn't have valid metadata
	{
		ctx := context.Background()

		md := MetadataFromContext(ctx)
		a.So(md, ShouldHaveLength, 0)

		ctx = metadata.NewContext(ctx, metadata.Pairs())

		md = MetadataFromContext(ctx)
		a.So(md, ShouldHaveLength, 0)

		_, err = TokenFromContext(ctx)
		a.So(err, ShouldNotBeNil)

		_, err = KeyFromContext(ctx)
		a.So(err, ShouldNotBeNil)

		_, err = IDFromContext(ctx)
		a.So(err, ShouldNotBeNil)

		limit, offset, err := LimitAndOffsetFromContext(ctx)
		a.So(err, ShouldBeNil)
		a.So(limit, ShouldEqual, 0)
		a.So(offset, ShouldEqual, 0)

		serviceName, serviceVersion, netAddress, err := ServiceInfoFromContext(ctx)
		a.So(err, ShouldBeNil)
		a.So(serviceName, ShouldEqual, "")
		a.So(serviceVersion, ShouldEqual, "")
		a.So(netAddress, ShouldEqual, "")
	}

	// Errors if the context has wrong metadata
	{
		_, _, err := LimitAndOffsetFromContext(metadata.NewContext(context.Background(), metadata.Pairs(
			"limit", "wut",
		)))
		a.So(err, ShouldNotBeNil)

		_, _, err = LimitAndOffsetFromContext(metadata.NewContext(context.Background(), metadata.Pairs(
			"offset", "wut",
		)))
		a.So(err, ShouldNotBeNil)
	}

	{
		ctx := context.Background()

		ctx = ContextWithToken(ctx, "token")
		token, err := TokenFromContext(ctx)
		a.So(err, ShouldBeNil)
		a.So(token, ShouldEqual, "token")

		ctx = ContextWithKey(ctx, "key")
		key, err := KeyFromContext(ctx)
		a.So(err, ShouldBeNil)
		a.So(key, ShouldEqual, "key")

		ctx = ContextWithID(ctx, "id")
		id, err := IDFromContext(ctx)
		a.So(err, ShouldBeNil)
		a.So(id, ShouldEqual, "id")

		ctx = ContextWithServiceInfo(ctx, "name", "version", "addr")
		serviceName, serviceVersion, netAddress, err := ServiceInfoFromContext(ctx)
		a.So(err, ShouldBeNil)
		a.So(serviceName, ShouldEqual, "name")
		a.So(serviceVersion, ShouldEqual, "version")
		a.So(netAddress, ShouldEqual, "addr")

		ctx = ContextWithLimitAndOffset(ctx, 2, 4)
		limit, offset, err := LimitAndOffsetFromContext(ctx)
		a.So(err, ShouldBeNil)
		a.So(limit, ShouldEqual, 2)
		a.So(offset, ShouldEqual, 4)

		// Try the token again
		token, err = TokenFromContext(ctx)
		a.So(err, ShouldBeNil)
		a.So(token, ShouldEqual, "token")
	}

	{
		ctx := ContextWithLimitAndOffset(metadata.NewContext(context.Background(), metadata.Pairs()), 0, 0)
		limit, offset, err := LimitAndOffsetFromContext(ctx)
		a.So(err, ShouldBeNil)
		a.So(limit, ShouldEqual, 0)
		a.So(offset, ShouldEqual, 0)
	}

}
