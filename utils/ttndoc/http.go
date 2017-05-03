// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import "github.com/TheThingsNetwork/ttn/utils/ttndoc/internal/annotations"
import "regexp"
import "strings"

type HTTPEndpoint struct {
	Descriptor    *annotations.HttpRule
	Method        *Method
	RequestMethod string
	RequestPath   string
	Fields        []*Field
}

func newHTTPEndpoint(msg *annotations.HttpRule) *HTTPEndpoint {
	m := &HTTPEndpoint{
		Descriptor: msg,
	}
	return m
}

var httpEndpointFieldRegex = regexp.MustCompile(`\{[a-zA-Z0-9_-]+\}`)

func (h *HTTPEndpoint) Enter() {
	switch pattern := h.Descriptor.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		h.RequestMethod = "GET"
		h.RequestPath = pattern.Get
	case *annotations.HttpRule_Put:
		h.RequestMethod = "PUT"
		h.RequestPath = pattern.Put
	case *annotations.HttpRule_Post:
		h.RequestMethod = "POST"
		h.RequestPath = pattern.Post
	case *annotations.HttpRule_Delete:
		h.RequestMethod = "DELETE"
		h.RequestPath = pattern.Delete
	case *annotations.HttpRule_Patch:
		h.RequestMethod = "PATCH"
		h.RequestPath = pattern.Patch
	case *annotations.HttpRule_Custom:
		h.RequestMethod = pattern.Custom.Kind
		h.RequestPath = pattern.Custom.Path
	}
	if matches := httpEndpointFieldRegex.FindAllString(h.RequestPath, -1); matches != nil {
		for _, match := range matches {
			if field, ok := h.Method.Input.Message.Fields[h.Method.Input.Message.Name+"."+strings.Trim(match, "{}")]; ok {
				h.Fields = append(h.Fields, field)
			}
		}
	}
	for _, msg := range h.Descriptor.AdditionalBindings {
		httpEndpoint := newHTTPEndpoint(msg)
		httpEndpoint.Method = h.Method
		h.Method.HTTPEndpoints = append(h.Method.HTTPEndpoints, httpEndpoint)
		httpEndpoint.Enter()
	}
}
