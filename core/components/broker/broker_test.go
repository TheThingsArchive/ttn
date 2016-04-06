// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	"golang.org/x/net/context"
)

func TestHandleData(t *testing.T) {
	{
		Desc(t, "Invalid LoRaWAN payload")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload:  nil,
			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrStructural
		var wantDataUp *core.DataUpHandlerReq
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt uint32

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Fail to lookup device -> Operational")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.Failures["read"] = errors.New(errors.Operational, "Mock Error")
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
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
			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrOperational
		var wantDataUp *core.DataUpHandlerReq
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt uint32

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Fail to lookup device -> Not Found")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.Failures["read"] = errors.New(errors.NotFound, "Mock Error")
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
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
			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrNotFound
		var wantDataUp *core.DataUpHandlerReq
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt uint32

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | Two db entries, second MIC valid")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.OutWholeCounter.FCnt = 2

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		dl2 := NewMockDialer()
		dl2.OutDial.Client = hl
		dl2.OutDial.Closer = NewMockCloser()

		nc.OutRead.Entries = []devEntry{
			{
				Dialer:  dl2,
				AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI:  []byte{8, 7, 6, 5, 4, 3, 2, 1},
				NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
				FCntUp:  1,
			},
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  1,
			},
		}
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    nc.OutWholeCounter.FCnt,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{0, 0, 0, 0}, // Temporary, computed below
			},
			Metadata: new(core.Metadata),
		}
		payload, err := core.NewLoRaWANData(req.Payload, true)
		FatalUnless(t, err)
		err = payload.SetMIC(lorawan.AES128Key(nc.OutRead.Entries[1].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]

		// Expect
		var wantErr *string
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutRead.Entries[1].AppEUI,
			DevEUI:   nc.OutRead.Entries[1].DevEUI,
			FCnt:     req.Payload.MACPayload.FHDR.FCnt,
			FPort:    1,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, FCnt invalid")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.Failures["wholeCounter"] = errors.New(errors.Structural, "Mock Error")

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutRead.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  1,
			},
		}
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    44567,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{0, 0, 0, 0}, // Temporary, computed below
			},
			Metadata: new(core.Metadata),
		}
		payload, err := core.NewLoRaWANData(req.Payload, true)
		FatalUnless(t, err)
		err = payload.SetMIC(lorawan.AES128Key(nc.OutRead.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]

		// Expect
		var wantErr = ErrNotFound
		var wantDataUp *core.DataUpHandlerReq
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt uint32
		var wantDialer bool

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, FCnt above 16-bits")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.OutWholeCounter.FCnt = 112534

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutRead.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  112500,
			},
		}
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    nc.OutWholeCounter.FCnt,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{0, 0, 0, 0}, // Temporary, computed below
			},
			Metadata: new(core.Metadata),
		}
		payload, err := core.NewLoRaWANData(req.Payload, true)
		FatalUnless(t, err)
		err = payload.SetMIC(lorawan.AES128Key(nc.OutRead.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr *string
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutRead.Entries[0].AppEUI,
			DevEUI:   nc.OutRead.Entries[0].DevEUI,
			FCnt:     nc.OutWholeCounter.FCnt,
			FPort:    1,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, Dial failed")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.OutWholeCounter.FCnt = 14

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()
		dl.Failures["Dial"] = errors.New(errors.Operational, "Mock Error")

		nc.OutRead.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    nc.OutWholeCounter.FCnt,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{0, 0, 0, 0}, // Temporary, computed below
			},
			Metadata: new(core.Metadata),
		}
		payload, err := core.NewLoRaWANData(req.Payload, true)
		FatalUnless(t, err)
		err = payload.SetMIC(lorawan.AES128Key(nc.OutRead.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr = ErrOperational
		var wantDataUp *core.DataUpHandlerReq
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, HandleDataUp failed")

		// Build
		hl := mocks.NewHandlerClient()
		hl.Failures["HandleDataUp"] = fmt.Errorf("Mock Error")

		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.OutWholeCounter.FCnt = 14

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutRead.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    nc.OutWholeCounter.FCnt,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{0, 0, 0, 0}, // Temporary, computed below
			},
			Metadata: new(core.Metadata),
		}
		payload, err := core.NewLoRaWANData(req.Payload, true)
		FatalUnless(t, err)
		err = payload.SetMIC(lorawan.AES128Key(nc.OutRead.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr = ErrOperational
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutRead.Entries[0].AppEUI,
			DevEUI:   nc.OutRead.Entries[0].DevEUI,
			FCnt:     nc.OutWholeCounter.FCnt,
			FPort:    1,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry | One valid downlink")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.OutWholeCounter.FCnt = 14

		hl := mocks.NewHandlerClient()
		hl.OutHandleDataUp.Res = &core.DataUpHandlerRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    nc.OutWholeCounter.FCnt,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{0, 0, 0, 0},
			},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutRead.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    nc.OutWholeCounter.FCnt,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{0, 0, 0, 0}, // Temporary, computed below
			},
			Metadata: new(core.Metadata),
		}
		payload, err := core.NewLoRaWANData(req.Payload, true)
		FatalUnless(t, err)
		err = payload.SetMIC(lorawan.AES128Key(nc.OutRead.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr *string
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutRead.Entries[0].AppEUI,
			DevEUI:   nc.OutRead.Entries[0].DevEUI,
			FCnt:     nc.OutWholeCounter.FCnt,
			FPort:    1,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes = &core.DataBrokerRes{
			Payload:  hl.OutHandleDataUp.Res.Payload,
			Metadata: hl.OutHandleDataUp.Res.Metadata,
		}
		payloadDown, err := core.NewLoRaWANData(req.Payload, false)
		FatalUnless(t, err)
		err = payloadDown.SetMIC(lorawan.AES128Key(nc.OutRead.Entries[0].NwkSKey))
		FatalUnless(t, err)
		wantRes.Payload.MIC = payloadDown.MIC[:]
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, UpdateFcnt failed")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.OutWholeCounter.FCnt = 14
		nc.Failures["upsert"] = errors.New(errors.Operational, "Mock Error")

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutRead.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    nc.OutWholeCounter.FCnt,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{0, 0, 0, 0}, // Temporary, computed below
			},
			Metadata: new(core.Metadata),
		}
		payload, err := core.NewLoRaWANData(req.Payload, true)
		FatalUnless(t, err)
		err = payload.SetMIC(lorawan.AES128Key(nc.OutRead.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr = ErrOperational
		var wantDataUp *core.DataUpHandlerReq
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer bool

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry | Invalid downlink")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.OutWholeCounter.FCnt = 14

		hl := mocks.NewHandlerClient()
		hl.OutHandleDataUp.Res = &core.DataUpHandlerRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: nil,
				MIC:        []byte{0, 0, 0, 0},
			},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutRead.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataUp),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: []byte{1, 2, 3, 4},
						FCnt:    nc.OutWholeCounter.FCnt,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      1,
					FRMPayload: []byte{14, 14, 42, 42},
				},
				MIC: []byte{0, 0, 0, 0}, // Temporary, computed below
			},
			Metadata: new(core.Metadata),
		}
		payload, err := core.NewLoRaWANData(req.Payload, true)
		FatalUnless(t, err)
		err = payload.SetMIC(lorawan.AES128Key(nc.OutRead.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr = ErrStructural
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutRead.Entries[0].AppEUI,
			DevEUI:   nc.OutRead.Entries[0].DevEUI,
			FCnt:     nc.OutWholeCounter.FCnt,
			FPort:    1,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes = new(core.DataBrokerRes)
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpsert.Entry.FCntUp, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}
}

func TestDialerCloser(t *testing.T) {
	{
		Desc(t, "Dial on a valid address, server is listening")

		// Build
		addr := "0.0.0.0:3300"
		conn, err := net.Listen("tcp", addr)
		FatalUnless(t, err)
		defer conn.Close()

		// Operate & Check
		dl := NewDialer([]byte(addr))
		_, cl, errDial := dl.Dial()
		CheckErrors(t, nil, errDial)
		errClose := cl.Close()
		CheckErrors(t, nil, errClose)
	}

	// --------------------

	{
		Desc(t, "Dial an invalid address")

		// Build & Operate & Check
		dl := NewDialer([]byte(""))
		_, _, errDial := dl.Dial()
		CheckErrors(t, ErrOperational, errDial)
	}
}

func TestHandleJoin(t *testing.T) {
	{
		Desc(t, "Valid Join Request | Valid Join Accept")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nc.OutReadNonces.Entry.AppEUI,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr *string
		var wantJoinReq = &core.JoinHandlerReq{
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			DevNonce: req.DevNonce,
			MIC:      req.MIC,
			Metadata: req.Metadata,
		}
		var wantRes = &core.JoinBrokerRes{
			Payload:  hl.OutHandleJoin.Res.Payload,
			Metadata: hl.OutHandleJoin.Res.Metadata,
		}
		var wantActivation = noncesEntry{
			AppEUI:    req.AppEUI,
			DevEUI:    req.DevEUI,
			DevNonces: [][]byte{req.DevNonce},
		}
		var wantDialer = true

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid Join Request - very first time, no devNonces | Valid Join Accept")

		// Build
		nc := NewMockNetworkController()
		nc.Failures["readNonces"] = errors.New(errors.NotFound, "Mock Error")
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:   []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr *string
		var wantJoinReq = &core.JoinHandlerReq{
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			DevNonce: req.DevNonce,
			MIC:      req.MIC,
			Metadata: req.Metadata,
		}
		var wantRes = &core.JoinBrokerRes{
			Payload:  hl.OutHandleJoin.Res.Payload,
			Metadata: hl.OutHandleJoin.Res.Metadata,
		}
		var wantActivation = noncesEntry{
			AppEUI:    req.AppEUI,
			DevEUI:    req.DevEUI,
			DevNonces: [][]byte{req.DevNonce},
		}
		var wantDialer = true

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Invalid Join Request -> Invalid AppEUI")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nil,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrStructural
		var wantJoinReq *core.JoinHandlerReq
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation noncesEntry
		var wantDialer bool

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Invalid Join Request -> Invalid DevEUI")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   as.OutRead.Entry.AppEUI,
			DevEUI:   []byte{1, 2, 3},
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrStructural
		var wantJoinReq *core.JoinHandlerReq
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation noncesEntry
		var wantDialer bool

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Invalid Join Request -> Invalid DevNonce")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nc.OutReadNonces.Entry.AppEUI,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14, 15, 16},
			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrStructural
		var wantJoinReq *core.JoinHandlerReq
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation noncesEntry
		var wantDialer bool

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Invalid Join Request -> Invalid Metadata")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nc.OutReadNonces.Entry.AppEUI,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: nil,
		}

		// Expect
		var wantErr = ErrStructural
		var wantJoinReq *core.JoinHandlerReq
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation noncesEntry
		var wantDialer bool

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid Join Request | ReadNonces failed")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.Failures["readNonces"] = errors.New(errors.Operational, "Mock Error")
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrOperational
		var wantJoinReq *core.JoinHandlerReq
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation noncesEntry
		var wantDialer bool

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid Join Request | DevNonce already exists")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{[]byte{14, 14}},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nc.OutReadNonces.Entry.AppEUI,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrStructural
		var wantJoinReq *core.JoinHandlerReq
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation noncesEntry
		var wantDialer bool

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid Join Request | Dial fails")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()
		dl.Failures["Dial"] = errors.New(errors.Operational, "Mock Error")

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nc.OutReadNonces.Entry.AppEUI,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrOperational
		var wantJoinReq *core.JoinHandlerReq
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation noncesEntry
		var wantDialer = true

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}
	// --------------------

	{
		Desc(t, "Valid Join Request | Handle join fails")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}
		hl.Failures["HandleJoin"] = fmt.Errorf("Mock Error")

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nc.OutReadNonces.Entry.AppEUI,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrOperational
		var wantJoinReq = &core.JoinHandlerReq{
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			DevNonce: req.DevNonce,
			MIC:      req.MIC,
			Metadata: req.Metadata,
		}
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation noncesEntry
		var wantDialer = true

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid Join Request | Invalid response from handler")

		// Build
		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  nil,
			NwkSKey:  nil,
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nc.OutReadNonces.Entry.AppEUI,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrOperational
		var wantJoinReq = &core.JoinHandlerReq{
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			DevNonce: req.DevNonce,
			MIC:      req.MIC,
			Metadata: req.Metadata,
		}
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation noncesEntry
		var wantDialer = true

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid Join Request | Update Activation fails")

		// Build
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.Failures["upsertNonces"] = errors.New(errors.Operational, "Mock Error")
		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nc.OutReadNonces.Entry.AppEUI,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrOperational
		var wantJoinReq = &core.JoinHandlerReq{
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			DevNonce: req.DevNonce,
			MIC:      req.MIC,
			Metadata: req.Metadata,
		}
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation = noncesEntry{
			AppEUI:    req.AppEUI,
			DevEUI:    req.DevEUI,
			DevNonces: [][]byte{req.DevNonce},
		}
		var wantDialer = true

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid Join Request | Store device fails")

		// Build
		hl := mocks.NewHandlerClient()
		hl.OutHandleJoin.Res = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{
				Payload: []byte{14, 42},
			},
			DevAddr:  []byte{1, 1, 1, 1},
			NwkSKey:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			Metadata: new(core.Metadata),
		}

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc := NewMockNetworkController()
		as := NewMockAppStorage()
		nc.Failures["upsert"] = errors.New(errors.Operational, "Mock Error")
		as.OutRead.Entry = appEntry{
			Dialer: dl,
			AppEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
		}
		nc.OutReadNonces.Entry = noncesEntry{
			AppEUI:    []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevEUI:    []byte{3, 3, 3, 3, 3, 3, 3, 3},
			DevNonces: [][]byte{},
		}

		br := New(Components{NetworkController: nc, AppStorage: as, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.JoinBrokerReq{
			AppEUI:   nc.OutReadNonces.Entry.AppEUI,
			DevEUI:   nc.OutReadNonces.Entry.DevEUI,
			DevNonce: []byte{14, 14},
			MIC:      []byte{14, 14, 14, 14},

			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrOperational
		var wantJoinReq = &core.JoinHandlerReq{
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			DevNonce: req.DevNonce,
			MIC:      req.MIC,
			Metadata: req.Metadata,
		}
		var wantRes = new(core.JoinBrokerRes)
		var wantActivation = noncesEntry{
			AppEUI:    req.AppEUI,
			DevEUI:    req.DevEUI,
			DevNonces: [][]byte{req.DevNonce},
		}
		var wantDialer = true

		// Operate
		res, err := br.HandleJoin(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantJoinReq, hl.InHandleJoin.Req, "Handler Join Requests")
		Check(t, wantRes, res, "Broker Join Responses")
		Check(t, wantActivation, nc.InUpsertNonces.Entry, "Activations")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}
}

func TestStart(t *testing.T) {
	broker := New(
		Components{
			Ctx:               GetLogger(t, "Broker"),
			NetworkController: NewMockNetworkController(),
		},
		Options{
			NetAddrUp:   "localhost:8883",
			NetAddrDown: "localhost:8884",
		},
	)

	cherr := make(chan error)
	go func() {
		err := broker.Start()
		cherr <- err
	}()

	var err error
	select {
	case err = <-cherr:
	case <-time.After(time.Millisecond * 250):
	}
	CheckErrors(t, nil, err)
}
