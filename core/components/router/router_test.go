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
	{
		Desc(t, "Handle invalid uplink | Invalid DevAddr")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
		st := NewMockStorage()
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq *core.DataBrokerReq
		var wantStore int

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle invalid uplink | Invalid MIC")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
		st := NewMockStorage()
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq *core.DataBrokerReq
		var wantStore int

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle invalid uplink | No Metadata")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
		st := NewMockStorage()
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq *core.DataBrokerReq
		var wantStore int

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle invalid uplink | No Payload")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
		st := NewMockStorage()
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
		}, Options{})
		req := &core.DataRouterReq{
			Payload:   nil,
			Metadata:  new(core.Metadata),
			GatewayID: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes *core.DataRouterRes
		var wantBrReq *core.DataBrokerReq
		var wantStore int

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle invalid uplink | Invalid GatewayID")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
		st := NewMockStorage()
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq *core.DataBrokerReq
		var wantStore int

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker ok | no downlink")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.OutLookup.Entries = []entry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 0

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 2 brokers unknown | no downlink")

		// Build
		dm := mocks.NewDutyManager()
		br1 := mocks.NewBrokerClient()
		br1.Failures["HandleData"] = errors.New(errors.NotFound, "Mock Error")
		br2 := mocks.NewBrokerClient()
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error")
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 1

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br1.InHandleData.Req, "Broker Data Requests")
		Check(t, wantBrReq, br2.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 2 brokers unknown | no downlink | Fail to store")

		// Build
		dm := mocks.NewDutyManager()
		br1 := mocks.NewBrokerClient()
		br1.Failures["HandleData"] = errors.New(errors.NotFound, "Mock Error")
		br2 := mocks.NewBrokerClient()
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error")
		st.Failures["Store"] = errors.New(errors.Operational, "Mock Error")
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 1

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br1.InHandleData.Req, "Broker Data Requests")
		Check(t, wantBrReq, br2.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | Fail Storage Lookup")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.Failures["Lookup"] = errors.New(errors.Operational, "Mock Error")

		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq *core.DataBrokerReq
		var wantStore = 0

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | Fail DutyManager Lookup")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error")
		br := mocks.NewBrokerClient()
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.OutLookup.Entries = []entry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 0

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | Unreckognized frequency")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.OutLookup.Entries = []entry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq *core.DataBrokerReq
		var wantStore = 0

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 2 brokers unknown | both errored")

		// Build
		dm := mocks.NewDutyManager()
		br1 := mocks.NewBrokerClient()
		br1.Failures["HandleData"] = errors.New(errors.NotFound, "Mock Error")
		br2 := mocks.NewBrokerClient()
		br2.Failures["HandleData"] = errors.New(errors.Operational, "Mock Error")
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error")
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 0

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br1.InHandleData.Req, "Broker Data Requests")
		Check(t, wantBrReq, br2.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 2 brokers unknown | both respond positively")

		// Build
		dm := mocks.NewDutyManager()
		br1 := mocks.NewBrokerClient()
		br2 := mocks.NewBrokerClient()
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error")
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br1, br2},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 0

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br1.InHandleData.Req, "Broker Data Requests")
		Check(t, wantBrReq, br2.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker known, not ok | no downlink")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
		br.Failures["HandleData"] = errors.New(errors.NotFound, "Mock Error")
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.OutLookup.Entries = []entry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br, br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 0

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker known ok | valid downlink")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
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
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.OutLookup.Entries = []entry{
			{
				BrokerIndex: 0,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 0
		var wantUpdateGtw = req.GatewayID

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker known ok | invalid downlink | no metadata")

		// Build
		dm := mocks.NewDutyManager()
		br := mocks.NewBrokerClient()
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
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.OutLookup.Entries = []entry{
			{
				BrokerIndex: 0,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 0
		var wantUpdateGtw []byte

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

	// --------------------

	{
		Desc(t, "Handle valid uplink | 1 broker known ok | valid downlink | fail update Duty")

		// Build
		dm := mocks.NewDutyManager()
		dm.Failures["Update"] = errors.New(errors.Operational, "Mock Error")
		br := mocks.NewBrokerClient()
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
		st := NewMockStorage()
		st.OutLookupStats.Metadata = core.StatsMetadata{
			Altitude:  14,
			Longitude: 14.0,
			Latitude:  -14.0,
		}
		st.OutLookup.Entries = []entry{
			{
				BrokerIndex: 0,
				until:       time.Now().Add(time.Hour),
			},
		}
		r := New(Components{
			DutyManager: dm,
			Brokers:     []core.BrokerClient{br},
			Ctx:         GetLogger(t, "Router"),
			Storage:     st,
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
		var wantRes *core.DataRouterRes
		var wantBrReq = &core.DataBrokerReq{
			Payload: req.Payload,
			Metadata: &core.Metadata{
				Altitude:  st.OutLookupStats.Metadata.Altitude,
				Longitude: st.OutLookupStats.Metadata.Longitude,
				Latitude:  st.OutLookupStats.Metadata.Latitude,
				Frequency: req.Metadata.Frequency,
			},
		}
		var wantStore = 0
		var wantUpdateGtw = req.GatewayID

		// Operate
		res, err := r.HandleData(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Router Data Responses")
		Check(t, wantBrReq, br.InHandleData.Req, "Broker Data Requests")
		Check(t, wantStore, st.InStore.BrokerIndex, "Brokers stored")
		Check(t, wantUpdateGtw, dm.InUpdate.ID, "Gateway updated")
	}

}

func TestStart(t *testing.T) {
	router := New(Components{
		Ctx:         GetLogger(t, "Router"),
		DutyManager: mocks.NewDutyManager(),
		Brokers:     []core.BrokerClient{mocks.NewBrokerClient()},
		Storage:     NewMockStorage(),
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
