// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
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
	adapter, err := NewAdapter("0.0.0.0:3115", toHTTPRecipient(recipients), ctx)
	if err != nil {
		panic(err)
	}
	var servers []chan string
	for _, r := range recipients {
		servers = append(servers, genMockServer(r))
	}
	<-time.After(200 * time.Millisecond)

	for _, test := range tests {
		// Describe
		Desc(t, "Sending packet %v to %v", test.Packet, test.Recipients)

		// Operate
		_, err := adapter.Send(test.Packet, toHTTPRecipient(test.Recipients)...)
		registrations := getRegistrations(adapter, test.WantRegistrations)
		payloads := getPayloads(servers)

		// Check
		CheckErrors(t, test.WantError, err)
		checkPayloads(t, test.WantPayload, payloads)
		checkRegistrations(t, test.WantRegistrations, registrations)
	}
}

func TestSubscribe(t *testing.T) {
	{
		Desc(t, "Subscribe a valid registration")

		// Build
		r := mocks.NewMockRegistration()
		r.OutRecipient = NewRecipient("0.0.0.0:4777", "PUT")
		a, _ := NewAdapter("0.0.0.0:4776", nil, GetLogger(t, "Adapter"))
		serveMux := http.NewServeMux()
		serveMux.HandleFunc("/end-devices", func(w http.ResponseWriter, req *http.Request) {
			// Check
			CheckContentTypes(t, req.Header.Get("Content-Type"), "application/json")
			CheckMethods(t, req.Method, r.OutRecipient.(Recipient).Method())

			buf := make([]byte, req.ContentLength)
			n, err := req.Body.Read(buf)
			if err == io.EOF {
				err = nil
			}
			CheckErrors(t, nil, err)
			wantJSON := []byte(fmt.Sprintf(`{"recipient":{"method":"POST","url":"0.0.0.0:4776"},"registration":%s}`, r.OutMarshalJSON))
			CheckJSONs(t, wantJSON, buf[:n])
		})
		go http.ListenAndServe(r.OutRecipient.(Recipient).URL(), serveMux)
		<-time.After(time.Millisecond * 100)

		// Operate
		err := a.Subscribe(r)
		<-time.After(time.Millisecond * 50)

		// Check
		CheckErrors(t, nil, err)
	}

	// --------------------

	{
		Desc(t, "Subscribe an invalid registration -> Invalid recipient")

		// Build
		r := mocks.NewMockRegistration()
		r.OutRecipient = NewRecipient("0.0.0.0:4777", "PUT")
		r.Failures["MarshalJSON"] = errors.New(errors.Structural, "Mock Error")
		a, _ := NewAdapter("0.0.0.0:4776", nil, GetLogger(t, "Adapter"))

		// Operate
		err := a.Subscribe(r)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}

	// --------------------

	{
		Desc(t, "Subscribe an invalid registration -> MarshalJSON fails")

		// Build
		r := mocks.NewMockRegistration()
		r.Failures["MarshalJSON"] = errors.New(errors.Structural, "Mock Error")
		a, _ := NewAdapter("0.0.0.0:4776", nil, GetLogger(t, "Adapter"))

		// Operate
		err := a.Subscribe(r)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}

	// --------------------

	{
		Desc(t, "Subscribe a valid registration | Refused by server")

		// Build
		r := mocks.NewMockRegistration()
		r.OutRecipient = NewRecipient("0.0.0.0:4778", "PUT")
		a, _ := NewAdapter("0.0.0.0:4776", nil, GetLogger(t, "Adapter"))
		serveMux := http.NewServeMux()
		serveMux.HandleFunc("/end-devices", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(nil)
		})
		go http.ListenAndServe(r.OutRecipient.(Recipient).URL(), serveMux)
		<-time.After(time.Millisecond * 100)

		// Operate
		err := a.Subscribe(r)
		<-time.After(time.Millisecond * 50)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
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

func getRegistrations(adapter *Adapter, want []testRegistration) []core.RRegistration {
	var got []core.RRegistration
	for range want {
		ch := make(chan core.RRegistration)
		go func() {
			r, an, err := adapter.NextRegistration()
			if err != nil {
				return
			}
			an.Ack(nil)
			ch <- r.(core.RRegistration)
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
