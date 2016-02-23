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

	core "github.com/TheThingsNetwork/ttn/refactor"
	//. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	//	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/refactor/adapters/http"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestCollect(t *testing.T) {
	tests := []struct {
		Desc    string
		Payload struct {
			ContentType string
			Method      string
			Content     string
		}
		ShouldAck bool
		AckPacket *core.Packet

		WantResponse []byte
		WantPacket   []byte
		WantError    *string
	}{}

	var port uint = 3000
	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build

		// CreateAdapter
		// Bind Collect Handler
		// Create Http Client

		// Operate
		// Send Payload + Content-type + Method
		// Try Next()
		// If Next Succeed -> Send given response

		// Check
		// Get server response from client
		// Get server statusCode from client
		// Check Errors
		// Check StatusCode
		// Check Content
		// Check Next packet[]
		port += 1
	}

}

// ----- BUILD utilities
func createAdapter(t *testing.T, port uint) *Adapter {
	adapter, err := NewAdapter(port, nil, GetLogger(t, "Adapter"))
	if err != nil {
		panic(err)
	}
	adapter.Bind(Collect{})
	return adapter
}

type testClient struct {
	http.Client
}

func (c testClient) Send(payload string, url string, method string, contentType string) (int, []byte) {
	buf := new(bytes.Buffer)
	if _, err := buf.Write([]byte(payload)); err != nil {
		panic(err)
	}

	request, err := http.NewRequest(method, url, buf)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", contentType)

	chresp := make(chan struct {
		StatusCode int
		Content    []byte
	})
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

		chresp <- struct {
			StatusCode int
			Content    []byte
		}{resp.StatusCode, data[:n]}
	}()

	select {
	case resp := <-chresp:
		return resp.StatusCode, resp.Content
	case <-time.After(time.Millisecond * 100):
		return 0, nil
	}
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

func checkContent(t *testing.T, want string, got []byte) {
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
