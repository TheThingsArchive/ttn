// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

//go:generate protoc -I=. --gogo_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:$GOPATH/src ./ttndoc.proto

package ttndoc

import (
	"fmt"
	"os"
)

type TTNDoc struct {
	FilterFiles []string
	Packages    map[string]*Package
	Files       map[string]*File
	Messages    map[string]*Message
	Services    map[string]*Service
	Methods     map[string]*Method
	Fields      map[string]*Field
	Enums       map[string]*Enum
	Comments    map[string]string
}

func New() *TTNDoc {
	return &TTNDoc{
		Packages: make(map[string]*Package),
		Files:    make(map[string]*File),
		Messages: make(map[string]*Message),
		Services: make(map[string]*Service),
		Methods:  make(map[string]*Method),
		Fields:   make(map[string]*Field),
		Enums:    make(map[string]*Enum),
		Comments: make(map[string]string),
	}
}

func (d *TTNDoc) GetComment(loc string) string {
	if comment, ok := d.Comments[loc]; ok {
		return comment
	}
	return ""
}

func (d *TTNDoc) FilterDocumented() *TTNDoc {
	out := New()
	out.Comments = d.Comments
	for k, v := range d.Packages {
		if v.Document() {
			out.Packages[k] = v
		}
	}
	for k, v := range d.Files {
		if v.Document() {
			out.Files[k] = v
		}
	}
	for k, v := range d.Messages {
		if v.Document() {
			out.Messages[k] = v
		}
	}
	for k, v := range d.Services {
		if v.Document() {
			out.Services[k] = v
		}
	}
	for k, v := range d.Methods {
		if v.Document() {
			out.Methods[k] = v
		}
	}
	for k, v := range d.Fields {
		if v.Document() {
			out.Fields[k] = v
		}
	}
	for k, v := range d.Enums {
		if v.Document() {
			out.Enums[k] = v
		}
	}
	return out
}

func newLoc(parent string, t int, idx int) (out string) {
	if parent != "" {
		out = parent + ","
	}
	out += fmt.Sprintf("%d,%d", t, idx)
	return
}

func dbg(a ...interface{}) {
	if len(a) == 0 {
		return
	}
	format, ok := a[0].(string)
	if len(a) < 2 || !ok {
		fmt.Fprintln(os.Stderr, a...)
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", a[1:]...)
}
