// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

type Field struct {
	Proto
	Descriptor *descriptor.FieldDescriptorProto
	Message    *Message
	Oneof      *Oneof
	Type       string
	Repeated   bool
}

func newField(proto Proto, msg *descriptor.FieldDescriptorProto) *Field {
	f := &Field{
		Proto:      proto,
		Descriptor: msg,
	}
	f.TTNDoc.Fields[f.Name] = f
	return f
}

func (f Field) Document() bool {
	return f.Message.Document()
}

func (f Field) Default() interface{} {
	if exampleValue, ok := exampleValues[f.Name]; ok {
		return exampleValue
	}

	if exampleValue, ok := exampleValues["."+f.File.Package.Name+".*."+f.Descriptor.GetName()]; ok {
		return exampleValue
	}

	switch f.Descriptor.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return float64(0)
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return float32(0)
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return int64(0)
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return uint64(0)
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return int32(0)
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return uint64(0)
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return uint32(0)
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return false
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return ""
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		return nil
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if message, ok := f.TTNDoc.Messages[f.Descriptor.GetTypeName()]; ok {
			return message.Default()
		}
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return []byte{}
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return uint32(0)
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		if enum, ok := f.TTNDoc.Enums[f.Descriptor.GetTypeName()]; ok {
			return enum.Default()
		}
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		return int32(0)
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return int64(0)
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		return int32(0)
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		return int64(0)
	}
	return nil
}

func (f *Field) Enter() {
	f.Repeated = f.Descriptor.IsRepeated()
	switch f.Descriptor.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		f.Type = "double"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		f.Type = "float"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		f.Type = "int64"
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		f.Type = "uint64"
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		f.Type = "int32"
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		f.Type = "fixed64"
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		f.Type = "fixed32"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		f.Type = "bool"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		f.Type = "string"
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		f.Type = "group"
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		f.Type = f.Descriptor.GetTypeName()
		if message, ok := f.TTNDoc.Messages[f.Type]; ok {
			message.UsedInMessages = append(message.UsedInMessages, f.Message)
		}
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		f.Type = "bytes"
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		f.Type = "uint32"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		f.Type = f.Descriptor.GetTypeName()
		if enum, ok := f.TTNDoc.Enums[f.Type]; ok {
			enum.UsedIn = append(enum.UsedIn, f.Message)
		}
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		f.Type = "sfixed32"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		f.Type = "sfixed64"
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		f.Type = "sint32"
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		f.Type = "sint64"
	}
}
