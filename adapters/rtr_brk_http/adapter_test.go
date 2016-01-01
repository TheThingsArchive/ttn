// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package rtr_brk_http

import (
	"encoding/json"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/testing/mock_components"
	"github.com/thethingsnetwork/core/utils/log"
	"github.com/thethingsnetwork/core/utils/pointer"
	. "github.com/thethingsnetwork/core/utils/testing"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

// ----- The adapter can be created and listen straigthforwardly
func TestListenOptionsTest(t *testing.T) {
	adapter, router := generateAdapterAndRouter(t)

	Desc(t, "Listen to adapter")
	if err := adapter.Listen(router, nil); err != nil {
		Ko(t, "No error was expected but got: %+v", err)
		return
	}
	Ok(t)
}

// ----- The adapter should forward a payload to a set of brokers
func TestForwardPayload(t *testing.T) {
	tests := []forwardPayloadTest{
		{generateValidPayload(), []string{"0.0.0.0:3000", "0.0.0.0:3001"}, nil},
		{generateInvalidPayload(), []string{"0.0.0.0:3002"}, core.ErrInvalidPayload},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type forwardPayloadTest struct {
	payload semtech.Payload
	brokers []string
	want    error
}

func (test forwardPayloadTest) run(t *testing.T) {
	Desc(t, "Forward %v to %v", test.payload, test.brokers)
	adapter, router := generateAdapterAndRouter(t)
	cmsg := listenHTTP(t, test.brokers)
	<-time.After(time.Millisecond * 250)
	got := adapter.Forward(router, test.payload, toBrokerAddrs(test.brokers)...)
	test.check(t, cmsg, got)
}

func (test forwardPayloadTest) check(t *testing.T, cmsg chan semtech.Payload, got error) {
	<-time.After(time.Millisecond * 500)

	// Check for the error
	if test.want != nil {
		if test.want != got {
			Ko(t, "Expected error %v but got %v", test.want, got)
			return
		}
		Ok(t)
		return
	}

	// Check if payload should have been sent
	if len(test.brokers) == 0 {
		Ok(t)
		return
	}

	// Gather payloads and check one of them
	var payloads []semtech.Payload
	select {
	case payload := <-cmsg:
		payloads = append(payloads, payload)
		if len(payloads) == len(test.brokers) {
			break
		}
	case <-time.After(time.Millisecond * 500):
		Ko(t, "%d payload(s) send to server(s) whereas %d was/were expected", len(payloads), len(test.brokers))
		return
	}

	if !reflect.DeepEqual(test.payload, payloads[0]) {
		Ko(t, "Expected %+v to be sent but server received: %+v", test.payload, payloads[0])
		return
	}

	Ok(t)
}

// ----- Build Utilities
func generateAdapterAndRouter(t *testing.T) (Adapter, core.Router) {
	return Adapter{
		Logger: log.TestLogger{
			Tag: "Adapter",
			T:   t,
		},
	}, mock_components.NewRouter()
}

func generateValidPayload() semtech.Payload {
	return semtech.Payload{
		RXPK: []semtech.RXPK{{
			Data: pointer.String("-DS4CGaDCdG+48eJNM3Vai-zDpsR71Pn9CPA9uCON84"),
			Freq: pointer.Float64(866.349812),
			Rssi: pointer.Int(-35),
		},
		},
	}
}

func generateInvalidPayload() semtech.Payload {
	return semtech.Payload{}
}

func toBrokerAddrs(addrs []string) []core.BrokerAddress {
	brokers := make([]core.BrokerAddress, 0)
	for _, addr := range addrs {
		brokers = append(brokers, core.BrokerAddress(addr))
	}
	return brokers
}

// ----- Operate Utilities
func listenHTTP(t *testing.T, addrs []string) chan semtech.Payload {
	cmsg := make(chan semtech.Payload)

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		res.Header().Set("Content-Type", "application/json")

		// Check the header type
		if req.Header.Get("Content-Type") != "application/json" {
			t.Log("Unexpected content-type ignore")
			res.WriteHeader(http.StatusBadRequest)
			res.Write(nil)
			return
		}

		// Check the body as well
		var payload semtech.Payload
		raw := make([]byte, 512)
		n, err := req.Body.Read(raw)
		if err != nil && err != io.EOF {
			t.Logf("Error reading request body: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			res.Write(nil)
			return
		}

		if err := json.Unmarshal(raw[:n], &payload); err != nil {
			t.Logf("Error while unmarshaling: %v", err)
			res.WriteHeader(http.StatusBadRequest)
			res.Write(nil)
			return
		}

		// Send a fake response
		res.WriteHeader(http.StatusOK)
		res.Write(nil)
		cmsg <- payload
	})

	for _, addr := range addrs {
		go func(addr string) {
			s := &http.Server{Addr: addr, Handler: serveMux}
			if err := s.ListenAndServe(); err != nil {
				panic(err)
			}
		}(addr)
	}

	return cmsg
}
