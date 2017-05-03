// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

type Service struct {
	Proto
	Descriptor *descriptor.ServiceDescriptorProto
	Methods    map[string]*Method
	document   *bool
}

func (s Service) Document() bool {
	if s.document != nil {
		return *s.document
	}
	for _, method := range s.Methods {
		if method.Document() {
			return true
		}
	}
	return false
}

func newService(proto Proto, msg *descriptor.ServiceDescriptorProto) *Service {
	s := &Service{
		Proto:      proto,
		Descriptor: msg,
		Methods:    make(map[string]*Method),
	}
	s.TTNDoc.Services[s.Name] = s
	return s
}

func (s *Service) Enter() {
	if s.File.main { // Don't consider documentation options for dependency protos
		if opts := s.Descriptor.GetOptions(); opts != nil && proto.HasExtension(opts, E_DocumentService) {
			if ext, err := proto.GetExtension(opts, E_DocumentService); err == nil && s.document == nil {
				s.document = ext.(*bool)
			}
		}
	}

	// Methods (type 2)
	for idx, msg := range s.Descriptor.GetMethod() {
		name := s.Name + "." + msg.GetName()
		method := newMethod(s.TTNDoc.newProto(name, newLoc(s.Loc, 2, idx)), msg)
		method.document = s.document
		s.Methods[name] = method
		method.Service = s
		method.File = s.File
		method.Enter()
	}
}
