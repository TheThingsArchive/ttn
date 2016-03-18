// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"testing"

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
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
		req := &core.DataBrokerReq{
			Payload:  nil,
			Metadata: new(core.Metadata),
		}

		// Expect
		var wantErr = ErrStructural
		var wantDataUp *core.DataUpHandlerReq
		var wantRes *core.DataBrokerRes
		var wantFCnt uint32

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Fail to lookup device -> Operational")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		nc.Failures["LookupDevices"] = errors.New(errors.Operational, "Mock Error")
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		var wantRes *core.DataBrokerRes
		var wantFCnt uint32

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Fail to lookup device -> Not Found")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		nc.Failures["LookupDevices"] = errors.New(errors.NotFound, "Mock Error")
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		var wantRes *core.DataBrokerRes
		var wantFCnt uint32

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | Two db entries, second MIC valid")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		nc.OutWholeCounter.FCnt = 2

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		dl2 := NewMockDialer()
		dl2.OutDial.Client = hl
		dl2.OutDial.Closer = NewMockCloser()

		nc.OutLookupDevices.Entries = []devEntry{
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
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		err = payload.SetMIC(lorawan.AES128Key(nc.OutLookupDevices.Entries[1].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]

		// Expect
		var wantErr *string
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutLookupDevices.Entries[1].AppEUI,
			DevEUI:   nc.OutLookupDevices.Entries[1].DevEUI,
			FCnt:     req.Payload.MACPayload.FHDR.FCnt,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes *core.DataBrokerRes
		var wantFCnt uint32 = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, FCnt invalid")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		nc.Failures["WholeCounter"] = errors.New(errors.Structural, "Mock Error")

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutLookupDevices.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  1,
			},
		}
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		err = payload.SetMIC(lorawan.AES128Key(nc.OutLookupDevices.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]

		// Expect
		var wantErr = ErrNotFound
		var wantDataUp *core.DataUpHandlerReq
		var wantRes *core.DataBrokerRes
		var wantFCnt uint32
		var wantDialer bool

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, FCnt above 16-bits")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		nc.OutWholeCounter.FCnt = 112534

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutLookupDevices.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  112500,
			},
		}
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		err = payload.SetMIC(lorawan.AES128Key(nc.OutLookupDevices.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr *string
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutLookupDevices.Entries[0].AppEUI,
			DevEUI:   nc.OutLookupDevices.Entries[0].DevEUI,
			FCnt:     nc.OutWholeCounter.FCnt,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes *core.DataBrokerRes
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, Dial failed")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		nc.OutWholeCounter.FCnt = 14

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()
		dl.Failures["Dial"] = errors.New(errors.Operational, "Mock Error")

		nc.OutLookupDevices.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		err = payload.SetMIC(lorawan.AES128Key(nc.OutLookupDevices.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr = ErrOperational
		var wantDataUp *core.DataUpHandlerReq
		var wantRes *core.DataBrokerRes
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, HandleDataUp failed")

		// Build
		hl := mocks.NewHandlerClient()
		hl.Failures["HandleDataUp"] = fmt.Errorf("Mock Error")

		nc := NewMockNetworkController()
		nc.OutWholeCounter.FCnt = 14

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutLookupDevices.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		err = payload.SetMIC(lorawan.AES128Key(nc.OutLookupDevices.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr = ErrOperational
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutLookupDevices.Entries[0].AppEUI,
			DevEUI:   nc.OutLookupDevices.Entries[0].DevEUI,
			FCnt:     nc.OutWholeCounter.FCnt,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes *core.DataBrokerRes
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry | One valid downlink")

		// Build
		nc := NewMockNetworkController()
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

		nc.OutLookupDevices.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		err = payload.SetMIC(lorawan.AES128Key(nc.OutLookupDevices.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr *string
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutLookupDevices.Entries[0].AppEUI,
			DevEUI:   nc.OutLookupDevices.Entries[0].DevEUI,
			FCnt:     nc.OutWholeCounter.FCnt,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes = &core.DataBrokerRes{
			Payload:  hl.OutHandleDataUp.Res.Payload,
			Metadata: hl.OutHandleDataUp.Res.Metadata,
		}
		payloadDown, err := core.NewLoRaWANData(req.Payload, false)
		FatalUnless(t, err)
		err = payloadDown.SetMIC(lorawan.AES128Key(nc.OutLookupDevices.Entries[0].NwkSKey))
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
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry, UpdateFcnt failed")

		// Build
		hl := mocks.NewHandlerClient()
		nc := NewMockNetworkController()
		nc.OutWholeCounter.FCnt = 14
		nc.Failures["UpdateFCnt"] = errors.New(errors.Operational, "Mock Error")

		dl := NewMockDialer()
		dl.OutDial.Client = hl
		dl.OutDial.Closer = NewMockCloser()

		nc.OutLookupDevices.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		err = payload.SetMIC(lorawan.AES128Key(nc.OutLookupDevices.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr = ErrOperational
		var wantDataUp *core.DataUpHandlerReq
		var wantRes *core.DataBrokerRes
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer bool

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

	// --------------------

	{
		Desc(t, "Valid uplink | One entry | Invalid downlink")

		// Build
		nc := NewMockNetworkController()
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

		nc.OutLookupDevices.Entries = []devEntry{
			{
				Dialer:  dl,
				AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
				DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
				NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				FCntUp:  10,
			},
		}
		br := New(Components{NetworkController: nc, Ctx: GetLogger(t, "Broker")}, Options{})
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
		err = payload.SetMIC(lorawan.AES128Key(nc.OutLookupDevices.Entries[0].NwkSKey))
		FatalUnless(t, err)
		req.Payload.MIC = payload.MIC[:]
		req.Payload.MACPayload.FHDR.FCnt %= 65536

		// Expect
		var wantErr = ErrStructural
		var wantDataUp = &core.DataUpHandlerReq{
			Payload:  req.Payload.MACPayload.FRMPayload,
			AppEUI:   nc.OutLookupDevices.Entries[0].AppEUI,
			DevEUI:   nc.OutLookupDevices.Entries[0].DevEUI,
			FCnt:     nc.OutWholeCounter.FCnt,
			MType:    req.Payload.MHDR.MType,
			Metadata: req.Metadata,
		}
		var wantRes *core.DataBrokerRes
		var wantFCnt = nc.OutWholeCounter.FCnt
		var wantDialer = true

		// Operate
		res, err := br.HandleData(context.Background(), req)

		// Checks
		CheckErrors(t, wantErr, err)
		Check(t, wantDataUp, hl.InHandleDataUp.Req, "Handler Data Requests")
		Check(t, wantRes, res, "Broker Data Responses")
		Check(t, wantFCnt, nc.InUpdateFcnt.FCnt, "Frame counters")
		Check(t, wantDialer, dl.InDial.Called, "Dialer calls")
	}

}

// SubscribePersonalized(bctx context.Context, req *core.ABPSubBrokerReq) (*core.ABPSubBrokerRes, error)
