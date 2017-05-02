// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

type Message struct {
	Proto
	Descriptor     *descriptor.DescriptorProto
	Fields         map[string]*Field   // 2
	Messages       map[string]*Message // 3
	Enums          map[string]*Enum    // 4
	Oneofs         map[int]*Oneof      // 8
	UsedInMethods  []*Method
	UsedInMessages []*Message
	document       *bool
}

func newMessage(proto Proto, msg *descriptor.DescriptorProto) *Message {
	m := &Message{
		Proto:      proto,
		Descriptor: msg,
		Fields:     make(map[string]*Field),
		Messages:   make(map[string]*Message),
		Enums:      make(map[string]*Enum),
		Oneofs:     make(map[int]*Oneof),
	}
	m.TTNDoc.Messages[m.Name] = m
	return m
}

func (m *Message) Document() bool {
	if m.document != nil {
		return *m.document
	}
	for _, message := range m.UsedInMessages {
		if message == m {
			continue
		}
		// NOTE: This will go terribly wrong for indirect recursive messages, but we don't have those
		if message.Document() {
			return true
		}
	}
	for _, method := range m.UsedInMethods {
		if method.Document() {
			return true
		}
	}
	return false
}

func (m *Message) Default() interface{} {
	out := make(map[string]interface{})
	for _, field := range m.Fields {
		if field.Type == m.Name {
			continue
		}
		// NOTE: This will go terribly wrong for indirect recursive messages, but we don't have those
		out[strings.TrimPrefix(field.Name, m.Name+".")] = field.Default()
	}
	return out
}

func (m *Message) Enter() {
	if m.File.main { // Don't consider documentation options for dependency protos
		if opts := m.Descriptor.GetOptions(); opts != nil && proto.HasExtension(opts, E_DocumentMessage) {
			if ext, err := proto.GetExtension(opts, E_DocumentMessage); err == nil && m.document == nil {
				m.document = ext.(*bool)
			}
		}
	}

	// Enums (type 4)
	for idx, msg := range m.Descriptor.GetEnumType() {
		name := m.Name + "." + msg.GetName()
		enum := newEnum(m.TTNDoc.newProto(name, newLoc(m.Loc, 4, idx)), msg)
		enum.document = m.document
		m.Enums[name] = enum
		enum.File = m.File
		enum.Enter()
	}

	// Messages (type 3)
	for idx, msg := range m.Descriptor.GetNestedType() {
		name := m.Name + "." + msg.GetName()
		subMessage := newMessage(m.TTNDoc.newProto(name, newLoc(m.Loc, 3, idx)), msg)
		subMessage.document = m.document
		m.Messages[name] = subMessage
		subMessage.File = m.File
		subMessage.Enter()
	}

	// OneOfs (type 8) (!! should be parsed before Fields)
	for idx, msg := range m.Descriptor.GetOneofDecl() {
		name := m.Name + "." + msg.GetName()
		oneof := newOneof(m.TTNDoc.newProto(name, newLoc(m.Loc, 8, idx)), msg)
		m.Oneofs[idx] = oneof
		oneof.File = m.File
		oneof.Message = m
		oneof.Enter()
	}

	// Fields (type 2)
	for idx, msg := range m.Descriptor.GetField() {
		name := m.Name + "." + msg.GetName()
		field := newField(m.TTNDoc.newProto(name, newLoc(m.Loc, 2, idx)), msg)
		m.Fields[name] = field
		field.File = m.File
		field.Message = m
		if field.Descriptor.OneofIndex != nil {
			field.Oneof = m.Oneofs[int(*field.Descriptor.OneofIndex)]
			field.Oneof.Fields = append(field.Oneof.Fields, field)
		}
		field.Enter()
	}

}
