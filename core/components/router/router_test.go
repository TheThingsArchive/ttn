// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	"golang.org/x/net/context"
)

func TestHandleStats(t *testing.T) {
	{
		Desc(t, "Handle Valid Stats Request")

		// Build
		components := Components{
			Ctx:        GetLogger(t, "Router"),
			BrkStorage: NewMockBrkStorage(),
			GtwStorage: NewMockGtwStorage(),
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
		var wantRes = new(core.StatsRes)
		var wantEntry = gtwEntry{
			GatewayID: req.GatewayID,
			Metadata:  *req.Metadata,
		}

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantEntry, components.GtwStorage.(*MockGtwStorage).InUpsert.Entry, "Gateway Entries")
	}

	// --------------------

	{
		Desc(t, "Handle Nil Stats Requests")

		// Build
		components := Components{
			Ctx:        GetLogger(t, "Router"),
			BrkStorage: NewMockBrkStorage(),
			GtwStorage: NewMockGtwStorage(),
		}
		r := New(components, Options{})
		var req *core.StatsReq

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.StatsRes)
		var wantEntry gtwEntry

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantEntry, components.GtwStorage.(*MockGtwStorage).InUpsert.Entry, "Gateway Entries")
	}

	// --------------------

	{
		Desc(t, "Handle Stats Request | Invalid GatewayID")

		// Build
		components := Components{
			Ctx:        GetLogger(t, "Router"),
			BrkStorage: NewMockBrkStorage(),
			GtwStorage: NewMockGtwStorage(),
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
		var wantRes = new(core.StatsRes)
		var wantEntry gtwEntry

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantEntry, components.GtwStorage.(*MockGtwStorage).InUpsert.Entry, "Gateway Entries")
	}

	// --------------------

	{
		Desc(t, "Handle Stats Request | Nil Metadata")

		// Build
		components := Components{
			Ctx:        GetLogger(t, "Router"),
			BrkStorage: NewMockBrkStorage(),
			GtwStorage: NewMockGtwStorage(),
		}
		r := New(components, Options{})
		req := &core.StatsReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.StatsRes)
		var wantEntry gtwEntry

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantEntry, components.GtwStorage.(*MockGtwStorage).InUpsert.Entry, "Gateway Entries")
	}

	// --------------------

	{
		Desc(t, "Handle Stats Request | Storage fails ")

		// Build
		components := Components{
			Ctx:        GetLogger(t, "Router"),
			BrkStorage: NewMockBrkStorage(),
			GtwStorage: NewMockGtwStorage(),
		}
		components.GtwStorage.(*MockGtwStorage).Failures["upsert"] = errors.New(errors.Operational, "Mock Error")
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
		var wantRes = new(core.StatsRes)
		var wantEntry = gtwEntry{
			GatewayID: req.GatewayID,
			Metadata:  *req.Metadata,
		}

		// Operate
		res, err := r.HandleStats(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Stats Responses")
		Check(t, wantEntry, components.GtwStorage.(*MockGtwStorage).InUpsert.Entry, "Gateway Entries")
	}
}

func TestHandleData(t *testing.T) {
	{
		Desc(t, "Handle invalid uplink | Invalid DevAddr")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4, 5},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata:  new(core.Metadata),
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataRouterRes)
		var wantBrReq *core.DataBrokerReq
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle invalid uplink | Invalid MIC")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3},
			},
			Metadata:  new(core.Metadata),
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataRouterRes)
		var wantBrReq *core.DataBrokerReq
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle invalid uplink | No Metadata")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata:  nil,
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataRouterRes)
		var wantBrReq *core.DataBrokerReq
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle invalid uplink | No Payload")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload:   nil,
			Metadata:  new(core.Metadata),
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataRouterRes)
		var wantBrReq *core.DataBrokerReq
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle invalid uplink | Invalid GatewayID")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata:  new(core.Metadata),
			GatewayID: []byte{1, 2, 3, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataRouterRes)
		var wantBrReq *core.DataBrokerReq
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker ok | no downlink")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.OutRead.Entries = []brkEntry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr *string
		var wantRes = new(core.DataRouterRes)
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 2 brokers unknown | no downlink")

		// Build
		dm := mocks.NewDutyManager()
		br1 := mocks.NewAuthBrokerClient()
		br1.Failures["HandleData"] = errors.New(errors.NotFound, "Mock Error")
		br2 := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.Failures["read"] = errors.New(errors.NotFound, "Mock Error")
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr *string
		var wantRes = new(core.DataRouterRes)
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16 = 1

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br1.InHandleData.Req, "Broker Data Requests")
		Check(t, wantBrReq, br2.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 2 brokers unknown | no downlink | Fail to store")

		// Build
		dm := mocks.NewDutyManager()
		br1 := mocks.NewAuthBrokerClient()
		br1.Failures["HandleData"] = errors.New(errors.NotFound, "Mock Error")
		br2 := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.Failures["read"] = errors.New(errors.NotFound, "Mock Error")
		st.Failures["Store"] = errors.New(errors.Operational, "Mock Error")
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr *string
		var wantRes = new(core.DataRouterRes)
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16 = 1

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br1.InHandleData.Req, "Broker Data Requests")
		Check(t, wantBrReq, br2.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | Fail Storage Lookup")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.Failures["read"] = errors.New(errors.Operational, "Mock Error")

		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes = new(core.DataRouterRes)
		var wantBrReq *core.DataBrokerReq
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | Fail DutyManager Lookup")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["read"] = errors.New(errors.NotFound, "Mock Error")
		br := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.OutRead.Entries = []brkEntry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr *string
		var wantRes = new(core.DataRouterRes)
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | Unreckognized frequency")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.OutRead.Entries = []brkEntry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 12.3,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataRouterRes)
		var wantBrReq *core.DataBrokerReq
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 2 brokers unknown | both errored")

		// Build
		dm := mocks.NewDutyManager()
		br1 := mocks.NewAuthBrokerClient()
		br1.Failures["HandleData"] = errors.New(errors.NotFound, "Mock Error")
		br2 := mocks.NewAuthBrokerClient()
		br2.Failures["HandleData"] = errors.New(errors.Operational, "Mock Error")
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.Failures["read"] = errors.New(errors.NotFound, "Mock Error")
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes = new(core.DataRouterRes)
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br1.InHandleData.Req, "Broker Data Requests")
		Check(t, wantBrReq, br2.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 2 brokers unknown | both respond positively")

		// Build
		dm := mocks.NewDutyManager()
		br1 := mocks.NewAuthBrokerClient()
		br2 := mocks.NewAuthBrokerClient()
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.Failures["read"] = errors.New(errors.NotFound, "Mock Error")
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrBehavioural
		var wantRes = new(core.DataRouterRes)
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br1.InHandleData.Req, "Broker Data Requests")
		Check(t, wantBrReq, br2.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker known, not ok | no downlink")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		br.Failures["HandleData"] = errors.New(errors.NotFound, "Mock Error")
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.OutRead.Entries = []brkEntry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrNotFound
		var wantRes = new(core.DataRouterRes)
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker known ok | valid downlink")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		br.OutHandleData.Res = &core.DataBrokerRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{5, 6, 7, 8},
						FCnt:    2,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      4,
					FRMPayload: []byte{42, 42, 14, 14},
				},
				MIC: []byte{8, 7, 6, 5},
			},
			Metadata: new(core.Metadata),
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.OutRead.Entries = []brkEntry{
			{
				BrokerIndex: 0,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr *string
		var wantRes = &core.DataRouterRes{
			Payload:  br.OutHandleData.Res.Payload,
			Metadata: br.OutHandleData.Res.Metadata,
		}
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16
		var wantUpdateGtw = req.GatewayID

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker known ok | invalid downlink | no metadata")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		br.OutHandleData.Res = &core.DataBrokerRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{5, 6, 7, 8},
						FCnt:    2,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      4,
					FRMPayload: []byte{42, 42, 14, 14},
				},
				MIC: []byte{8, 7, 6, 5},
			},
			Metadata: nil,
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.OutRead.Entries = []brkEntry{
			{
				BrokerIndex: 0,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataRouterRes)
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16
		var wantUpdateGtw []byte

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker known ok | valid downlink | fail update Duty")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["Update"] = errors.New(errors.Operational, "Mock Error")
		br := mocks.NewAuthBrokerClient()
		br.OutHandleData.Res = &core.DataBrokerRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{5, 6, 7, 8},
						FCnt:    2,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      4,
					FRMPayload: []byte{42, 42, 14, 14},
				},
				MIC: []byte{8, 7, 6, 5},
			},
			Metadata: new(core.Metadata),
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		st.OutRead.Entries = []brkEntry{
			{
				BrokerIndex: 0,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.DataRouterReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{4, 3, 2, 1},
			},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes = new(core.DataRouterRes)
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16
		var wantUpdateGtw = req.GatewayID

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}
}

func TestHandleJoin(t *testing.T) {
	{
		Desc(t, "Handle valid join request | valid join response")

		// Build
		dm := mocks.NewDutyManager()
		br1 := mocks.NewAuthBrokerClient()
		br1.Failures["HandleJoin"] = errors.New(errors.NotFound, "Mock Error")
		br2 := mocks.NewAuthBrokerClient()
		br2.OutHandleJoin.Res = &core.JoinBrokerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{1, 2, 3, 4},
			},
			Metadata: &core.Metadata{},
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.JoinRouterReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce:  []byte{3, 3},
			MIC:       []byte{14, 14, 14, 14},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
		}

		// Expect
		var wantErr *string
		var wantRes = &core.JoinRouterRes{
			Payload:  br2.OutHandleJoin.Res.Payload,
			Metadata: br2.OutHandleJoin.Res.Metadata,
		}
		var wantBrReq = &core.JoinBrokerReq{
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			DevNonce: req.DevNonce,
			MIC:      req.MIC,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16 = 1
		var wantUpdateGtw = req.GatewayID

		// Operate
		res, err := r.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Join Responses")
		Check(t, wantBrReq, br1.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantBrReq, br2.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle valid join request | invalid join response -> fails to handle down")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["Update"] = errors.New(errors.Operational, "Mock Error")
		br1 := mocks.NewAuthBrokerClient()
		br1.Failures["HandleJoin"] = errors.New(errors.NotFound, "Mock Error")
		br2 := mocks.NewAuthBrokerClient()
		br2.OutHandleJoin.Res = &core.JoinBrokerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{1, 2, 3, 4},
			},
			Metadata: &core.Metadata{},
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.JoinRouterReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce:  []byte{3, 3},
			MIC:       []byte{14, 14, 14, 14},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes = new(core.JoinRouterRes)
		var wantBrReq = &core.JoinBrokerReq{
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			DevNonce: req.DevNonce,
			MIC:      req.MIC,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16 = 1
		var wantUpdateGtw = req.GatewayID

		// Operate
		res, err := r.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Join Responses")
		Check(t, wantBrReq, br1.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantBrReq, br2.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle invalid join -> No Metadata")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["Update"] = errors.New(errors.Operational, "Mock Error")
		br := mocks.NewAuthBrokerClient()
		br.OutHandleJoin.Res = &core.JoinBrokerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{1, 2, 3, 4},
			},
			Metadata: new(core.Metadata),
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.JoinRouterReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce:  []byte{3, 3},
			MIC:       []byte{14, 14, 14, 14},
			Metadata:  nil,
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinRouterRes)
		var wantBrReq *core.JoinBrokerReq
		var wantStore uint16
		var wantUpdateGtw []byte

		// Operate
		res, err := r.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Join Responses")
		Check(t, wantBrReq, br.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle invalid Join Request -> Invalid DevEUI")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["Update"] = errors.New(errors.Operational, "Mock Error")
		br := mocks.NewAuthBrokerClient()
		br.OutHandleJoin.Res = &core.JoinBrokerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{1, 2, 3, 4},
			},
			Metadata: new(core.Metadata),
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.JoinRouterReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:    []byte{},
			DevNonce:  []byte{3, 3},
			MIC:       []byte{14, 14, 14, 14},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinRouterRes)
		var wantBrReq *core.JoinBrokerReq
		var wantStore uint16
		var wantUpdateGtw []byte

		// Operate
		res, err := r.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Join Responses")
		Check(t, wantBrReq, br.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle invalid Join Request -> Invalid AppEUI")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["Update"] = errors.New(errors.Operational, "Mock Error")
		br := mocks.NewAuthBrokerClient()
		br.OutHandleJoin.Res = &core.JoinBrokerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{1, 2, 3, 4},
			},
			Metadata: new(core.Metadata),
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.JoinRouterReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:    []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce:  []byte{3, 3},
			MIC:       []byte{14, 14, 14, 14},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinRouterRes)
		var wantBrReq *core.JoinBrokerReq
		var wantStore uint16
		var wantUpdateGtw []byte

		// Operate
		res, err := r.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Join Responses")
		Check(t, wantBrReq, br.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle invalid Join Request -> Invalid DevNonce")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["Update"] = errors.New(errors.Operational, "Mock Error")
		br := mocks.NewAuthBrokerClient()
		br.OutHandleJoin.Res = &core.JoinBrokerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{1, 2, 3, 4},
			},
			Metadata: new(core.Metadata),
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.JoinRouterReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce:  []byte{3},
			MIC:       []byte{14, 14, 14, 14},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinRouterRes)
		var wantBrReq *core.JoinBrokerReq
		var wantStore uint16
		var wantUpdateGtw []byte

		// Operate
		res, err := r.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Join Responses")
		Check(t, wantBrReq, br.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle invalid Join Request -> Invalid GatewayID")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["Update"] = errors.New(errors.Operational, "Mock Error")
		br := mocks.NewAuthBrokerClient()
		br.OutHandleJoin.Res = &core.JoinBrokerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{1, 2, 3, 4},
			},
			Metadata: new(core.Metadata),
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.JoinRouterReq{
			GatewayID: nil,
			AppEUI:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce:  []byte{3, 3},
			MIC:       []byte{14, 14, 14, 14},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinRouterRes)
		var wantBrReq *core.JoinBrokerReq
		var wantStore uint16
		var wantUpdateGtw []byte

		// Operate
		res, err := r.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Join Responses")
		Check(t, wantBrReq, br.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle valid join request | fails to send, no broker")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		br.Failures["HandleJoin"] = errors.New(errors.NotFound, "Mock Error")
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.JoinRouterReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce:  []byte{3, 3},
			MIC:       []byte{14, 14, 14, 14},
			Metadata: &core.Metadata{
				Frequency: 868.5,
			},
		}

		// Expect
		var wantErr = ErrNotFound
		var wantRes = new(core.JoinRouterRes)
		var wantBrReq = &core.JoinBrokerReq{
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			DevNonce: req.DevNonce,
			MIC:      req.MIC,
			Metadata: &core.Metadata{
				Altitude:  gt.OutRead.Entry.Metadata.Altitude,
				Longitude: gt.OutRead.Entry.Metadata.Longitude,
				Latitude:  gt.OutRead.Entry.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore uint16
		var wantUpdateGtw []byte

		// Operate
		res, err := r.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Join Responses")
		Check(t, wantBrReq, br.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle invalid join request -> bad frequency")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewAuthBrokerClient()
		br.OutHandleJoin.Res = &core.JoinBrokerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{1, 2, 3, 4},
			},
			Metadata: &core.Metadata{},
		}
		st := NewMockBrkStorage()
		gt := NewMockGtwStorage()

		gid := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		gt.OutRead.Entry = gtwEntry{
			GatewayID: gid,
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: 14.0,
				Latitude:  -14.0,
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			BrkStorage:  st,
			GtwStorage:  gt,
		}, Options{})
		req := &core.JoinRouterReq{
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:    []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce:  []byte{3, 3},
			MIC:       []byte{14, 14, 14, 14},
			Metadata: &core.Metadata{
				Frequency: 14.42,
			},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinRouterRes)
		var wantBrReq *core.JoinBrokerReq
		var wantStore uint16
		var wantUpdateGtw []byte

		// Operate
		res, err := r.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Join Responses")
		Check(t, wantBrReq, br.InHandleJoin.Req, "Broker Join Requests")
		Check(t, wantStore, st.InCreate.Entry.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

}

func TestStart(t *testing.T) {
	router := New(Components{
		Ctx:         GetLogger(t, "Router"),
		DutyManager: mocks.NewDutyManager(),
		Brokers:     []core.BrokerClient{mocks.NewAuthBrokerClient()},
		BrkStorage:  NewMockBrkStorage(),
		GtwStorage:  NewMockGtwStorage(),
	}, Options{NetAddr: "localhost:8886"})

	cherr := make(chan error)
	go func() {
		err := router.Start()
		cherr <- err
	}()

	var err error
	select {
	case err = <-cherr:
	case <-time.After(time.Millisecond * 250):
	}
	CheckErrors(t, nil, err)
}
