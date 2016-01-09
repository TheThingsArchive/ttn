// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
	components "github.com/thethingsnetwork/core/refactored_components"
	"github.com/thethingsnetwork/core/utils/log"
	"github.com/thethingsnetwork/core/utils/pointer"
	. "github.com/thethingsnetwork/core/utils/testing"
	"io"
	"net/http"
	"testing"
	"time"
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
	adapter, err := NewAdapter(log.TestLogger{Tag: "BRK_HDL_ADAPTER", T: t})
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
		w.Write(nil)
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
		Metadata: &components.Metadata{Rssi: pointer.Int(-20), Modu: pointer.String("LORA")},
	}
}

func genJSONPayload(p core.Packet) string {
	raw, err := p.Payload.MarshalBinary()
	if err != nil {
		panic(err)
	}
	payload := base64.StdEncoding.EncodeToString(raw)
	metadatas, err := json.Marshal(p.Metadata)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`{"payload":"%s","metadata":%s}`, payload, string(metadatas))
}
