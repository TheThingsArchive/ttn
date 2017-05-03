// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

type Oneof struct {
	Proto
	Descriptor *descriptor.OneofDescriptorProto
	Message    *Message
	Fields     []*Field
}

func newOneof(proto Proto, msg *descriptor.OneofDescriptorProto) *Oneof {
	return &Oneof{
		Proto:      proto,
		Descriptor: msg,
	}
}

func (o Oneof) Document() bool {
	return o.Message.Document()
}

func (o *Oneof) Enter() {}
