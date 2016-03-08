// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

// ----- TYPES utilities
type testPacket struct {
	payload string
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p testPacket) MarshalBinary() ([]byte, error) {
	if p.payload == "" {
		return nil, errors.New(errors.Structural, "Fake error")
	}

	return []byte(p.payload), nil
}

// String implements the core.Packet interface
func (p testPacket) String() string {
	return p.payload
}

// DevEUI implements the devEUI
func (p testPacket) DevEUI() lorawan.EUI64 {
	return lorawan.EUI64{}
}

// ----- BUILD utilities
func createPubSubAdapter(t *testing.T, port uint) (*Adapter, string) {
	net := fmt.Sprintf("0.0.0.0:%d", port)
	adapter, err := NewAdapter(net, nil, GetLogger(t, "Adapter"))
	if err != nil {
		panic(err)
	}
	<-time.After(time.Millisecond * 250) // Let the connection starts
	handler := PubSub{}
	adapter.Bind(handler)
	return adapter, fmt.Sprintf("http://0.0.0.0:%d%s", port, handler.URL())
}

func createApplicationsAdapter(t *testing.T, port uint) (*Adapter, string) {
	net := fmt.Sprintf("0.0.0.0:%d", port)
	adapter, err := NewAdapter(net, nil, GetLogger(t, "Adapter"))
	if err != nil {
		panic(err)
	}
	<-time.After(time.Millisecond * 250) // Let the connection starts
	handler := Applications{}
	adapter.Bind(handler)
	return adapter, fmt.Sprintf("http://0.0.0.0:%d%s", port, handler.URL())
}

func createCollectAdapter(t *testing.T, port uint) (*Adapter, string) {
	net := fmt.Sprintf("0.0.0.0:%d", port)
	adapter, err := NewAdapter(net, nil, GetLogger(t, "Adapter"))
	if err != nil {
		panic(err)
	}
	<-time.After(time.Millisecond * 250) // Let the connection starts
	handler := Collect{}
	adapter.Bind(handler)
	return adapter, fmt.Sprintf("http://0.0.0.0:%d%s", port, handler.URL())
}

type testClient struct {
	http.Client
}

func (c testClient) Send(payload string, url string, method string, contentType string) chan MsgRes {
	buf := new(bytes.Buffer)
	if _, err := buf.Write([]byte(payload)); err != nil {
		panic(err)
	}

	request, err := http.NewRequest(method, url, buf)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", contentType)

	chresp := make(chan MsgRes)
	go func() {
		resp, err := c.Do(request)
		if err != nil {
			panic(err)
		}

		data := make([]byte, 2048)
		n, err := resp.Body.Read(data)
		if err != nil && err != io.EOF {
			panic(err)
		}

		chresp <- MsgRes{resp.StatusCode, data[:n]}
	}()
	return chresp
}

// ----- OPERATE utilities
func tryNext(adapter core.Adapter, shouldAck bool, packet core.Packet) ([]byte, error) {
	chresp := make(chan struct {
		Packet []byte
		Error  error
	})
	go func() {
		pkt, an, err := adapter.Next()
		defer func() {
			chresp <- struct {
				Packet []byte
				Error  error
			}{pkt, err}
		}()
		if err != nil {
			return
		}

		if shouldAck {
			an.Ack(packet)
		} else {
			an.Nack(nil)
		}
	}()

	select {
	case resp := <-chresp:
		return resp.Packet, resp.Error
	case <-time.After(time.Millisecond * 100):
		return nil, nil
	}
}

func tryNextRegistration(adapter core.Adapter, shouldAck bool, packet core.Packet) (core.Registration, error) {
	chresp := make(chan struct {
		Registration core.Registration
		Error        error
	})
	go func() {
		reg, an, err := adapter.NextRegistration()
		defer func() {
			chresp <- struct {
				Registration core.Registration
				Error        error
			}{reg, err}
		}()

		if err != nil {
			return
		}

		if shouldAck {
			an.Ack(packet)
		} else {
			an.Nack(nil)
		}
	}()

	select {
	case resp := <-chresp:
		return resp.Registration, resp.Error
	case <-time.After(time.Millisecond * 100):
		return nil, nil
	}
}

// ----- CHECK utilities
func checkStatusCode(t *testing.T, want int, got int) {
	if want == got {
		Ok(t, "Check status code")
		return
	}
	Ko(t, "Expected status code to be %d but got %d", want, got)
}

func checkContent(t *testing.T, want string, got []byte) {
	if strings.Contains(string(got), want) {
		Ok(t, "Check content")
		return
	}
	Ko(t, "Received content does not match expectations.\nWant: %s\nGot:  %s", want, string(got))
}

func checkPacket(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check packet")
		return
	}
	Ko(t, "Received packet does not match expectations.\nWant: %v\nGot:  %v", want, got)
}

func checkRegistration(t *testing.T, want core.Registration, got core.Registration) {
	if !reflect.DeepEqual(want, got) {
		Ko(t, "Received registration does not match expectations.\nWant: %v\nGot:  %v", want, got)
		return
	}
	Ok(t, "check Registrations")
}
