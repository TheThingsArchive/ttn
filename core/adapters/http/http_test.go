// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

type testRecipient struct {
	Recipient
	Behavior string
}

type testRegistration struct {
	Recipient testRecipient
	DevEUI    lorawan.EUI64
}

type testPacket struct {
	devEUI  lorawan.EUI64
	payload string
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p testPacket) MarshalBinary() ([]byte, error) {
	if p.payload == "" {
		return nil, errors.New(errors.Structural, "Fake error")
	}

	return []byte(p.payload), nil
}

// DevEUI implements the core.Addressable interface
func (p testPacket) DevEUI() lorawan.EUI64 {
	return p.devEUI
}

// String implements the core.Packet interface
func (p testPacket) String() string {
	return p.payload
}

func TestSend(t *testing.T) {
	recipients := []testRecipient{
		testRecipient{
			Recipient: recipient{
				url:    "0.0.0.0:3110",
				method: "POST",
			},
			Behavior: "AlwaysReject",
		},
		testRecipient{
			Recipient: recipient{
				url:    "0.0.0.0:3111",
				method: "POST",
			},
			Behavior: "AlwaysAccept",
		},
		testRecipient{
			Recipient: recipient{
				url:    "0.0.0.0:3112",
				method: "POST",
			},
			Behavior: "AlwaysReject",
		},
		testRecipient{
			Recipient: recipient{
				url:    "0.0.0.0:3113",
				method: "POST",
			},
			Behavior: "AlwaysReject",
		},
	}

	tests := []struct {
		Recipients        []testRecipient
		Packet            testPacket
		WantRegistrations []testRegistration
		WantPayload       string
		WantError         *string
	}{
		{ // Send to recipient a valid packet
			Recipients: recipients[1:2], // TODO test with a rejection. Need better error handling
			Packet: testPacket{
				devEUI:  lorawan.EUI64([8]byte{0, 0, 0, 0, 1, 2, 3, 4}),
				payload: "payload",
			},
			WantRegistrations: nil,
			WantPayload:       "payload",
			WantError:         nil,
		},
		{ // Broadcast a valid packet
			Recipients: nil,
			Packet: testPacket{
				devEUI:  lorawan.EUI64([8]byte{0, 0, 0, 0, 1, 2, 3, 4}),
				payload: "payload",
			},
			WantRegistrations: []testRegistration{
				{
					Recipient: recipients[1],
					DevEUI:    lorawan.EUI64([8]byte{0, 0, 0, 0, 1, 2, 3, 4}),
				},
			},
			WantPayload: "payload",
			WantError:   nil,
		},
		{ // Send to two recipients an invalid packet
			Recipients:        recipients[:2],
			Packet:            testPacket{},
			WantRegistrations: nil,
			WantPayload:       "",
			WantError:         pointer.String(string(errors.Structural)),
		},
		{ // Broadcast an invalid packet
			Recipients:        nil,
			Packet:            testPacket{},
			WantRegistrations: nil,
			WantPayload:       "",
			WantError:         pointer.String(string(errors.Structural)),
		},
	}

	// Logging
	ctx := GetLogger(t, "Adapter")

	// Build
	adapter, err := NewAdapter(3115, toHTTPRecipient(recipients), ctx)
	if err != nil {
		panic(err)
	}
	var servers []chan string
	for _, r := range recipients {
		servers = append(servers, genMockServer(r))
	}
	<-time.After(100 * time.Millisecond)

	for _, test := range tests {
		// Describe
		Desc(t, "Sending packet %v to %v", test.Packet, test.Recipients)

		// Operate
		_, err := adapter.Send(test.Packet, toHTTPRecipient(test.Recipients)...)
		registrations := getRegistrations(adapter, test.WantRegistrations)
		payloads := getPayloads(servers)

		// Check
		<-time.After(time.Second)
		CheckErrors(t, test.WantError, err)
		checkPayloads(t, test.WantPayload, payloads)
		checkRegistrations(t, test.WantRegistrations, registrations)
	}
}

// Convert testRecipient to core.Recipient
func toHTTPRecipient(recipients []testRecipient) []core.Recipient {
	var https []core.Recipient
	for _, r := range recipients {
		https = append(https, r.Recipient)
	}
	return https
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

func getRegistrations(adapter *Adapter, want []testRegistration) []core.Registration {
	var got []core.Registration
	for range want {
		ch := make(chan core.Registration)
		go func() {
			r, an, err := adapter.NextRegistration()
			if err != nil {
				return
			}
			an.Ack(nil)
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
		if req.Header.Get("Content-Type") != "application/octet-stream" {
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

		switch recipient.(testRecipient).Behavior {
		case "AlwaysReject":
			w.WriteHeader(http.StatusNotFound)
			w.Write(nil)
		case "AlwaysAccept":
			w.Header().Add("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write(buf[:n]) // TODO, should respond another packet, not the same
		}
		go func() { chresp <- string(buf[:n]) }()
	})

	server := http.Server{
		Addr:    recipient.(Recipient).URL(),
		Handler: serveMux,
	}
	go server.ListenAndServe()
	return chresp
}

// Check utilities
func checkRegistrations(t *testing.T, want []testRegistration, got []core.Registration) {
	if len(want) != len(got) {
		Ko(t, "Expected %d registrations but got %d", len(want), len(got))
		return
	}

outer:
	for _, rw := range want {
		for _, rg := range got {
			if rg.DevEUI() != rw.DevEUI {
				Ko(t, "Expected registration for %v but got for %v", rw.DevEUI, rg.DevEUI())
			}
			if reflect.DeepEqual(rw.Recipient.Recipient, rg.Recipient()) {
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
