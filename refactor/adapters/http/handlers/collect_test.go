// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"bytes"
	//	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/adapters/http"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	//"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestCollect(t *testing.T) {
	tests := []struct {
		Desc        string
		Payload     string
		ContentType string
		Method      string
		ShouldAck   bool
		AckPacket   *core.Packet

		WantContent    []byte
		WantStatusCode int
		WantPacket     []byte
		WantError      *string
	}{}

	var port uint = 3000
	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		handler := Collect{}
		adapter := createAdapter(t, port, handler)
		client := testClient{}

		// Operate
		chresp := client.Send(test.Payload, handler.Url(), test.Method, test.ContentType)
		packet, err := tryNext(adapter, test.ShouldAck, test.AckPacket)
		var statusCode int
		var content []byte
		select {
		case resp := <-chresp:
			statusCode = resp.StatusCode
			content = resp.Content
		case <-time.After(time.Millisecond * 100):
		}

		// Check
		checkErrors(t, test.WantError, err)
		checkStatusCode(t, test.WantStatusCode, statusCode)
		checkContent(t, test.WantContent, content)
		checkPacket(t, test.WantPacket, packet)
		port += 1
	}

}

// ----- TYPES utilities
type testPacket struct {
	payload string
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p testPacket) MarshalBinary() ([]byte, error) {
	if p.payload == "" {
		return nil, errors.New(ErrInvalidStructure, "Fake error")
	}

	return []byte(p.payload), nil
}

// String implements the core.Packet interface
func (p testPacket) String() string {
	return p.payload
}

// ----- BUILD utilities
func createAdapter(t *testing.T, port uint, handler Handler) *Adapter {
	adapter, err := NewAdapter(port, nil, GetLogger(t, "Adapter"))
	if err != nil {
		panic(err)
	}
	adapter.Bind(handler)
	return adapter
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
func tryNext(adapter core.Adapter, shouldAck bool, packet *core.Packet) ([]byte, error) {
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
			an.Nack()
		}
	}()

	select {
	case resp := <-chresp:
		return resp.Packet, resp.Error
	case <-time.After(time.Millisecond * 100):
		return nil, nil
	}
}

// ----- CHECK utilities
func checkErrors(t *testing.T, want *string, got error) {
	if got == nil {
		if want == nil {
			Ok(t, "Check errors")
			return
		}
		Ko(t, "Expected error to be {%s} but got nothing", *want)
		return
	}

	if want == nil {
		Ko(t, "Expected no error but got {%v}", got)
		return
	}

	if got.(errors.Failure).Nature == *want {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%s} but got {%v}", *want, got)
}

func checkStatusCode(t *testing.T, want int, got int) {
	if want == got {
		Ok(t, "Check status code")
		return
	}
	Ko(t, "Expected status code to be %d but got %d", want, got)
}

func checkContent(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual([]byte(want), got) {
		Ok(t, "Check content")
		return
	}
	Ko(t, "Received content does not match expectations.\nWant: %v\nGot:  %v", want, got)
}

func checkPacket(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check packet")
		return
	}
	Ko(t, "Received packet does not match expectations.\nWant: %v\nGot:  %v", want, got)
}
