// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

type Enum struct {
	Proto
	Descriptor *descriptor.EnumDescriptorProto
	Values     map[string]*EnumValue // 2
	UsedIn     []*Message
	document   *bool
}

func newEnum(proto Proto, msg *descriptor.EnumDescriptorProto) *Enum {
	e := &Enum{
		Proto:      proto,
		Descriptor: msg,
		Values:     make(map[string]*EnumValue),
	}
	e.TTNDoc.Enums[e.Name] = e
	return e
}

func (e Enum) Document() bool {
	if e.document != nil {
		return *e.document
	}
	for _, message := range e.UsedIn {
		if message.Document() {
			return true
		}
	}
	return false
}

func (e Enum) Default() interface{} {
	for _, value := range e.Values {
		return value.Default()
	}
	return ""
}

func (e *Enum) Enter() {
	if e.File.main { // Don't consider documentation options for dependency protos
		if opts := e.Descriptor.GetOptions(); opts != nil && proto.HasExtension(opts, E_DocumentEnum) {
			if ext, err := proto.GetExtension(opts, E_DocumentEnum); err == nil && e.document == nil {
				e.document = ext.(*bool)
			}
		}
	}

	// Values (type 2)
	for idx, msg := range e.Descriptor.GetValue() {
		name := e.Name + "." + msg.GetName()
		value := newEnumValue(e.TTNDoc.newProto(name, newLoc(e.Loc, 2, idx)), msg)
		value.Enum = e
		value.File = e.File
		value.Enter()
		e.Values[name] = value
	}
}

type EnumValue struct {
	Proto
	Descriptor *descriptor.EnumValueDescriptorProto
	Enum       *Enum
}

func newEnumValue(proto Proto, msg *descriptor.EnumValueDescriptorProto) *EnumValue {
	return &EnumValue{
		Proto:      proto,
		Descriptor: msg,
	}
}

func (v *EnumValue) Default() interface{} {
	return v.Descriptor.GetName()
}

func (v *EnumValue) Enter() {

}
