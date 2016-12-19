// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package proxy

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/apex/log"
)

type tokenProxier struct {
	handler http.Handler
}

func (p *tokenProxier) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if authorization := req.Header.Get("authorization"); authorization != "" {
		if len(authorization) >= 7 && strings.ToLower(authorization[0:7]) == "bearer " {
			req.Header.Set("Grpc-Metadata-Token", authorization[7:])
		}
		if len(authorization) >= 4 && strings.ToLower(authorization[0:4]) == "key " {
			req.Header.Set("Grpc-Metadata-Key", authorization[4:])
		}
	}
	p.handler.ServeHTTP(res, req)
}

// WithToken wraps the handler so that each request gets the Bearer token attached
func WithToken(handler http.Handler) http.Handler {
	return &tokenProxier{handler}
}

type logProxier struct {
	ctx     log.Interface
	handler http.Handler
}

func (p *logProxier) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	p.ctx.WithFields(log.Fields{
		"RemoteAddress": req.RemoteAddr,
		"Method":        req.Method,
		"URI":           req.RequestURI,
	}).Info("Proxy HTTP request")
	p.handler.ServeHTTP(res, req)
}

// WithLogger wraps the handler so that each request gets logged
func WithLogger(handler http.Handler, ctx log.Interface) http.Handler {
	return &logProxier{ctx, handler}
}

type paginatedHandler struct {
	handler http.Handler
}

func (h *paginatedHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var b struct {
		offsetQuery string
		offset      int

		limitQuery string
		limit      int

		err error
	}

	if b.offsetQuery = req.URL.Query().Get("offset"); b.offsetQuery != "" {
		if b.offset, b.err = strconv.Atoi(b.offsetQuery); b.err != nil {
			http.Error(res, b.err.Error(), http.StatusBadRequest)
			return
		}
	}

	if b.limitQuery = req.URL.Query().Get("limit"); b.limitQuery != "" {
		if b.limit, b.err = strconv.Atoi(b.limitQuery); b.err != nil {
			http.Error(res, b.err.Error(), http.StatusBadRequest)
			return
		}
	}

	req.Header.Set("Grpc-Metadata-Offset", strconv.Itoa(b.offset))
	req.Header.Set("Grpc-Metadata-Limit", strconv.Itoa(b.limit))

	h.handler.ServeHTTP(res, req)
}

// WithPagination wraps the handler so that each request gets the Page value attached
func WithPagination(h http.Handler) http.Handler {
	return &paginatedHandler{h}
}
