// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	//"github.com/TheThingsNetwork/ttn/core/mocks"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"golang.org/x/net/context"
)

func TestHandleStats(t *testing.T) {
	{
		Desc(t, "Handle Valid Stats Request")

		// Build
		components := Components{
			Ctx:     GetLogger(t, "Router"),
			Storage: NewMockStorage(),
		}
		r := New(components, Options{})
		req := &core.StatsReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata: &core.StatsMetadata{
				Altitude:  -14,
				Longitude: 43.333,
				Latitude:  -2.342,
			},
		}

		// Expect
		var wantErr *string
		var wantRes *core.StatsRes
		var wantID = req.GatewayID
		var wantMeta = *req.Metadata

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantID, components.Storage.(*MockStorage).InUpdateStats.GID, "Gateway IDs")
		Check(t, wantMeta, components.Storage.(*MockStorage).InUpdateStats.Metadata, "Gateways Metas")
	}

	// --------------------

	{
		Desc(t, "Handle Nil Stats Requests")

		// Build
		components := Components{
			Ctx:     GetLogger(t, "Router"),
			Storage: NewMockStorage(),
		}
		r := New(components, Options{})
		var req *core.StatsReq

		// Expect
		var wantErr = ErrStructural
		var wantRes *core.StatsRes
		var wantID []byte
		var wantMeta core.StatsMetadata

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantID, components.Storage.(*MockStorage).InUpdateStats.GID, "Gateway IDs")
		Check(t, wantMeta, components.Storage.(*MockStorage).InUpdateStats.Metadata, "Gateways Metas")
	}

	// --------------------

	{
		Desc(t, "Handle Stats Request | Invalid GatewayID")

		// Build
		components := Components{
			Ctx:     GetLogger(t, "Router"),
			Storage: NewMockStorage(),
		}
		r := New(components, Options{})
		req := &core.StatsReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			Metadata: &core.StatsMetadata{
				Altitude:  -14,
				Longitude: 43.333,
				Latitude:  -2.342,
			},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes *core.StatsRes
		var wantID []byte
		var wantMeta core.StatsMetadata

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantID, components.Storage.(*MockStorage).InUpdateStats.GID, "Gateway IDs")
		Check(t, wantMeta, components.Storage.(*MockStorage).InUpdateStats.Metadata, "Gateways Metas")
	}

	// --------------------

	{
		Desc(t, "Handle Stats Request | Nil Metadata")

		// Build
		components := Components{
			Ctx:     GetLogger(t, "Router"),
			Storage: NewMockStorage(),
		}
		r := New(components, Options{})
		req := &core.StatsReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes *core.StatsRes
		var wantID []byte
		var wantMeta core.StatsMetadata

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantID, components.Storage.(*MockStorage).InUpdateStats.GID, "Gateway IDs")
		Check(t, wantMeta, components.Storage.(*MockStorage).InUpdateStats.Metadata, "Gateways Metas")
	}

	// --------------------

	{
		Desc(t, "Handle Stats Request | Storage fails ")

		// Build
		components := Components{
			Ctx:     GetLogger(t, "Router"),
			Storage: NewMockStorage(),
		}
		components.Storage.(*MockStorage).Failures["UpdateStats"] = errors.New(errors.Operational, "")
		r := New(components, Options{})
		req := &core.StatsReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata: &core.StatsMetadata{
				Altitude:  -14,
				Longitude: 43.333,
				Latitude:  -2.342,
			},
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes *core.StatsRes
		var wantID = req.GatewayID
		var wantMeta = *req.Metadata

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantID, components.Storage.(*MockStorage).InUpdateStats.GID, "Gateway IDs")
		Check(t, wantMeta, components.Storage.(*MockStorage).InUpdateStats.Metadata, "Gateways Metas")
	}
}

func TestHandleData(t *testing.T) {

}
