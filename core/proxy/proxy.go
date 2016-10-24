// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package proxy

import (
	"net/http"
	"strings"
)

type tokenProxier struct {
	handler http.Handler
}

func (p *tokenProxier) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if authorization := req.Header.Get("authorization"); authorization != "" {
		if len(authorization) > 6 && strings.ToLower(authorization[0:7]) == "bearer " {
			req.Header.Set("Grpc-Metadata-Token", authorization[7:])
		}
	}
	p.handler.ServeHTTP(res, req)
}

// WithToken wraps the handler so that each request gets the Bearer token attached
func WithToken(handler http.Handler) http.Handler {
	return &tokenProxier{handler}
}
