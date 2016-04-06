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
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataDownHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
			Payload: []byte("TheThingsNetwork"),
			TTL:     "2h",
		}

		// Expect
		var wantError *string
		var wantRes = new(core.DataDownHandlerRes)
		var wantEntry = pktEntry{Payload: req.Payload}

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataDown(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Data Down Handler Responses")
		Check(t, wantEntry.Payload, pktStorage.InEnqueue.Entry.Payload, "Packet Entries")
	}

	// --------------------

	{
		Desc(t, "Handle invalid downlink ~> Invalid Payload")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataDownHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
			TTL:     "2h",
			Payload: nil,
		}

		// Expect
		var wantError = ErrStructural
		var wantRes = new(core.DataDownHandlerRes)
		var wantEntry pktEntry

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataDown(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Data Down Handler Responses")
		Check(t, wantEntry, pktStorage.InEnqueue.Entry, "Packet Entries")
	}

	// --------------------

	{
		Desc(t, "Handle invalid downlink ~> Invalid AppEUI")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataDownHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1},
			DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2},
			TTL:     "2h",
			Payload: []byte("TheThingsNetwork"),
		}

		// Expect
		var wantError = ErrStructural
		var wantRes = new(core.DataDownHandlerRes)
		var wantEntry pktEntry

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataDown(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Data Down Handler Responses")
		Check(t, wantEntry, pktStorage.InEnqueue.Entry, "Packet Entries")
	}

	// --------------------

	{
		Desc(t, "Handle invalid downlink ~> Invalid DevEUI")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataDownHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2, 2},
			TTL:     "2h",
			Payload: []byte("TheThingsNetwork"),
		}

		// Expect
		var wantError = ErrStructural
		var wantRes = new(core.DataDownHandlerRes)
		var wantEntry pktEntry

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataDown(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Data Down Handler Responses")
		Check(t, wantEntry, pktStorage.InEnqueue.Entry, "Packet Entries")
	}

	// --------------------

	{
		Desc(t, "Handle invalid downlink ~> Invalid TTL")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataDownHandlerReq{
			AppEUI:  []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:  []byte{2, 2, 2, 2, 2, 2, 2, 2, 2},
			TTL:     "0s",
			Payload: []byte("TheThingsNetwork"),
		}

		// Expect
		var wantError = ErrStructural
		var wantRes = new(core.DataDownHandlerRes)
		var wantEntry pktEntry

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataDown(context.Background(), req)

		// Check
		CheckErrors(t, wantError, err)
		Check(t, wantRes, res, "Data Down Handler Responses")
		Check(t, wantEntry, pktStorage.InEnqueue.Entry, "Packet Entries")
	}
}

func TestHandleDataUp(t *testing.T) {
	{
		Desc(t, "Handle uplink, 1 packet | Unknown")

		// Build
		devStorage := NewMockDevStorage()
		devStorage.Failures["read"] = errors.New(errors.NotFound, "Mock Error")
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  []byte("Payload"),
			Metadata: new(core.Metadata),
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:     14,
			FPort:    1,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrNotFound
		var wantRes = new(core.DataUpHandlerRes)
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, invalid Payload")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  nil,
			Metadata: new(core.Metadata),
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:     14,
			FPort:    1,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataUpHandlerRes)
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, invalid Metadata")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  []byte("Payload"),
			Metadata: nil,
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:     14,
			FPort:    1,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataUpHandlerRes)
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, invalid DevEUI")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  []byte("Payload"),
			Metadata: new(core.Metadata),
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   nil,
			FCnt:     14,
			FPort:    1,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataUpHandlerRes)
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, invalid AppEUI")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		req := &core.DataUpHandlerReq{
			Payload:  []byte("Payload"),
			Metadata: new(core.Metadata),
			AppEUI:   []byte{1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:     14,
			FPort:    1,
			MType:    uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.DataUpHandlerRes)
		var wantData *core.DataAppReq
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | No downlink")

		// Build
		devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr[:],
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			true,
			devAddr,
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
			FPort:  1,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr *string
		var wantRes = new(core.DataUpHandlerRes)
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			FPort:    1,
			FCnt:     14,
		}
		var wantFCnt = devStorage.OutRead.Entry.FCntDown

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "2 packets in a row, same device | No Downlink")

		// Build
		devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr[:],
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			true,
			devAddr,
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
			FPort:  10,
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
			FPort:  req1.FPort,
			MType:  req1.MType,
		}

		// Expect
		var wantErr1 *string
		var wantErr2 *string
		var wantRes1 = new(core.DataUpHandlerRes)
		var wantRes2 = new(core.DataUpHandlerRes)
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req1.Metadata, req2.Metadata},
			AppEUI:   req1.AppEUI,
			DevEUI:   req1.DevEUI,
			FPort:    10,
			FCnt:     14,
		}
		var wantFCnt = devStorage.OutRead.Entry.FCntDown

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})

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
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | one downlink ready")

		// Build
		tmst := time.Now()
		devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr[:],
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		pktStorage.OutDequeue.Entry.Payload = []byte("Downlink")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			true,
			devAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			FPort:  1,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr *string
		encodedDown, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			false,
			devAddr,
			devStorage.OutRead.Entry.FCntDown+1,
			pktStorage.OutDequeue.Entry.Payload,
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
						DevAddr: devStorage.OutRead.Entry.DevAddr[:],
						FCnt:    devStorage.OutRead.Entry.FCntDown + 1,
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
				Timestamp:   uint32(tmst.Add(time.Second).Unix() * 1000000),
				PayloadSize: 21,
				Power:       14,
				InvPolarity: true,
			},
		}
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			FPort:    1,
			FCnt:     14,
		}
		var wantFCnt = wantRes.Payload.MACPayload.FHDR.FCnt

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	// See Issue #87
	// {
	// 	Desc(t, "Handle late uplink | No Downlink")
	//
	// 	// Build
	// 	devStorage := NewMockDevStorage()
	// 	devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
	// 	devStorage.OutRead.Entry = devEntry{
	// 		DevAddr:  devAddr[:],
	// 		AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
	// 		NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	// 		FCntDown: 3,
	// 	}
	// 	pktStorage := NewMockPktStorage()
	// 	appAdapter := mocks.NewAppClient()
	// 	broker := mocks.NewAuthBrokerClient()
	// 	payload, fcnt := []byte("Payload"), uint32(14)
	// 	encoded, err := lorawan.EncryptFRMPayload(
	// 		devStorage.OutRead.Entry.AppSKey,
	// 		true,
	// 		devAddr,
	// 		fcnt,
	// 		payload,
	// 	)
	// 	FatalUnless(t, err)
	// 	req1 := &core.DataUpHandlerReq{
	// 		Payload: encoded,
	// 		Metadata: &core.Metadata{
	// 			DataRate: "SF7BW125",
	// 			DutyRX1:  uint32(dutycycle.StateWarning),
	// 			DutyRX2:  uint32(dutycycle.StateWarning),
	// 			Rssi:     -20,
	// 			Lsnr:     5.0,
	// 		},
	// 		AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
	// 		DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
	// 		FCnt:   fcnt,
	// 		MType:  uint32(lorawan.UnconfirmedDataUp),
	// 	}
	// 	req2 := &core.DataUpHandlerReq{
	// 		Payload: req1.Payload,
	// 		Metadata: &core.Metadata{
	// 			DataRate: "SF7BW125",
	// 			DutyRX1:  uint32(dutycycle.StateAvailable),
	// 			DutyRX2:  uint32(dutycycle.StateAvailable),
	// 			Rssi:     -20,
	// 			Lsnr:     5.0,
	// 		},
	// 		AppEUI: req1.AppEUI,
	// 		DevEUI: req1.DevEUI,
	// 		FCnt:   req1.FCnt,
	// 		MType:  req1.MType,
	// 	}
	//
	// 	// Expect
	// 	var wantErr1 *string
	// 	var wantErr2 = ErrBehavioural
	// 	var wantRes1 = new(core.DataUpHandlerRes)
	// 	var wantRes2 = new(core.DataUpHandlerRes)
	// 	var wantData = &core.DataAppReq{
	// 		Payload:  payload,
	// 		Metadata: []*core.Metadata{req1.Metadata},
	// 		AppEUI:   req1.AppEUI,
	// 		DevEUI:   req1.DevEUI,
	// 	}
	// 	var wantFCnt = devStorage.OutRead.Entry.FCntDown
	//
	// 	// Operate
	// 	handler := New(Components{
	// 		Ctx:        GetLogger(t, "Handler"),
	// 		Broker:     broker,
	// 		AppAdapter: appAdapter,
	// 		DevStorage: devStorage,
	// 		PktStorage: pktStorage,
	// 	}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
	//
	// 	chack := make(chan bool)
	// 	go func() {
	// 		var ok bool
	// 		defer func(ok *bool) { chack <- *ok }(&ok)
	// 		res, err := handler.HandleDataUp(context.Background(), req1)
	//
	// 		// Check
	// 		CheckErrors(t, wantErr1, err)
	// 		Check(t, wantRes1, res, "Data Up Handler Responses")
	// 		ok = true
	// 	}()
	//
	// 	go func() {
	// 		<-time.After(2 * bufferDelay)
	// 		var ok bool
	// 		defer func(ok *bool) { chack <- *ok }(&ok)
	// 		res, err := handler.HandleDataUp(context.Background(), req2)
	//
	// 		// Check
	// 		CheckErrors(t, wantErr2, err)
	// 		Check(t, wantRes2, res, "Data Up Handler Responses")
	// 		ok = true
	// 	}()
	//
	// 	// Check
	// 	ok1, ok2 := <-chack, <-chack
	// 	Check(t, true, ok1 && ok2, "Acknowledgements")
	// 	Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
	// 	Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	// }

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | AppAdapter fails")

		// Build
		devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr[:],
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		appAdapter.Failures["HandleData"] = fmt.Errorf("Mock Error")
		broker := mocks.NewAuthBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			true,
			devAddr,
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
			FPort:  1,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes = new(core.DataUpHandlerRes)
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			FPort:    1,
			FCnt:     14,
		}
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | PktStorage fails")

		// Build
		devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr[:],
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		pktStorage.Failures["dequeue"] = errors.New(errors.Operational, "Mock Error")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			true,
			devAddr,
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
			FPort:  1,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes = new(core.DataUpHandlerRes)
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			FPort:    1,
			FCnt:     14,
		}
		var wantFCnt uint32

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle two successive uplink | no downlink")

		// Build
		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		payload1, fcnt1 := []byte("Payload1"), uint32(14)
		devAddr1, appSKey1 := lorawan.DevAddr([4]byte{1, 2, 3, 4}), [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
		payload2, fcnt2 := []byte("Payload2"), uint32(35346)
		devAddr2, appSKey2 := lorawan.DevAddr([4]byte{4, 3, 2, 1}), [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1}
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
			FPort:  1,
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
			FPort:  1,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr1 *string
		var wantErr2 *string
		var wantRes1 = new(core.DataUpHandlerRes)
		var wantRes2 = new(core.DataUpHandlerRes)
		var wantData1 = &core.DataAppReq{
			Payload:  payload1,
			Metadata: []*core.Metadata{req1.Metadata},
			AppEUI:   req1.AppEUI,
			DevEUI:   req1.DevEUI,
			FPort:    1,
			FCnt:     14,
		}
		var wantData2 = &core.DataAppReq{
			Payload:  payload2,
			Metadata: []*core.Metadata{req2.Metadata},
			AppEUI:   req2.AppEUI,
			DevEUI:   req2.DevEUI,
			FPort:    1,
			FCnt:     35346,
		}
		var wantFCnt1 uint32 = 3
		var wantFCnt2 uint32 = 11

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})

		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr1[:],
			AppSKey:  appSKey1,
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		res1, err1 := handler.HandleDataUp(context.Background(), req1)

		// Check
		CheckErrors(t, wantErr1, err1)
		Check(t, wantRes1, res1, "Data Up Handler Responses")
		Check(t, wantData1, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt1, devStorage.InUpsert.Entry.FCntDown, "Frame counters")

		// Operate
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr2[:],
			AppSKey:  appSKey2,
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 11,
		}
		res2, err2 := handler.HandleDataUp(context.Background(), req2)

		// Check
		CheckErrors(t, wantErr2, err2)
		Check(t, wantRes2, res2, "Data Up Handler Responses")
		Check(t, wantData2, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt2, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | one downlink ready | Only RX2 available")

		// Build
		tmst := time.Now()
		devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr[:],
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		pktStorage.OutDequeue.Entry.Payload = []byte("Downlink")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			true,
			devAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateBlocked),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			FPort:  1,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr *string
		encodedDown, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			false,
			devAddr,
			devStorage.OutRead.Entry.FCntDown+1,
			pktStorage.OutDequeue.Entry.Payload,
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
						DevAddr: devStorage.OutRead.Entry.DevAddr[:],
						FCnt:    devStorage.OutRead.Entry.FCntDown + 1,
						FCtrl:   new(core.LoRaWANFCtrl),
					},
					FPort:      uint32(1),
					FRMPayload: encodedDown,
				},
				MIC: []byte{0, 0, 0, 0},
			},
			Metadata: &core.Metadata{
				DataRate:    "SF9BW125",
				Frequency:   869.525,
				CodingRate:  "4/5",
				Timestamp:   uint32(tmst.Add(2*time.Second).Unix() * 1000000),
				PayloadSize: 21,
				Power:       27,
				InvPolarity: true,
			},
		}
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			FPort:    1,
			FCnt:     14,
		}
		var wantFCnt = wantRes.Payload.MACPayload.FHDR.FCnt

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle uplink, 1 packet | one downlink ready | Update FCnt fails")

		// Build
		tmst := time.Now()
		devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr[:],
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		devStorage.Failures["upsert"] = errors.New(errors.Operational, "Mock Error")
		pktStorage := NewMockPktStorage()
		pktStorage.OutDequeue.Entry.Payload = []byte("Downlink")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			true,
			devAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateBlocked),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			FPort:  1,
			MType:  uint32(lorawan.UnconfirmedDataUp),
		}

		// Expect
		var wantErr = ErrOperational
		var wantRes = new(core.DataUpHandlerRes)
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			FPort:    1,
			FCnt:     14,
		}
		var wantFCnt = devStorage.OutRead.Entry.FCntDown + 1

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle confirmed uplink, 1 packet | one downlink ready")

		// Build
		tmst := time.Now()
		devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr[:],
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		pktStorage.OutDequeue.Entry.Payload = []byte("Downlink")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			true,
			devAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			FPort:  1,
			MType:  uint32(lorawan.ConfirmedDataUp),
		}

		// Expect
		var wantErr *string
		encodedDown, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			false,
			devAddr,
			devStorage.OutRead.Entry.FCntDown+1,
			pktStorage.OutDequeue.Entry.Payload,
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
						DevAddr: devStorage.OutRead.Entry.DevAddr[:],
						FCnt:    devStorage.OutRead.Entry.FCntDown + 1,
						FCtrl: &core.LoRaWANFCtrl{
							Ack: true,
						},
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
				Timestamp:   uint32(tmst.Add(time.Second).Unix() * 1000000),
				PayloadSize: 21,
				Power:       14,
				InvPolarity: true,
			},
		}
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			FPort:    1,
			FCnt:     14,
		}
		var wantFCnt = wantRes.Payload.MACPayload.FHDR.FCnt

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

	// --------------------

	{
		Desc(t, "Handle confirmed uplink, 1 packet | no downlink")

		// Build
		tmst := time.Now()
		devAddr := lorawan.DevAddr([4]byte{3, 4, 2, 4})
		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			DevAddr:  devAddr[:],
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		pktStorage := NewMockPktStorage()
		pktStorage.Failures["dequeue"] = errors.New(errors.NotFound, "Mock Error")
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()
		payload, fcnt := []byte("Payload"), uint32(14)
		encoded, err := lorawan.EncryptFRMPayload(
			devStorage.OutRead.Entry.AppSKey,
			true,
			devAddr,
			fcnt,
			payload,
		)
		FatalUnless(t, err)
		req := &core.DataUpHandlerReq{
			Payload: encoded,
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
			AppEUI: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI: []byte{2, 2, 2, 2, 2, 2, 2, 2},
			FCnt:   fcnt,
			FPort:  1,
			MType:  uint32(lorawan.ConfirmedDataUp),
		}

		// Expect
		var wantErr *string
		var wantRes = &core.DataUpHandlerRes{
			Payload: &core.LoRaWANData{
				MHDR: &core.LoRaWANMHDR{
					MType: uint32(lorawan.UnconfirmedDataDown),
					Major: uint32(lorawan.LoRaWANR1),
				},
				MACPayload: &core.LoRaWANMACPayload{
					FHDR: &core.LoRaWANFHDR{
						DevAddr: devStorage.OutRead.Entry.DevAddr[:],
						FCnt:    devStorage.OutRead.Entry.FCntDown + 1,
						FCtrl: &core.LoRaWANFCtrl{
							Ack: true,
						},
					},
					FPort:      uint32(1),
					FRMPayload: nil,
				},
				MIC: []byte{0, 0, 0, 0},
			},
			Metadata: &core.Metadata{
				DataRate:    "SF7BW125",
				Frequency:   865.5,
				CodingRate:  "4/5",
				Timestamp:   uint32(tmst.Add(time.Second).Unix() * 1000000),
				PayloadSize: 13,
				Power:       14,
				InvPolarity: true,
			},
		}
		var wantData = &core.DataAppReq{
			Payload:  payload,
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
			FPort:    1,
			FCnt:     14,
		}
		var wantFCnt = wantRes.Payload.MACPayload.FHDR.FCnt

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleDataUp(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Data Up Handler Responses")
		Check(t, wantData, appAdapter.InHandleData.Req, "Data Application Requests")
		Check(t, wantFCnt, devStorage.InUpsert.Entry.FCntDown, "Frame counters")
	}

}

func TestHandleJoin(t *testing.T) {
	{
		Desc(t, "Handle valid join-request | get join-accept")

		// Build
		tmst := time.Now()

		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			AppKey: &[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppEUI: req.AppEUI,
			DevEUI: req.DevEUI,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		payload := lorawan.NewPHYPayload(true)
		payload.MHDR = lorawan.MHDR{MType: lorawan.JoinRequest, Major: lorawan.LoRaWANR1}
		joinPayload := lorawan.JoinRequestPayload{}
		copy(joinPayload.AppEUI[:], req.AppEUI)
		copy(joinPayload.DevEUI[:], req.DevEUI)
		copy(joinPayload.DevNonce[:], req.DevNonce)
		payload.MACPayload = &joinPayload
		err := payload.SetMIC(lorawan.AES128Key(*devStorage.OutRead.Entry.AppKey))
		FatalUnless(t, err)
		req.MIC = payload.MIC[:]

		// Expect
		var wantErr *string
		var wantRes = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{}, // We'll check it by decoding
			NwkSKey: nil,                       // We'll assume it's correct if payload is okay
			Metadata: &core.Metadata{
				DataRate:    "SF7BW125",
				Frequency:   865.5,
				CodingRate:  "4/5",
				Timestamp:   uint32(tmst.Add(5*time.Second).Unix() * 1000000),
				PayloadSize: 33,
				Power:       14,
				InvPolarity: true,
			},
		}
		var wantAppReq = &core.JoinAppReq{
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
		}

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes.Metadata, res.Metadata, "Join Handler Responses")
		Check(t, 16, len(res.NwkSKey), "Network session keys' length")
		Check(t, 4, len(res.DevAddr), "Device addresses' length")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
		joinaccept := lorawan.NewPHYPayload(false)
		err = joinaccept.UnmarshalBinary(res.Payload.Payload)
		CheckErrors(t, nil, err)
		err = joinaccept.DecryptJoinAcceptPayload(lorawan.AES128Key(*devStorage.InUpsert.Entry.AppKey))
		CheckErrors(t, nil, err)
		Check(t, handler.(*component).Configuration.NetID, joinaccept.MACPayload.(*lorawan.JoinAcceptPayload).NetID, "Network IDs")
	}

	// --------------------

	{
		Desc(t, "Handle valid join-request, fails to notify app.")

		// Build
		tmst := time.Now()

		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			AppKey: &[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppEUI: req.AppEUI,
			DevEUI: req.DevEUI,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		appAdapter.Failures["HandleJoin"] = errors.New(errors.Operational, "Mock Error")
		broker := mocks.NewAuthBrokerClient()

		payload := lorawan.NewPHYPayload(true)
		payload.MHDR = lorawan.MHDR{MType: lorawan.JoinRequest, Major: lorawan.LoRaWANR1}
		joinPayload := lorawan.JoinRequestPayload{}
		copy(joinPayload.AppEUI[:], req.AppEUI)
		copy(joinPayload.DevEUI[:], req.DevEUI)
		copy(joinPayload.DevNonce[:], req.DevNonce)
		payload.MACPayload = &joinPayload
		err := payload.SetMIC(lorawan.AES128Key(*devStorage.OutRead.Entry.AppKey))
		FatalUnless(t, err)
		req.MIC = payload.MIC[:]

		// Expect
		var wantErr *string
		var wantRes = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{}, // We'll check it by decoding
			NwkSKey: nil,                       // We'll assume it's correct if payload is okay
			Metadata: &core.Metadata{
				DataRate:    "SF7BW125",
				Frequency:   865.5,
				CodingRate:  "4/5",
				Timestamp:   uint32(tmst.Add(5*time.Second).Unix() * 1000000),
				PayloadSize: 33,
				Power:       14,
				InvPolarity: true,
			},
		}
		var wantAppReq = &core.JoinAppReq{
			Metadata: []*core.Metadata{req.Metadata},
			AppEUI:   req.AppEUI,
			DevEUI:   req.DevEUI,
		}

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes.Metadata, res.Metadata, "Join Handler Responses")
		Check(t, 16, len(res.NwkSKey), "Network session keys' length")
		Check(t, 4, len(res.DevAddr), "Device addresses' length")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
		joinaccept := lorawan.NewPHYPayload(false)
		err = joinaccept.UnmarshalBinary(res.Payload.Payload)
		CheckErrors(t, nil, err)
		err = joinaccept.DecryptJoinAcceptPayload(lorawan.AES128Key(*devStorage.InUpsert.Entry.AppKey))
		CheckErrors(t, nil, err)
		Check(t, handler.(*component).Configuration.NetID, joinaccept.MACPayload.(*lorawan.JoinAcceptPayload).NetID, "Network IDs")
	}

	// --------------------

	{
		Desc(t, "Handle valid join-request, fails to store")

		// Build
		tmst := time.Now()

		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			AppKey: &[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppEUI: req.AppEUI,
			DevEUI: req.DevEUI,
		}
		devStorage.Failures["upsert"] = errors.New(errors.Operational, "Mock Error")
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		payload := lorawan.NewPHYPayload(true)
		payload.MHDR = lorawan.MHDR{MType: lorawan.JoinRequest, Major: lorawan.LoRaWANR1}
		joinPayload := lorawan.JoinRequestPayload{}
		copy(joinPayload.AppEUI[:], req.AppEUI)
		copy(joinPayload.DevEUI[:], req.DevEUI)
		copy(joinPayload.DevNonce[:], req.DevNonce)
		payload.MACPayload = &joinPayload
		err := payload.SetMIC(lorawan.AES128Key(*devStorage.OutRead.Entry.AppKey))
		FatalUnless(t, err)
		req.MIC = payload.MIC[:]

		// Expect
		var wantErr = ErrOperational
		var wantRes = new(core.JoinHandlerRes)
		var wantAppReq *core.JoinAppReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Join Handler Responses")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
	}

	// --------------------

	{
		Desc(t, "Handle valid join-request, no gateway available")

		// Build
		tmst := time.Now()

		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateBlocked),
				DutyRX2:    uint32(dutycycle.StateBlocked),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			AppKey: &[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppEUI: req.AppEUI,
			DevEUI: req.DevEUI,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		payload := lorawan.NewPHYPayload(true)
		payload.MHDR = lorawan.MHDR{MType: lorawan.JoinRequest, Major: lorawan.LoRaWANR1}
		joinPayload := lorawan.JoinRequestPayload{}
		copy(joinPayload.AppEUI[:], req.AppEUI)
		copy(joinPayload.DevEUI[:], req.DevEUI)
		copy(joinPayload.DevNonce[:], req.DevNonce)
		payload.MACPayload = &joinPayload
		err := payload.SetMIC(lorawan.AES128Key(*devStorage.OutRead.Entry.AppKey))
		FatalUnless(t, err)
		req.MIC = payload.MIC[:]

		// Expect
		var wantErr = ErrOperational
		var wantRes = new(core.JoinHandlerRes)
		var wantAppReq *core.JoinAppReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Join Handler Responses")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
	}

	// --------------------

	{
		Desc(t, "Handle invalid join request -> Invalid datarate")

		// Build
		tmst := time.Now()

		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: &core.Metadata{
				DataRate:   "Not A DataRate",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			AppKey: &[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppEUI: req.AppEUI,
			DevEUI: req.DevEUI,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		payload := lorawan.NewPHYPayload(true)
		payload.MHDR = lorawan.MHDR{MType: lorawan.JoinRequest, Major: lorawan.LoRaWANR1}
		joinPayload := lorawan.JoinRequestPayload{}
		copy(joinPayload.AppEUI[:], req.AppEUI)
		copy(joinPayload.DevEUI[:], req.DevEUI)
		copy(joinPayload.DevNonce[:], req.DevNonce)
		payload.MACPayload = &joinPayload
		err := payload.SetMIC(lorawan.AES128Key(*devStorage.OutRead.Entry.AppKey))
		FatalUnless(t, err)
		req.MIC = payload.MIC[:]

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinHandlerRes)
		var wantAppReq *core.JoinAppReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Join Handler Responses")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
	}

	// --------------------

	{
		Desc(t, "Handle invalid join-request, lookup fails")

		// Build
		tmst := time.Now()

		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			MIC:      []byte{1, 2, 3, 4},
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		devStorage.Failures["read"] = errors.New(errors.NotFound, "Mock Error")
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		// Expect
		var wantErr = ErrNotFound
		var wantRes = new(core.JoinHandlerRes)
		var wantAppReq *core.JoinAppReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Join Handler Responses")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
	}

	// --------------------

	{
		Desc(t, "Handle invalid join-request -> invalid devEUI")

		// Build
		tmst := time.Now()

		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinHandlerRes)
		var wantAppReq *core.JoinAppReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Join Handler Responses")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
	}

	// --------------------

	{
		Desc(t, "Handle invalid join-request -> invalid appEUI")

		// Build
		tmst := time.Now()

		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinHandlerRes)
		var wantAppReq *core.JoinAppReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Join Handler Responses")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
	}

	// --------------------

	{
		Desc(t, "Handle invalid join-request -> invalid devNonce")

		// Build
		tmst := time.Now()

		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: nil,
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinHandlerRes)
		var wantAppReq *core.JoinAppReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Join Handler Responses")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
	}

	// --------------------

	{
		Desc(t, "Handle invalid join-request -> invalid Metadata")

		// Build
		req := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: nil,
		}

		devStorage := NewMockDevStorage()
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		// Expect
		var wantErr = ErrStructural
		var wantRes = new(core.JoinHandlerRes)
		var wantAppReq *core.JoinAppReq

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})
		res, err := handler.HandleJoin(context.Background(), req)

		// Check
		CheckErrors(t, wantErr, err)
		Check(t, wantRes, res, "Join Handler Responses")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
	}

	// -------------------

	{
		Desc(t, "Handle valid join-request (2 packets)")

		// Build
		tmst1 := time.Now()
		tmst2 := time.Now().Add(42 * time.Millisecond)

		req1 := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: &core.Metadata{
				DataRate:   "SF7BW125",
				Frequency:  865.5,
				Timestamp:  uint32(tmst1.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateWarning),
				DutyRX2:    uint32(dutycycle.StateWarning),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		req2 := &core.JoinHandlerReq{
			AppEUI:   []byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   []byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: []byte{14, 42},
			Metadata: &core.Metadata{
				DataRate:   "SF8BW125",
				Frequency:  867.234,
				Timestamp:  uint32(tmst2.Unix() * 1000000),
				CodingRate: "4/5",
				DutyRX1:    uint32(dutycycle.StateAvailable),
				DutyRX2:    uint32(dutycycle.StateAvailable),
				Rssi:       -20,
				Lsnr:       5.0,
			},
		}

		devStorage := NewMockDevStorage()
		devStorage.OutRead.Entry = devEntry{
			AppKey: &[16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			AppEUI: req1.AppEUI,
			DevEUI: req1.DevEUI,
		}
		pktStorage := NewMockPktStorage()
		appAdapter := mocks.NewAppClient()
		broker := mocks.NewAuthBrokerClient()

		payload := lorawan.NewPHYPayload(true)
		payload.MHDR = lorawan.MHDR{MType: lorawan.JoinRequest, Major: lorawan.LoRaWANR1}
		joinPayload := lorawan.JoinRequestPayload{}
		copy(joinPayload.AppEUI[:], req1.AppEUI)
		copy(joinPayload.DevEUI[:], req1.DevEUI)
		copy(joinPayload.DevNonce[:], req1.DevNonce)
		payload.MACPayload = &joinPayload
		err := payload.SetMIC(lorawan.AES128Key(*devStorage.OutRead.Entry.AppKey))
		FatalUnless(t, err)
		req1.MIC = payload.MIC[:]
		req2.MIC = payload.MIC[:]

		// Expect
		var wantErr1 *string
		var wantErr2 *string
		var wantRes1 = new(core.JoinHandlerRes)
		var wantRes2 = &core.JoinHandlerRes{
			Payload: &core.LoRaWANJoinAccept{}, // We'll check it by decoding
			NwkSKey: nil,                       // We'll assume it's correct if payload is okay
			Metadata: &core.Metadata{
				DataRate:    "SF8BW125",
				Frequency:   867.234,
				CodingRate:  "4/5",
				Timestamp:   uint32(tmst2.Add(5*time.Second).Unix() * 1000000),
				PayloadSize: 33,
				Power:       14,
				InvPolarity: true,
			},
		}
		var wantAppReq = &core.JoinAppReq{
			Metadata: []*core.Metadata{req1.Metadata, req2.Metadata},
			AppEUI:   req1.AppEUI,
			DevEUI:   req1.DevEUI,
		}

		// Operate
		handler := New(Components{
			Ctx:        GetLogger(t, "Handler"),
			Broker:     broker,
			AppAdapter: appAdapter,
			DevStorage: devStorage,
			PktStorage: pktStorage,
		}, Options{PublicNetAddr: "localhost", PrivateNetAddr: "localhost"})

		chack := make(chan bool)
		go func() {
			var ok bool
			defer func(ok *bool) { chack <- *ok }(&ok)
			res, err := handler.HandleJoin(context.Background(), req1)

			// Check
			CheckErrors(t, wantErr1, err)
			Check(t, wantRes1, res, "Data Up Handler Responses")
			ok = true
		}()

		go func() {
			<-time.After(bufferDelay / 3)
			var ok bool
			defer func(ok *bool) { chack <- *ok }(&ok)
			res, err := handler.HandleJoin(context.Background(), req2)

			// Check
			CheckErrors(t, wantErr2, err)
			Check(t, wantRes2.Metadata, res.Metadata, "Join Handler Responses")
			Check(t, 16, len(res.NwkSKey), "Network session keys' length")
			Check(t, 4, len(res.DevAddr), "Device addresses' length")
			joinaccept := lorawan.NewPHYPayload(false)
			err = joinaccept.UnmarshalBinary(res.Payload.Payload)
			CheckErrors(t, nil, err)
			err = joinaccept.DecryptJoinAcceptPayload(lorawan.AES128Key(*devStorage.InUpsert.Entry.AppKey))
			CheckErrors(t, nil, err)
			Check(t, handler.(*component).Configuration.NetID, joinaccept.MACPayload.(*lorawan.JoinAcceptPayload).NetID, "Network IDs")
			ok = true
		}()

		// Check
		ok1, ok2 := <-chack, <-chack
		Check(t, true, ok1 && ok2, "Acknowledgements")
		Check(t, wantAppReq, appAdapter.InHandleJoin.Req, "Join Application Requests")
	}

}

func TestStart(t *testing.T) {
	handler := New(Components{
		Ctx:        GetLogger(t, "Handler"),
		DevStorage: NewMockDevStorage(),
		PktStorage: NewMockPktStorage(),
		AppAdapter: mocks.NewAppClient(),
	}, Options{PublicNetAddr: "localhost:8888", PrivateNetAddr: "localhost:8889"})

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
