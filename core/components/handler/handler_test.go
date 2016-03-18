// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	"golang.org/x/net/context"
)

func TestHandleDataDown(t *testing.T) {
	{
		Desc(t, "Handle valid downlink")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.DataDownHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
			Payload: []byte("TheThingsNetwork"),
		}

		// Expect
		var wantError *string
		var wantRes *core.DataDownHandlerRes
		var wantEntry = pktEntry{Payload: req.Payload}

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataDown(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Data Down Handler Responses")
		Check(t, wantEntry, pktStorage.InPush.Payload, "Packet Entries")
	}

	// --------------------

	{
		Desc(t, "Handle invalid downlink ~> Invalid Payload")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.DataDownHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
			Payload: nil,
		}

		// Expect
		var wantError = ErrStructural
		var wantRes *core.DataDownHandlerRes
		var wantEntry pktEntry

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataDown(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Data Down Handler Responses")
		Check(t, wantEntry, pktStorage.InPush.Payload, "Packet Entries")
	}

	// --------------------

	{
		Desc(t, "Handle invalid downlink ~> Invalid AppEUI")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.DataDownHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1},
			DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
			Payload: []byte("TheThingsNetwork"),
		}

		// Expect
		var wantError = ErrStructural
		var wantRes *core.DataDownHandlerRes
		var wantEntry pktEntry

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataDown(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Data Down Handler Responses")
		Check(t, wantEntry, pktStorage.InPush.Payload, "Packet Entries")
	}

	// --------------------

	{
		Desc(t, "Handle invalid downlink ~> Invalid DevEUI")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.DataDownHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2, 2},
			Payload: []byte("TheThingsNetwork"),
		}

		// Expect
		var wantError = ErrStructural
		var wantRes *core.DataDownHandlerRes
		var wantEntry pktEntry

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataDown(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Data Down Handler Responses")
		Check(t, wantEntry, pktStorage.InPush.Payload, "Packet Entries")
	}
}

func TestHandleDataUp(t *testing.T) {
	{
		Desc(t, "Handle uplink, 1 packet | Unknown")

		// Build
		devStorage := NewMockDevStorage()
		devStorage.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error")
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  []byte("Payload"),
			Metadata: new(core.Metadata),
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:     14,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrNotFound
		var wantRes *core.DataUpHandlerRes
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, invalid Payload")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  nil,
			Metadata: new(core.Metadata),
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:     14,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes *core.DataUpHandlerRes
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, invalid Metadata")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  []byte("Payload"),
			Metadata: nil,
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:     14,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes *core.DataUpHandlerRes
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, invalid DevEUI")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  []byte("Payload"),
			Metadata: new(core.Metadata),
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   nil,
			FCnt:     14,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes *core.DataUpHandlerRes
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, invalid AppEUI")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  []byte("Payload"),
			Metadata: new(core.Metadata),
			AppEUI:   []byte{1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:     14,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes *core.DataUpHandlerRes
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | No downlink")

		// Build
		devStorage := NewMockDevStorage()
		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  [4]byte{3, 4, 2, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			true,
			devStorage.OutLookup.Entry.DevAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate: "SF7BW125",
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr *string
		var wantRes *core.DataUpHandlerRes
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
		}
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "2 packets in a row, same device | No Downlink")

		// Build
		devStorage := NewMockDevStorage()
		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  [4]byte{3, 4, 2, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			true,
			devStorage.OutLookup.Entry.DevAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req1 := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate: "SF7BW125",
				DutyRX1:  uint32(dutycycle.StateWarning),
				DutyRX2:  uint32(dutycycle.StateWarning),
				Rssi:     -20,
				Lsnr:     5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}
		req2 := &core.DataUpHandlerReq{
			Payload: req1.Payload,
			Metadata: &core.Metadata{
				DataRate: "SF7BW125",
				DutyRX1:  uint32(dutycycle.StateAvailable),
				DutyRX2:  uint32(dutycycle.StateAvailable),
				Rssi:     -20,
				Lsnr:     5.0,
			},
			AppEUI: req1.AppEUI,
			DevEUI: req1.DevEUI,
			FCnt:   req1.FCnt,
			MType:  req1.MType,
		}

		// Expect
		var wantErr1 *string
		var wantErr2 *string
		var wantRes1 *core.DataUpHandlerRes
		var wantRes2 *core.DataUpHandlerRes
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req1.Metadata, req2.Metadata},
			AppEUI:   req1.AppEUI,
			DevEUI:   req1.DevEUI,
		}
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})

		chack := make(chan bool)
		go func() {
			var ok bool
			defer func(ok *bool) { chack <- *ok }(&ok)
			res, err := handler.HandleDataUp(context.Background(), req1)

			// Check
			CheckErrors(t, wantErr1, err)
			Check(t, wantRes1, res, "Data Up Handler Responses")
			ok = true
		}()

		go func() {
			<-time.After(time.Millisecond * 50)
			var ok bool
			defer func(ok *bool) { chack <- *ok }(&ok)
			res, err := handler.HandleDataUp(context.Background(), req2)

			// Check
			CheckErrors(t, wantErr2, err)
			Check(t, wantRes2, res, "Data Up Handler Responses")
			ok = true
		}()

		// Check
		ok1, ok2 := <-chack, <-chack
		Check(t, true, ok1 && ok2, "Acknowledgements")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | one downlink ready")

		// Build
		tmst := time.Now()
		devStorage := NewMockDevStorage()
		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  [4]byte{3, 4, 2, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		pktStorage.OutPull.Entry.Payload = []byte("Downlink")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			true,
			devStorage.OutLookup.Entry.DevAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr *string
		encodedDown, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			false,
			devStorage.OutLookup.Entry.DevAddr,
			devStorage.OutLookup.Entry.FCntDown+1,
			pktStorage.OutPull.Entry.Payload,
		)
		FatalUnless(t, err)
		var wantRes = &core.DataUpHandlerRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: devStorage.OutLookup.Entry.DevAddr[:],
						FCnt:    devStorage.OutLookup.Entry.FCntDown + 1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      uint32(1),
					FRMPayload: encodedDown,
				},
				MIC: []byte{0, 0, 0, 0},
			},
			Metadata: &core.Metadata{
				DataRate:    "SF7BW125",
				Frequency:   865.5,
				CodingRate:  "4/5",
				Timestamp:   uint32(tmst.Add(time.Second).Unix() * 1000),
				PayloadSize: 21,
			},
		}
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
		}
		var wantFCnt = wantRes.Payload.MACPayload.FHDR.FCnt

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle late uplink | No Downlink")

		// Build
		devStorage := NewMockDevStorage()
		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  [4]byte{3, 4, 2, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			true,
			devStorage.OutLookup.Entry.DevAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req1 := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate: "SF7BW125",
				DutyRX1:  uint32(dutycycle.StateWarning),
				DutyRX2:  uint32(dutycycle.StateWarning),
				Rssi:     -20,
				Lsnr:     5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}
		req2 := &core.DataUpHandlerReq{
			Payload: req1.Payload,
			Metadata: &core.Metadata{
				DataRate: "SF7BW125",
				DutyRX1:  uint32(dutycycle.StateAvailable),
				DutyRX2:  uint32(dutycycle.StateAvailable),
				Rssi:     -20,
				Lsnr:     5.0,
			},
			AppEUI: req1.AppEUI,
			DevEUI: req1.DevEUI,
			FCnt:   req1.FCnt,
			MType:  req1.MType,
		}

		// Expect
		var wantErr1 *string
		var wantErr2 = ErrBehavioural
		var wantRes1 *core.DataUpHandlerRes
		var wantRes2 *core.DataUpHandlerRes
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req1.Metadata},
			AppEUI:   req1.AppEUI,
			DevEUI:   req1.DevEUI,
		}
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})

		chack := make(chan bool)
		go func() {
			var ok bool
			defer func(ok *bool) { chack <- *ok }(&ok)
			res, err := handler.HandleDataUp(context.Background(), req1)

			// Check
			CheckErrors(t, wantErr1, err)
			Check(t, wantRes1, res, "Data Up Handler Responses")
			ok = true
		}()

		go func() {
			<-time.After(2 * bufferDelay)
			var ok bool
			defer func(ok *bool) { chack <- *ok }(&ok)
			res, err := handler.HandleDataUp(context.Background(), req2)

			// Check
			CheckErrors(t, wantErr2, err)
			Check(t, wantRes2, res, "Data Up Handler Responses")
			ok = true
		}()

		// Check
		ok1, ok2 := <-chack, <-chack
		Check(t, true, ok1 && ok2, "Acknowledgements")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | AppAdapter fails")

		// Build
		devStorage := NewMockDevStorage()
		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  [4]byte{3, 4, 2, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		appAdapter.Failures["HandleData"] = fmt.Errorf("Mock Error")
		broker := mocks.NewBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			true,
			devStorage.OutLookup.Entry.DevAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate: "SF7BW125",
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes *core.DataUpHandlerRes
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
		}
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | PktStorage fails")

		// Build
		devStorage := NewMockDevStorage()
		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  [4]byte{3, 4, 2, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		pktStorage.Failures["Pull"] = errors.New(errors.Operational, "Mock Error")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			true,
			devStorage.OutLookup.Entry.DevAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate: "SF7BW125",
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes *core.DataUpHandlerRes
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
		}
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle two successive uplink | no downlink")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		payload1, fcnt1 := []byte("Payload1"), uint32(14)
		devAddr1, appSKey1 := [4]byte{1, 2, 3, 4}, [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
		payload2, fcnt2 := []byte("Payload2"), uint32(35346)
		devAddr2, appSKey2 := [4]byte{4, 3, 2, 1}, [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1}
		encoded1, err := lorawan.EncryptFRMPayload(
			appSKey1,
			true,
			devAddr1,
			fcnt1,
			payload1,
		)
		FatalUnless(t, err)
		encoded2, err := lorawan.EncryptFRMPayload(
			appSKey2,
			true,
			devAddr2,
			fcnt2,
			payload2,
		)
		FatalUnless(t, err)
		req1 := &core.DataUpHandlerReq{
			Payload: encoded1,
			Metadata: &core.Metadata{
				DataRate: "SF7BW125",
				DutyRX1:  uint32(dutycycle.StateWarning),
				DutyRX2:  uint32(dutycycle.StateWarning),
				Rssi:     -20,
				Lsnr:     5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt1,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}
		req2 := &core.DataUpHandlerReq{
			Payload: encoded2,
			Metadata: &core.Metadata{
				DataRate: "SF7BW125",
				DutyRX1:  uint32(dutycycle.StateAvailable),
				DutyRX2:  uint32(dutycycle.StateAvailable),
				Rssi:     -20,
				Lsnr:     5.0,
			},
			AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			FCnt:   fcnt2,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr1 *string
		var wantErr2 *string
		var wantRes1 *core.DataUpHandlerRes
		var wantRes2 *core.DataUpHandlerRes
		var wantData1 = &core.DataAppReq{
			Payload:  payload1,
			Metadata: []*core.Metadata{req1.Metadata},
			AppEUI:   req1.AppEUI,
			DevEUI:   req1.DevEUI,
		}
		var wantData2 = &core.DataAppReq{
			Payload:  payload2,
			Metadata: []*core.Metadata{req2.Metadata},
			AppEUI:   req2.AppEUI,
			DevEUI:   req2.DevEUI,
		}
		var wantFCnt1 uint32
		var wantFCnt2 uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})

		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  devAddr1,
			AppSKey:  appSKey1,
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		res1, err1 := handler.HandleDataUp(context.Background(), req1)

		// Check
		CheckErrors(t, wantErr1, err1)
		Check(t, wantRes1, res1, "Data Up Handler Responses")
		Check(t, wantData1, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt1, devStorage.InUpdateFCnt.FCnt, "Frame counters")

		// Operate
		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  devAddr2,
			AppSKey:  appSKey2,
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 11,
		}
		res2, err2 := handler.HandleDataUp(context.Background(), req2)

		// Check
		CheckErrors(t, wantErr2, err2)
		Check(t, wantRes2, res2, "Data Up Handler Responses")
		Check(t, wantData2, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt2, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | one downlink ready | Only RX2 available")

		// Build
		tmst := time.Now()
		devStorage := NewMockDevStorage()
		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  [4]byte{3, 4, 2, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		pktStorage.OutPull.Entry.Payload = []byte("Downlink")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			true,
			devStorage.OutLookup.Entry.DevAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateBlocked),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr *string
		encodedDown, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			false,
			devStorage.OutLookup.Entry.DevAddr,
			devStorage.OutLookup.Entry.FCntDown+1,
			pktStorage.OutPull.Entry.Payload,
		)
		FatalUnless(t, err)
		var wantRes = &core.DataUpHandlerRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: devStorage.OutLookup.Entry.DevAddr[:],
						FCnt:    devStorage.OutLookup.Entry.FCntDown + 1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      uint32(1),
					FRMPayload: encodedDown,
				},
				MIC: []byte{0, 0, 0, 0},
			},
			Metadata: &core.Metadata{
				DataRate:    "SF9BW125",
				Frequency:   869.5,
				CodingRate:  "4/5",
				Timestamp:   uint32(tmst.Add(2*time.Second).Unix() * 1000),
				PayloadSize: 21,
			},
		}
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
		}
		var wantFCnt = wantRes.Payload.MACPayload.FHDR.FCnt

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | one downlink ready | Update FCnt fails")

		// Build
		tmst := time.Now()
		devStorage := NewMockDevStorage()
		devStorage.OutLookup.Entry = devEntry{
			DevAddr:  [4]byte{3, 4, 2, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		devStorage.Failures["UpdateFCnt"] = errors.New(errors.Operational, "Mock Error")
		pktStorage := NewMockPktStorage()
		pktStorage.OutPull.Entry.Payload = []byte("Downlink")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutLookup.Entry.AppSKey,
			true,
			devStorage.OutLookup.Entry.DevAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateBlocked),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes *core.DataUpHandlerRes
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
		}
		var wantFCnt = devStorage.OutLookup.Entry.FCntDown + 1

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpdateFCnt.FCnt, "Frame counters")
	}
}

func TestSubscribePersonalized(t *testing.T) {
	{
		Desc(t, "Handle Valid Perso Subscription")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.ABPSubHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevAddr: []byte{2, 2, 2, 2},
			NwkSKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppSKey: []byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		}
		addr := "localhost"

		// Expect
		var wantError *string
		var wantRes *core.ABPSubHandlerRes
		var wantSub = req.AppEUI
		var wantReq = &core.ABPSubBrokerReq{
			HandlerNet: addr,
			AppEUI:     req.AppEUI,
			DevAddr:    req.DevAddr,
			NwkSKey:    req.NwkSKey,
		}

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: addr})
		res, err := handler.SubscribePersonalized(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Subscribe Handler Responses")
		Check(t, wantSub, devStorage.InStorePersonalized.AppEUI, "Subscriptions")
		Check(t, wantReq, broker.InSubscribePersonalized.Req, "Subscribe Broker Requests")
	}

	// ----------------------

	{
		Desc(t, "Handle Invalid Perso Subscription -> Invalid AppEUI")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.ABPSubHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1},
			DevAddr: []byte{2, 2, 2, 2},
			NwkSKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppSKey: []byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		}
		addr := "localhost"

		// Expect
		var wantError = ErrStructural
		var wantRes *core.ABPSubHandlerRes
		var wantSub []byte
		var wantReq *core.ABPSubBrokerReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: addr})
		res, err := handler.SubscribePersonalized(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Subscribe Handler Responses")
		Check(t, wantSub, devStorage.InStorePersonalized.AppEUI, "Subscriptions")
		Check(t, wantReq, broker.InSubscribePersonalized.Req, "Subscribe Broker Requests")
	}

	// ----------------------

	{
		Desc(t, "Handle Invalid Perso Subscription -> Invalid DevAddr")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.ABPSubHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevAddr: nil,
			NwkSKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppSKey: []byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		}
		addr := "localhost"

		// Expect
		var wantError = ErrStructural
		var wantRes *core.ABPSubHandlerRes
		var wantSub []byte
		var wantReq *core.ABPSubBrokerReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: addr})
		res, err := handler.SubscribePersonalized(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Subscribe Handler Responses")
		Check(t, wantSub, devStorage.InStorePersonalized.AppEUI, "Subscriptions")
		Check(t, wantReq, broker.InSubscribePersonalized.Req, "Subscribe Broker Requests")
	}

	// ----------------------

	{
		Desc(t, "Handle Invalid Perso Subscription -> Invalid NwkSKey")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.ABPSubHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevAddr: []byte{2, 2, 2, 2},
			NwkSKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 12, 14, 12},
			AppSKey: []byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		}
		addr := "localhost"

		// Expect
		var wantError = ErrStructural
		var wantRes *core.ABPSubHandlerRes
		var wantSub []byte
		var wantReq *core.ABPSubBrokerReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: addr})
		res, err := handler.SubscribePersonalized(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Subscribe Handler Responses")
		Check(t, wantSub, devStorage.InStorePersonalized.AppEUI, "Subscriptions")
		Check(t, wantReq, broker.InSubscribePersonalized.Req, "Subscribe Broker Requests")
	}

	// ----------------------

	{
		Desc(t, "Handle Invalid Perso Subscription -> Invalid AppSKey")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.ABPSubHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevAddr: []byte{2, 2, 2, 2},
			NwkSKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppSKey: []byte{6, 5, 4, 3, 7, 6, 5, 4, 3, 2, 1},
		}
		addr := "localhost"

		// Expect
		var wantError = ErrStructural
		var wantRes *core.ABPSubHandlerRes
		var wantSub []byte
		var wantReq *core.ABPSubBrokerReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: addr})
		res, err := handler.SubscribePersonalized(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Subscribe Handler Responses")
		Check(t, wantSub, devStorage.InStorePersonalized.AppEUI, "Subscriptions")
		Check(t, wantReq, broker.InSubscribePersonalized.Req, "Subscribe Broker Requests")
	}

	// ----------------------

	{
		Desc(t, "Handle Invalid Perso Subscription -> Storage fails")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		devStorage.Failures["StorePersonalized"] = errors.New(errors.Operational, "Mock Error")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		req := &core.ABPSubHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevAddr: []byte{2, 2, 2, 2},
			NwkSKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppSKey: []byte{6, 5, 4, 3, 7, 6, 5, 4, 3, 2, 1, 0, 1, 1, 1, 1},
		}
		addr := "localhost"

		// Expect
		var wantError = ErrOperational
		var wantRes *core.ABPSubHandlerRes
		var wantSub = req.AppEUI
		var wantReq *core.ABPSubBrokerReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: addr})
		res, err := handler.SubscribePersonalized(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Subscribe Handler Responses")
		Check(t, wantSub, devStorage.InStorePersonalized.AppEUI, "Subscriptions")
		Check(t, wantReq, broker.InSubscribePersonalized.Req, "Subscribe Broker Requests")
	}

	// ----------------------

	{
		Desc(t, "Handle Invalid Perso Subscription -> Brokers fails")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewBrokerClient()
		broker.Failures["SubscribePersonalized"] = fmt.Errorf("Mock Error")
		req := &core.ABPSubHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevAddr: []byte{2, 2, 2, 2},
			NwkSKey: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppSKey: []byte{6, 5, 4, 3, 7, 6, 5, 4, 3, 2, 1, 0, 1, 1, 1, 1},
		}
		addr := "localhost"

		// Expect
		var wantError = ErrOperational
		var wantRes *core.ABPSubHandlerRes
		var wantSub = req.AppEUI
		var wantReq = &core.ABPSubBrokerReq{
			HandlerNet: addr,
			AppEUI:     req.AppEUI,
			DevAddr:    req.DevAddr,
			NwkSKey:    req.NwkSKey,
		}

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{NetAddr: addr})
		res, err := handler.SubscribePersonalized(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Subscribe Handler Responses")
		Check(t, wantSub, devStorage.InStorePersonalized.AppEUI, "Subscriptions")
		Check(t, wantReq, broker.InSubscribePersonalized.Req, "Subscribe Broker Requests")
	}
}

func TestStart(t *testing.T) {
	handler := New(Components{
		Ctx:        GetLogger(t, "Handler"),
		DevStorage: NewMockDevStorage(),
		PktStorage: NewMockPktStorage(),
		AppAdapter: mocks.NewAppClient(),
	}, Options{NetAddr: "localhost:8888"})

	cherr := make(chan error)
	go func() {
		err := handler.Start()
		cherr <- err
	}()

	var err error
	select {
	case err = <-cherr:
	case <-time.After(time.Millisecond * 250):
	}
	CheckErrors(t, nil, err)
}
