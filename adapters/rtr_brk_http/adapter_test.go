// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package rtr_brk_http

import (
	"encoding/json"
	"fmt"
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
	adapter, router := genAdapterAndRouter(t)

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
		{genValidPayload(), genBrokers([]int{200}), nil},
		{genValidPayload(), genBrokers([]int{200, 200}), nil},
		{genInvalidPayload(), nil, core.ErrInvalidPayload},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type forwardPayloadTest struct {
	payload semtech.Payload
	brokers map[string]int
	want    error
}

func (test forwardPayloadTest) run(t *testing.T) {
	//Describe
	Desc(t, "Forward %v to %v", test.payload, test.brokers)

	// Build
	adapter, router := genAdapterAndRouter(t)
	adapter.Listen(router, toBrokerAddrs(test.brokers))
	cmsg := listenHTTP(t, test.brokers)

	// Operate
	<-time.After(time.Millisecond * 250)
	got := adapter.Forward(router, test.payload, toBrokerAddrs(test.brokers)...)

	// Check
	<-time.After(time.Millisecond * 250)
	checkErrors(t, test.want, got)
	checkReception(t, len(test.brokers), test.payload, cmsg)
}

// ----- The adapter should broadcast a payload to a set of broker
func TestBroadcastPayload(t *testing.T) {
	tests := []broadcastPayloadTest{
		{genValidPayload(), genBrokers([]int{200, 200}), nil},
		{genValidPayload(), genBrokers([]int{200, 404}), nil},
		{genValidPayloadInvalidDevAddr(), nil, core.ErrInvalidPayload},
		{genInvalidPayload(), nil, core.ErrInvalidPayload},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type broadcastPayloadTest struct {
	payload semtech.Payload
	brokers map[string]int
	want    error
}

func (test broadcastPayloadTest) run(t *testing.T) {
	// Describe
	Desc(t, "Forward %v to %v", test.payload, test.brokers)

	// Build
	adapter, router := genAdapterAndRouter(t)
	adapter.Listen(router, toBrokerAddrs(test.brokers))
	cmsg := listenHTTP(t, test.brokers)

	// Operate
	<-time.After(time.Millisecond * 250)
	got := adapter.Broadcast(router, test.payload)

	// Check
	<-time.After(time.Millisecond * 250)
	checkErrors(t, test.want, got)
	checkReception(t, len(test.brokers), test.payload, cmsg)
	checkRegistration(t, router, test.payload, test.brokers)
}

// ----- Build Utilities

// Create an instance of an Adapter with a predefined logger + a mock router
func genAdapterAndRouter(t *testing.T) (Adapter, core.Router) {
	return Adapter{
		Logger: log.TestLogger{
			Tag: "Adapter",
			T:   t,
		},
	}, mock_components.NewRouter()
}

// gen a very basic payload holding an RXPK packet and identifying a valid device address
func genValidPayload() semtech.Payload {
	return semtech.Payload{
		RXPK: []semtech.RXPK{{
			Data: pointer.String(""),
			Freq: pointer.Float64(866.349812),
			Rssi: pointer.Int(-35),
		},
		},
	}
}

// gen a very basic payload holding an RXPK packet but with scrap data
func genValidPayloadInvalidDevAddr() semtech.Payload {
	return semtech.Payload{
		RXPK: []semtech.RXPK{{
			Data: pointer.String("-DS4CGaDCdG+48eJNM3Vai-zDpsR71Pn9CPA9uCON84"),
			Freq: pointer.Float64(866.349812),
			Rssi: pointer.Int(-35),
		},
		},
	}
}

// gen a payload with no RXPK nor STAT packet
func genInvalidPayload() semtech.Payload {
	return semtech.Payload{}
}

// Keep track of open TCP ports
var port int = 3000

// gen a list of brokers given a list of http response status in the form address -> status
func genBrokers(status []int) map[string]int {
	brokers := make(map[string]int)
	for _, s := range status {
		brokers[fmt.Sprintf("0.0.0.0:%d", port)] = s
		port += 1
	}
	return brokers
}

// Transform the broker map address -> status to a list a BrokerAddress
func toBrokerAddrs(addrs map[string]int) []core.BrokerAddress {
	brokers := make([]core.BrokerAddress, 0)
	for addr := range addrs {
		brokers = append(brokers, core.BrokerAddress(addr))
	}
	return brokers
}

// Create an http handler that will listen to json request on "/" and forward payload into a
// dedicated channel. A custom response status can be given in param.
func createServeMux(t *testing.T, status int, cmsg chan semtech.Payload) *http.ServeMux {
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
		res.WriteHeader(status)
		res.Write(nil)
		cmsg <- payload
	})
	return serveMux
}

// ----- Operate Utilities

// Start one http server per address which will forward request to the returned channel of payloads.
func listenHTTP(t *testing.T, addrs map[string]int) chan semtech.Payload {
	cmsg := make(chan semtech.Payload)

	for addr, status := range addrs {
		go func(addr string, status int) {
			s := &http.Server{Addr: addr, Handler: createServeMux(t, status, cmsg)}
			if err := s.ListenAndServe(); err != nil {
				panic(err)
			}
		}(addr, status)
	}

	return cmsg
}

// ----- Check Utilities
func checkErrors(t *testing.T, want error, got error) bool {
	// Check for the error
	if want != nil {
		if want != got {
			Ko(t, "Expected error %v but got %v", want, got)
			return false
		}
		Ok(t)
		return true
	}
	return true
}

func checkReception(t *testing.T, nbExpected int, want semtech.Payload, cmsg chan semtech.Payload) bool {
	// Check if payload should have been sent
	if nbExpected <= 0 {
		Ok(t)
		return true
	}

	// Gather payloads and check one of them
	var payloads []semtech.Payload
	select {
	case payload := <-cmsg:
		payloads = append(payloads, payload)
		if len(payloads) == nbExpected {
			break
		}
	case <-time.After(time.Millisecond * 500):
		Ko(t, "%d payload(s) send to server(s) whereas %d was/were expected", len(payloads), nbExpected)
		return false
	}

	if !reflect.DeepEqual(want, payloads[0]) {
		Ko(t, "Expected %+v to be sent but server received: %+v", want, payloads[0])
		return false
	}

	Ok(t)
	return true
}

func checkRegistration(t *testing.T, router core.Router, payload semtech.Payload, brokers map[string]int) bool {
	if len(brokers) == 0 {
		Ok(t)
		return true
	}

	devAddr, err := payload.UniformDevAddr()
	if err != nil {
		panic(err)
	}

	mockRouter := router.(*mock_components.Router) // Need to access to registered devices of mock router

outer:
	for addr, status := range brokers {
		if status != 200 { // Not a HTTP 200 OK, broker probably does not handle that device
			continue
		}

		addrs, ok := mockRouter.Devices[*devAddr] // Get all registered brokers for that device
		if !ok {
			Ko(t, "Broker %s wasn't registered for payload %v", addr, payload)
			return false
		}

		for _, broker := range addrs {
			if string(broker) == addr {
				continue outer // We are registered, everything's fine for that broker
			}
		}

		Ko(t, "Broker %s wasn't registered for payload %v", addr, payload)
		return false
	}

	Ok(t)
	return true
}
