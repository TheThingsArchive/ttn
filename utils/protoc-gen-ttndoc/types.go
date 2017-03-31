// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	options "google.golang.org/genproto/googleapis/api/annotations"
)

type service struct {
	key     string
	comment string
	*descriptor.ServiceDescriptorProto
	methods []*method
}

func (s *service) GoString() string {
	return s.key
}

type method struct {
	key     string
	comment string
	*descriptor.MethodDescriptorProto
	input        *message
	inputStream  bool
	output       *message
	outputStream bool
	endpoints    []*endpoint
}

func (m *method) GoString() string {
	var inputType, outputType string
	if m.inputStream {
		inputType += "stream "
	}
	inputType += m.input.GoString()

	if m.outputStream {
		outputType += "stream "
	}
	outputType += m.output.GoString()
	return fmt.Sprintf("%s (%s) -> (%s)", m.key, inputType, outputType)
}

type endpoint struct {
	method string
	url    string
}

func newEndpoint(opts *options.HttpRule) *endpoint {
	if opts == nil {
		return nil
	}
	switch opt := opts.GetPattern().(type) {
	case *options.HttpRule_Get:
		return &endpoint{"GET", opt.Get}
	case *options.HttpRule_Put:
		return &endpoint{"PUT", opt.Put}
	case *options.HttpRule_Post:
		return &endpoint{"POST", opt.Post}
	case *options.HttpRule_Delete:
		return &endpoint{"DELETE", opt.Delete}
	case *options.HttpRule_Patch:
		return &endpoint{"PATCH", opt.Patch}
	}
	return nil
}

func (e *endpoint) GoString() string {
	return fmt.Sprintf("%s %s", e.method, e.url)
}

type message struct {
	key     string
	comment string
	*descriptor.DescriptorProto
	fields []*field
	nested []*message
	oneofs []*oneof
}

func (m *message) GoString() string {
	return m.key
}

func (m *message) GetOneof(idx int32) *oneof {
	for _, oneof := range m.oneofs {
		if oneof.index == idx {
			return oneof
		}
	}
	return nil
}

type oneof struct {
	index int32
	*descriptor.OneofDescriptorProto
	fields []*field
}

func (o *oneof) GoString() string {
	return o.GetName()
}

type field struct {
	key      string
	comment  string
	repeated bool
	isOneOf  bool
	*descriptor.FieldDescriptorProto
}

func (f *field) GoString() string {
	var fieldInfo string
	if f.repeated {
		fieldInfo += "repeated "
	}
	typ := strings.ToLower(strings.TrimPrefix(f.GetType().String(), "TYPE_"))
	if typ == "message" {
		fieldInfo += f.GetTypeName()
	} else {
		fieldInfo += typ
	}
	return fmt.Sprintf("%s (%s)", f.key, fieldInfo)
}

type enumValue struct {
	comment string
	*descriptor.EnumValueDescriptorProto
}

func (e *enumValue) GoString() string {
	return e.GetName()
}

type enum struct {
	key     string
	comment string
	*descriptor.EnumDescriptorProto
	values []*enumValue
}

func (e *enum) GoString() string {
	return e.key
}

type tree struct {
	services map[string]*service
	methods  map[string]*method
	messages map[string]*message
	fields   map[string]*field
	enums    map[string]*enum
}
