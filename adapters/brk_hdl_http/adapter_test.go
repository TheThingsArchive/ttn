// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package brk_hdl_http

import (
	"bytes"
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
	components "github.com/thethingsnetwork/core/refactored_components"
	"github.com/thethingsnetwork/core/utils/log"
	"github.com/thethingsnetwork/core/utils/pointer"
	. "github.com/thethingsnetwork/core/utils/testing"
	"net/http"
	"reflect"
	"testing"
	"time"
)

// NewAdapter(port uint, loggers ...log.Logger) (*Adapter, error)
func TestNewAdapter(t *testing.T) {
	tests := []struct {
		Port      uint
		WantError error
	}{
		{3000, nil},
		{0, ErrInvalidPort},
	}

	for _, test := range tests {
		Desc(t, "Create new adapter bound to %d", test.Port)
		_, err := NewAdapter(test.Port)
		checkErrors(t, test.WantError, err)
	}
}

// NextRegistration() (core.Registration, core.AckNacker, error)
func TestNextRegistration(t *testing.T) {
	tests := []struct {
		AppId      string
		AppUrl     string
		DevAddr    string
		NwsKey     string
		WantResult *core.Registration
		WantError  error
	}{
		// Valid device address
		{
			AppId:   "appid",
			AppUrl:  "myhandler.com:3000",
			NwsKey:  "000102030405060708090a0b0c0d0e0f",
			DevAddr: "14aab0a4",
			WantResult: &core.Registration{
				DevAddr: lorawan.DevAddr([4]byte{0x14, 0xaa, 0xb0, 0xa4}),
				Handler: core.Recipient{Id: "appid", Address: "myhandler.com:3000"},
				NwsKey:  lorawan.AES128Key([16]byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf}),
			},
			WantError: nil,
		},
		// Invalid device address
		{
			AppId:      "appid",
			AppUrl:     "myhandler.com:3000",
			NwsKey:     "000102030405060708090a0b0c0d0e0f",
			DevAddr:    "INVALID",
			WantResult: nil,
			WantError:  nil,
		},
		// Invalid nwskey address
		{
			AppId:      "appid",
			AppUrl:     "myhandler.com:3000",
			NwsKey:     "00112233445566778899af",
			DevAddr:    "14aab0a4",
			WantResult: nil,
			WantError:  nil,
		},
	}

	adapter, err := NewAdapter(3001, log.TestLogger{Tag: "BRK_HDL_ADAPTER", T: t})
	client := &client{adapter: "0.0.0.0:3001"}
	<-time.After(time.Millisecond * 200)
	if err != nil {
		panic(err)
	}

	for _, test := range tests {
		// Describe
		Desc(t, "Trying to register %s -> %s, %s, %s", test.DevAddr, test.AppId, test.AppUrl, test.NwsKey)

		// Build
		gotErr := make(chan error)
		gotConf := make(chan core.Registration)
		go client.send(test.AppId, test.AppUrl, test.DevAddr, test.NwsKey)

		// Operate
		go func() {
			config, _, err := adapter.NextRegistration()
			gotErr <- err
			gotConf <- config
		}()

		// Check
		select {
		case err := <-gotErr:
			checkErrors(t, test.WantError, err)
		case <-time.After(time.Millisecond * 250):
			checkErrors(t, test.WantError, nil)
		}

		select {
		case conf := <-gotConf:
			checkRegistrationResult(t, test.WantResult, &conf)
		case <-time.After(time.Millisecond * 250):
			checkRegistrationResult(t, test.WantResult, nil)
		}
	}
}

// Send(p core.Packet, r ...core.Recipient) error
func TestSend(t *testing.T) {
	tests := []struct {
		Packet      core.Packet
		WantPayload string
		WantError   error
	}{
		{
			core.Packet{
				Payload:  genPHYPayload("myData"),
				Metadata: &components.Metadata{Rssi: pointer.Int(-20), Modu: pointer.String("LORA")},
			},
			`{"metadata":{"rssi":-20,"modu":"LORA"},"payload":"myData"}`,
			nil,
		},
		{
			core.Packet{
				Payload:  lorawan.PHYPayload{},
				Metadata: &components.Metadata{},
			},
			"",
			ErrInvalidPacket,
		},
	}

	s := genMockServer(3100)
	adapter, err := NewAdapter(3101, log.TestLogger{Tag: "BRK_HDL_ADAPTER", T: t})
	if err != nil {
		panic(err)
	}

	for _, test := range tests {
		Desc(t, "Sending packet: %v", test.Packet)
		err := adapter.Send(test.Packet, s.Recipient)
		checkErrors(t, test.WantError, err)
		checkSend(t, test.WantPayload, s)
	}
}

func checkErrors(t *testing.T, want error, got error) {
	if want == got {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%v} but got {%v}", want, got)
}

func checkRegistrationResult(t *testing.T, want, got *core.Registration) {
	if !reflect.DeepEqual(want, got) {
		Ko(t, "Received configuration doesn't match expectations")
		return
	}

	Ok(t, "Check registration result")
}

func checkSend(t *testing.T, want string, s MockServer) {
	select {
	case got := <-s.Payloads:
		if want != got {
			Ko(t, "Expected payload %s to be sent but got %s", want, got)
			return
		}
	case <-time.After(time.Millisecond * 100):
		if want != "" {
			Ko(t, "Expected payload %s to be sent but got nothing", want)
			return
		}
	}
	Ok(t, "Check send result")
}

// Operate utilities

// Wrapper around the http client
type client struct {
	http.Client
	adapter string
}

// send is a convinient helper to send HTTP from a handler to the adapter
func (c *client) send(appId, appUrl, devAddr, nwsKey string) http.Response {
	buf := new(bytes.Buffer)
	if _, err := buf.WriteString(fmt.Sprintf(`{"app_id":"%s","app_url":"%s","nws_key":"%s"}`, appId, appUrl, nwsKey)); err != nil {
		panic(err)
	}
	request, err := http.NewRequest("PUT", fmt.Sprintf("http://%s/end-device/%s", c.adapter, devAddr), buf)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(request)
	if err != nil {
		panic(err)
	}
	return *resp
}

// Build utilities

type MockServer struct {
	Recipient core.Recipient
	Payloads  chan string
}

func genMockServer(port uint) MockServer {
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	payloads := make(chan string)

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		body := make([]byte, 256)
		n, err := req.Body.Read(body)
		if err != nil {
			panic(err)
		}
		payloads <- string(body[:n])
	})

	go func() {
		server := http.Server{
			Handler: serveMux,
			Addr:    addr,
		}
		server.ListenAndServe()
	}()

	<-time.After(time.Millisecond * 50)

	return MockServer{
		Recipient: core.Recipient{
			Address: addr,
			Id:      "Mock server",
		},
		Payloads: payloads,
	}
}

// Generate a Physical payload representing an uplink message
func genPHYPayload(msg string) lorawan.PHYPayload {
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: lorawan.DevAddr([4]byte{1, 2, 3, 4}),
		FCtrl: lorawan.FCtrl{
			ADR:       false,
			ADRACKReq: false,
			ACK:       false,
		},
		FCnt: 0,
	}
	macPayload.FPort = 10
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(msg)}}

	if err := macPayload.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	payload := lorawan.NewPHYPayload(true)
	payload.MHDR = lorawan.MHDR{
		MType: lorawan.ConfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	payload.MACPayload = macPayload

	if err := payload.SetMIC(nwkSKey); err != nil {
		panic(err)
	}

	return payload
}
