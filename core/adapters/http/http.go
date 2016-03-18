// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"time"

	"github.com/apex/log"
)

// Interface defines the public interfaces used by the adapter
type Interface interface {
	Bind(Handler)
}

// adapter type materializes an http adapter
type adapter struct {
	Client   http.Client    // Adapter is also an http client
	ServeMux *http.ServeMux // Holds a references to the adapter servemux in order to dynamically define endpoints
	Components
}

// Handler defines endpoint-specific handler.
type Handler interface {
	URL() string
	Handle(w http.ResponseWriter, req *http.Request) error
}

// Components makes instantiation easier to read
type Components struct {
	Ctx log.Interface
}

// Options makes instantiation easier to read
type Options struct {
	NetAddr string
	Timeout time.Duration
}

// New instantiates a new adapter
func New(c Components, o Options) Interface {
	a := adapter{
		Components: c,
		ServeMux:   http.NewServeMux(),
		Client:     http.Client{Timeout: o.Timeout},
	}
	go a.listenRequests(o.NetAddr)
	return &a
}

// Bind registers a handler to a specific endpoint
func (a *adapter) Bind(h Handler) {
	a.Ctx.WithField("url", h.URL()).Info("Register new endpoint")
	a.ServeMux.HandleFunc(h.URL(), func(w http.ResponseWriter, req *http.Request) {
		Ctx := a.Ctx.WithField("url", h.URL())
		Ctx.Debug("Handle new request")
		err := h.Handle(w, req)
		if err != nil {
			Ctx.WithError(err).Debug("Failed to handle the request")
		}
	})
}

// listenRequests handles incoming registration request sent through http to the adapter
func (a *adapter) listenRequests(net string) {
	server := http.Server{
		Addr:    net,
		Handler: a.ServeMux,
	}
	a.Ctx.WithField("bind", net).Info("Starting Server")
	err := server.ListenAndServe()
	a.Ctx.WithError(err).Warn("HTTP connection lost")
}
