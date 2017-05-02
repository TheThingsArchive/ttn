// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttndoc

import (
	"strconv"
	"strings"

	"unicode"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

type File struct {
	TTNDoc     *TTNDoc
	Name       string
	Path       string
	Descriptor *descriptor.FileDescriptorProto
	Package    *Package
	Messages   map[string]*Message
	Services   map[string]*Service
	Enums      map[string]*Enum
	document   *bool
	main       bool
}

func (f File) Document() bool {
	if f.document != nil {
		return *f.document
	}
	for _, message := range f.Messages {
		if message.Document() {
			return true
		}
	}
	for _, service := range f.Services {
		if service.Document() {
			return true
		}
	}
	for _, enum := range f.Enums {
		if enum.Document() {
			return true
		}
	}
	return false
}

func (d *TTNDoc) AddFile(msg *descriptor.FileDescriptorProto) {
	file := newFile(msg.GetPackage(), msg)
	file.TTNDoc = d
	file.Path = msg.GetName()
	for _, filter := range d.FilterFiles {
		if file.Path == filter {
			file.main = true
			break
		}
	}
	d.Files[msg.GetName()] = file
	pk, ok := d.Packages[msg.GetPackage()]
	if !ok {
		pk = d.newPackage(msg.GetPackage())
	}
	pk.AddFile(file)
	file.Package = pk
	d.Packages[msg.GetPackage()] = pk
	file.Enter()
}

func newFile(name string, msg *descriptor.FileDescriptorProto) *File {
	f := &File{
		Name:       "." + name,
		Descriptor: msg,
		Messages:   make(map[string]*Message),
		Services:   make(map[string]*Service),
		Enums:      make(map[string]*Enum),
	}
	return f
}

func (f *File) Enter() {
	if f.main { // Don't consider documentation options for dependency protos
		if opts := f.Descriptor.GetOptions(); opts != nil && proto.HasExtension(opts, E_DocumentFile) {
			if ext, err := proto.GetExtension(opts, E_DocumentFile); err == nil {
				f.document = ext.(*bool)
			} else {
				dbg(err)
			}
		}
	}

	// Comments
	for _, loc := range f.Descriptor.GetSourceCodeInfo().GetLocation() {
		comment := loc.GetLeadingComments()
		if comment == "" {
			continue
		}
		var p []string
		for _, n := range loc.Path {
			p = append(p, strconv.Itoa(int(n)))
		}
		var cleaned string
		for _, line := range strings.Split(comment, "\n") {
			cleaned += strings.TrimRightFunc(strings.TrimPrefix(line, " "), unicode.IsSpace) + "\n"
		}
		f.TTNDoc.Comments[f.Name+":"+strings.Join(p, ",")] = strings.TrimRightFunc(cleaned, unicode.IsSpace)
	}

	// Enums (type 5)
	for idx, msg := range f.Descriptor.GetEnumType() {
		name := f.Name + "." + msg.GetName()
		enum := newEnum(f.TTNDoc.newProto(name, newLoc("", 5, idx)), msg)
		enum.document = f.document
		f.Enums[name] = enum
		enum.File = f
		enum.Enter()
	}

	// Messages (type 4)
	for idx, msg := range f.Descriptor.GetMessageType() {
		name := f.Name + "." + msg.GetName()
		message := newMessage(f.TTNDoc.newProto(name, newLoc("", 4, idx)), msg)
		message.document = f.document
		f.Messages[name] = message
		message.File = f
		message.Enter()
	}

	// Services (type 6)
	for idx, msg := range f.Descriptor.GetService() {
		name := f.Name + "." + msg.GetName()
		service := newService(f.TTNDoc.newProto(name, newLoc("", 6, idx)), msg)
		service.document = f.document
		f.Services[name] = service
		service.File = f
		service.Enter()
	}
	return

}
