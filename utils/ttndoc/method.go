// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import (
	"github.com/TheThingsNetwork/ttn/utils/ttndoc/internal/annotations"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

type MethodType struct {
	Message *Message
	Stream  bool
}

type Method struct {
	Proto
	Descriptor    *descriptor.MethodDescriptorProto
	Service       *Service
	Input         MethodType
	Output        MethodType
	HTTPEndpoints []*HTTPEndpoint
	document      *bool
}

func newMethod(proto Proto, msg *descriptor.MethodDescriptorProto) *Method {
	m := &Method{
		Proto:      proto,
		Descriptor: msg,
	}
	m.TTNDoc.Methods[m.Name] = m
	return m
}

func (m Method) Document() bool {
	if m.document != nil {
		return *m.document
	}
	return false
}

func (m *Method) Enter() {
	if m.File.main { // Don't consider documentation options for dependency protos
		if opts := m.Descriptor.GetOptions(); opts != nil && proto.HasExtension(opts, E_DocumentMethod) {
			if ext, err := proto.GetExtension(opts, E_DocumentMethod); err == nil && m.document == nil {
				m.document = ext.(*bool)
			}
		}
	}

	m.Input.Message = m.TTNDoc.Messages[m.Descriptor.GetInputType()]
	m.Input.Message.UsedInMethods = append(m.Input.Message.UsedInMethods, m)
	m.Input.Stream = m.Descriptor.GetClientStreaming()

	m.Output.Message = m.TTNDoc.Messages[m.Descriptor.GetOutputType()]
	m.Output.Message.UsedInMethods = append(m.Output.Message.UsedInMethods, m)
	m.Output.Stream = m.Descriptor.GetServerStreaming()

	if opts := m.Descriptor.GetOptions(); opts != nil && proto.HasExtension(opts, annotations.E_Http) {
		if ext, err := proto.GetExtension(opts, annotations.E_Http); err == nil {
			httpEndpoint := newHTTPEndpoint(ext.(*annotations.HttpRule))
			httpEndpoint.Method = m
			httpEndpoint.Enter()
			m.HTTPEndpoints = append(m.HTTPEndpoints, httpEndpoint)
		}
	}

}
