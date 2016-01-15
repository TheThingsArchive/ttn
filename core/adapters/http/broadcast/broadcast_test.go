// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broadcast

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	httpadapter "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/lorawan"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestSend(t *testing.T) {
	packet, devAddr, payload := genSample()
	recipients := []core.Recipient{
		core.Recipient{Address: "0.0.0.0:3010", Id: "AlwaysReject"},
		core.Recipient{Address: "0.0.0.0:3011", Id: "AlwaysAccept"},
		core.Recipient{Address: "0.0.0.0:3012", Id: "AlwaysReject"},
		core.Recipient{Address: "0.0.0.0:3013", Id: "AlwaysReject"},
	}
	registrations := []core.Registration{
		core.Registration{DevAddr: devAddr, Recipient: recipients[0]},
		core.Registration{DevAddr: devAddr, Recipient: recipients[1]},
		core.Registration{DevAddr: devAddr, Recipient: recipients[2]},
		core.Registration{DevAddr: devAddr, Recipient: recipients[3]},
	}

	tests := []struct {
		Recipients        []core.Recipient
		Packet            core.Packet
		WantRegistrations []core.Registration
		WantPayload       string
		WantError         error
	}{
		{ // Send to recipient a valid packet
			Recipients:        recipients[1:2], // TODO test with a rejection. Need better error handling
			Packet:            packet,
			WantRegistrations: []core.Registration{},
			WantPayload:       payload,
			WantError:         nil,
		},
		{ // Broadcast a valid packet
			Recipients:        []core.Recipient{},
			Packet:            packet,
			WantRegistrations: registrations[1:2],
			WantPayload:       payload,
			WantError:         nil,
		},
		{ // Send to two recipients an invalid packet
			Recipients:        recipients[:2],
			Packet:            core.Packet{},
			WantRegistrations: []core.Registration{},
			WantPayload:       "",
			WantError:         ErrInvalidPacket,
		},
		{ // Broadcast an invalid packet
			Recipients:        []core.Recipient{},
			Packet:            core.Packet{},
			WantRegistrations: []core.Registration{},
			WantPayload:       "",
			WantError:         ErrInvalidPacket,
		},
	}

	// Logging
	ctx := GetLogger(t, "Adapter")

	// Build
	a, err := httpadapter.NewAdapter(3015, httpadapter.JSONPacketParser{}, ctx)
	if err != nil {
		panic(err)
	}
	adapter, err := NewAdapter(a, recipients, ctx)
	if err != nil {
		panic(err)
	}
	var servers []chan string
	for _, r := range recipients {
		servers = append(servers, genMockServer(r))
	}

	for _, test := range tests {
		// Describe
		Desc(t, "Sending packet %v to %v", test.Packet, test.Recipients)
		<-time.After(time.Millisecond * 100)

		// Operate
		_, err := adapter.Send(test.Packet, test.Recipients...)
		registrations := getRegistrations(adapter, test.WantRegistrations)
		payloads := getPayloads(servers)

		// Check
		checkErrors(t, test.WantError, err)
		checkPayloads(t, test.WantPayload, payloads)
		checkRegistrations(t, test.WantRegistrations, registrations)
	}
}

// Operate utilities
func getPayloads(chpayloads []chan string) []string {
	var got []string
	for _, ch := range chpayloads {
		select {
		case payload := <-ch:
			got = append(got, payload)
		case <-time.After(50 * time.Millisecond):
		}
	}
	return got
}

func getRegistrations(adapter *Adapter, want []core.Registration) []core.Registration {
	var got []core.Registration
	for range want {
		ch := make(chan core.Registration)
		go func() {
			r, an, err := adapter.NextRegistration()
			if err != nil {
				return
			}
			an.Ack(core.Packet{})
			ch <- r
		}()
		select {
		case r := <-ch:
			got = append(got, r)
		case <-time.After(50 * time.Millisecond):
		}
	}
	return got
}

// Build utilities

func genMockServer(recipient core.Recipient) chan string {
	chresp := make(chan string)
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(nil)
			return
		}

		buf := make([]byte, req.ContentLength)
		n, err := req.Body.Read(buf)
		if err != nil && err != io.EOF {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(nil)
			return
		}

		switch recipient.Id {
		case "AlwaysReject":
			w.WriteHeader(http.StatusNotFound)
			w.Write(nil)
		case "AlwaysAccept":
			w.WriteHeader(http.StatusOK)
			w.Write(buf[:n]) // TODO, should respond another packet, not the same
		}
		go func() { chresp <- string(buf[:n]) }()
	})

	server := http.Server{
		Addr:    recipient.Address.(string),
		Handler: serveMux,
	}
	go server.ListenAndServe()
	return chresp
}

// Generate a Physical payload representing an uplink message
func genSample() (core.Packet, lorawan.DevAddr, string) {

	// 1. Generate a PHYPayload
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	devAddr := lorawan.DevAddr([4]byte{0x1, 0x14, 0x2, 0x42})

	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: devAddr,
		FCtrl: lorawan.FCtrl{
			ADR:       false,
			ADRACKReq: false,
			ACK:       false,
		},
		FCnt: 0,
	}
	macPayload.FPort = 10
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte("My Data")}}

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

	// 2. Generate a JSON payload received by the server
	packet := core.Packet{
		Payload:  payload,
		Metadata: core.Metadata{Rssi: pointer.Int(-20), Modu: pointer.String("LORA")},
	}
	jsonPayload, err := json.Marshal(packet)
	if err != nil {
		panic(err)
	}

	// 3. Return valuable info for the test
	return packet, devAddr, string(jsonPayload)
}

// Check utilities

func checkErrors(t *testing.T, want error, got error) {
	if want == got {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error %v but got %v", want, got)
}

func checkRegistrations(t *testing.T, want []core.Registration, got []core.Registration) {
outer:
	for _, rw := range want {
		for _, rg := range got {
			if reflect.DeepEqual(rg, rw) {
				continue outer
			}
		}
		Ko(t, "Registrations don't match expectation.\nWant: %v\nGot:  %v", want, got)
		return
	}
	Ok(t, "Check registrations")
}

func checkPayloads(t *testing.T, want string, got []string) {
	for _, payload := range got {
		if want != payload {
			Ko(t, "Paylaod don't match expectation.\nWant: %s\nGot:  %s", want, payload)
			return
		}
	}
	Ok(t, "Check payloads")
}
