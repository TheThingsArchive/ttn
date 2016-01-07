// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package brk_hdl_http

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/utils/log"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var ErrInvalidPort = fmt.Errorf("The given port is invalid")

type Adapter struct {
	logger log.Logger
	regs   chan regMsg
}

type regMsg struct {
	config core.Registration
	resp   chan httpResponse
}

type httpResponse struct {
	statusCode int
	content    []byte
}

func NewAdapter(port uint) (*Adapter, error) {
	if port == 0 {
		return nil, ErrInvalidPort
	}

	a := Adapter{
		regs:   make(chan regMsg),
		logger: log.DebugLogger{Tag: "BRK_HDL_ADAPTER"},
	}

	a.listen(port)

	return &a, nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, an core.AckNacker) error {
	return nil
}

// Next implements the core.Adapter inerface
func (a *Adapter) Next() (core.Packet, core.AckNacker, error) {
	return core.Packet{}, nil, nil
}

// NextRegistration implements the core.BrkHdlAdapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	return core.Registration{}, nil, nil
}

func (a *Adapter) listen(port uint) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/end-device/", func(w http.ResponseWriter, req *http.Request) {
		a.log("Receive new registration request")
		// Check Content-type
		contentType := req.Header.Get("Content-Type")
		if contentType != "application/json" {
			a.log("registration request rejected: not json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad content type"))
			return
		}

		// Check the query parameter
		reg := regexp.MustCompile("end-device/([a-fA-F0-9]{8})$")
		query := reg.FindStringSubmatch(req.RequestURI)
		if len(query) < 2 {
			a.log("registration request rejected: devAddr invalid")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect end-device address format"))
			return
		}
		devAddr := query[1]

		// Check configuration in body
		body := make([]byte, 256)
		n, err := req.Body.Read(body)
		if err != nil && err != io.EOF {
			a.log("registration request rejected: body unreadable")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect request body"))
			return
		}
		body = body[:n]
		a.log("registration payload: %s", string(body))
		params := &struct {
			Id     string `json:"app_id"`
			Url    string `json:"app_url"`
			NwsKey string `json:"nws_key"`
		}{}
		if err := json.Unmarshal(body, params); err != nil {
			a.log("registration request rejected: payload invalid or incomplete")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect body payload"))
			return
		}

		nwsKey, err := hex.DecodeString(params.NwsKey)
		if err != nil || len(nwsKey) != 16 {
			a.log("registration request rejected: nwsKey invalid")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect network session nws_key"))
			return
		}

		params.Id = strings.Trim(params.Id, " ")
		params.Url = strings.Trim(params.Url, " ")
		if len(params.Id) <= 0 {
			a.log("registration request rejected: appId invalid")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect config app_id"))
			return
		}
		if len(params.Url) <= 0 {
			a.log("registration request rejected: appUrl invalid")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect config app_url"))
			return
		}

		// Create registration
		config := core.Registration{
			Handler: core.Recipient{Id: params.Id, Address: params.Url},
		}
		copy(config.NwsKey[:], nwsKey)
		copy(config.DevAddr[:], devAddr)

		// Send the registration and wait for ack / nack
		resp := make(chan httpResponse)
		a.regs <- regMsg{config: config, resp: resp}
		r, ok := <-resp
		if !ok {
			a.log("Unexpected channel closure")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Core server not responding"))
			return
		}
		a.log("sending normal response to registration request")
		w.WriteHeader(r.statusCode)
		w.Write(r.content)
	})

	go func() {
		server := http.Server{
			Addr:    fmt.Sprintf("0.0.0.0:%d", port),
			Handler: serveMux,
		}
		a.log("Start listening on %d", port)
		err := server.ListenAndServe()
		a.log("HTTP connection lost: %v", err)
	}()
}

func (a *Adapter) log(format string, i ...interface{}) {
	if a == nil || a.logger == nil {
		return
	}
	a.logger.Log(format, i...)
}
