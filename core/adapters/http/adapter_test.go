// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"reflect"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

// Send(p core.Packet, r ...core.Recipient) error
func TestSend(t *testing.T) {
	tests := []struct {
		Packet      core.Packet
		WantPayload string
		WantError   error
	}{
		{
			genCorePacket(),
			genJSONPayload(genCorePacket()),
			nil,
		},
		{
			core.Packet{},
			"",
			ErrInvalidPacket,
		},
	}

	s := genMockServer(3100)

	// Logging
	ctx := GetLogger(t, "Adapter")

	adapter, err := NewAdapter(3101, JSONPacketParser{}, ctx)
	if err != nil {
		panic(err)
	}

	for _, test := range tests {
		Desc(t, "Sending packet: %v", test.Packet)
		<-time.After(time.Millisecond * 100)
		_, err := adapter.Send(test.Packet, s.Recipient)
		checkErrors(t, test.WantError, err)
		checkSend(t, test.WantPayload, s)
	}
}

// Next() (core.Packet, an core.AckNacker, error)
func TestNext(t *testing.T) {
	tests := []struct {
		Payload    string
		IsNotFound bool
		WantPacket core.Packet
		WantStatus int
		WantError  error
	}{
		{
			Payload:    genJSONPayload(genCorePacket()),
			IsNotFound: false,
			WantPacket: genCorePacket(),
			WantStatus: http.StatusOK,
			WantError:  nil,
		},
		{
			Payload:    genJSONPayload(genCorePacket()),
			IsNotFound: true,
			WantPacket: genCorePacket(),
			WantStatus: http.StatusNotFound,
			WantError:  nil,
		},
		{
			Payload:    "Patate",
			IsNotFound: false,
			WantPacket: core.Packet{},
			WantStatus: http.StatusBadRequest,
			WantError:  nil,
		},
	}
	// Build
	log.SetHandler(NewLogHandler(t))
	ctx := log.WithFields(log.Fields{"tag": "Adapter"})
	adapter, err := NewAdapter(3102, JSONPacketParser{}, ctx)
	if err != nil {
		panic(err)
	}

	c := client{adapter: "0.0.0.0:3102"}

	for _, test := range tests {
		// Describe
		Desc(t, "Send payload to the adapter %s. Will send ack ? %v", test.Payload, !test.IsNotFound)
		<-time.After(time.Millisecond * 100)

		// Operate
		gotPacket := make(chan core.Packet)
		gotError := make(chan error)
		go func() {
			packet, an, err := adapter.Next()
			if err == nil {
				if test.IsNotFound {
					an.Nack()
				} else {
					an.Ack()
				}
			}
			gotError <- err
			gotPacket <- packet
		}()

		resp := c.send(test.Payload)

		// Check
		select {
		case err := <-gotError:
			checkErrors(t, test.WantError, err)
		case <-time.After(time.Millisecond * 250):
			checkErrors(t, test.WantError, nil)
		}

		checkStatus(t, test.WantStatus, resp.StatusCode)

		// NOTE: See https://github.com/brocaar/lorawan/issues/3
		//select {
		//case packet := <-gotPacket:
		//	checkPackets(t, test.WantPacket, packet)
		//case <-time.After(time.Millisecond * 250):
		//	checkPackets(t, test.WantPacket, core.Packet{})
		//}

	}
}

// Check utilities
func checkErrors(t *testing.T, want error, got error) {
	if want == got {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%v} but got {%v}", want, got)
}

func checkSend(t *testing.T, want string, s MockServer) {
	select {
	case got := <-s.Payloads:
		if want != got {
			Ko(t, "Received payload does not match expectations.\nWant: %s\nGot:  %s", want, got)
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

func checkPackets(t *testing.T, want core.Packet, got core.Packet) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check packets")
		return
	}
	Ko(t, "Received packet does not match expectations.\nWant: %s\nGot:  %s", want, got)
}

func checkStatus(t *testing.T, want int, got int) {
	if want == got {
		Ok(t, "Check status")
		return
	}
	Ko(t, "Expected status to be %d but got %d", want, got)
}

// Build utilities
type MockServer struct {
	Recipient core.Recipient
	Payloads  chan string
}

func genMockServer(port uint) MockServer {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	payloads := make(chan string)

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		body := make([]byte, 256)
		n, err := req.Body.Read(body)
		if err != nil && err != io.EOF {
			panic(err)
		}
		w.Write(body[:n]) // NOTE TEMPORARY, the response is supposed to be different
		go func() { payloads <- string(body[:n]) }()
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
func genPHYPayload(msg string, devAddr [4]byte) lorawan.PHYPayload {
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: lorawan.DevAddr(devAddr),
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

func genCorePacket() core.Packet {
	return core.Packet{
		Payload:  genPHYPayload("myData", [4]byte{0x1, 0x2, 0x3, 0x4}),
		Metadata: core.Metadata{Rssi: pointer.Int(-20), Modu: pointer.String("LORA")},
	}
}

func genJSONPayload(p core.Packet) string {
	raw, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

type client struct {
	http.Client
	adapter string
}

// Operate utilities
// send is a convinient helper to send HTTP from a handler to the adapter
func (c *client) send(payload string) http.Response {
	buf := new(bytes.Buffer)
	if _, err := buf.WriteString(payload); err != nil {
		panic(err)
	}
	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s/packets", c.adapter), buf)
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
