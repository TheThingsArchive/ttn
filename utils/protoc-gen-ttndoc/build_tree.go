// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strconv"
	"strings"

	protobuf "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/genproto/googleapis/api/annotations"
)

func buildTree(files []*descriptor.FileDescriptorProto) *tree {
	tree := &tree{
		services: make(map[string]*service),
		methods:  make(map[string]*method),
		messages: make(map[string]*message),
		fields:   make(map[string]*field),
		enums:    make(map[string]*enum),
	}
	for _, file := range files {
		fillTreeWithFile(tree, file)
	}
	return tree
}

func fillTreeWithFile(tree *tree, file *descriptor.FileDescriptorProto) {
	key := fmt.Sprintf(".%s", file.GetPackage())
	locs := make(map[string]*descriptor.SourceCodeInfo_Location)
	for _, loc := range file.GetSourceCodeInfo().GetLocation() {
		if loc.LeadingComments == nil {
			continue
		}
		var p []string
		for _, n := range loc.Path {
			p = append(p, strconv.Itoa(int(n)))
		}
		locs[strings.Join(p, ",")] = loc
	}

	// Messages
	for idx, proto := range file.GetMessageType() {
		fillTreeWithMessage(tree, key, proto, fmt.Sprintf("4,%d", idx), locs)
	}

	// Enums
	for idx, proto := range file.GetEnumType() {
		fillTreeWithEnum(tree, key, proto, fmt.Sprintf("5,%d", idx), locs)
	}

	// Services
	for idx, proto := range file.GetService() {
		fillTreeWithService(tree, key, proto, fmt.Sprintf("6,%d", idx), locs)
	}
}

func fillTreeWithService(tree *tree, key string, proto *descriptor.ServiceDescriptorProto, loc string, locs map[string]*descriptor.SourceCodeInfo_Location) *service {
	key = fmt.Sprintf("%s.%s", key, proto.GetName())
	tree.services[key] = &service{key: key, comment: getComment(loc, locs), ServiceDescriptorProto: proto}

	// Methods
	for idx, proto := range proto.GetMethod() {
		method := fillTreeWithMethod(tree, key, proto, fmt.Sprintf("%s,2,%d", loc, idx), locs)
		tree.services[key].methods = append(tree.services[key].methods, method)
	}

	return tree.services[key]
}

func fillTreeWithMethod(tree *tree, key string, proto *descriptor.MethodDescriptorProto, loc string, locs map[string]*descriptor.SourceCodeInfo_Location) *method {
	key = fmt.Sprintf("%s.%s", key, proto.GetName())
	tree.methods[key] = &method{key: key, comment: getComment(loc, locs), MethodDescriptorProto: proto}
	if input, ok := tree.messages[proto.GetInputType()]; ok {
		tree.methods[key].input = input
	}
	if proto.GetClientStreaming() {
		tree.methods[key].inputStream = true
	}
	if output, ok := tree.messages[proto.GetOutputType()]; ok {
		tree.methods[key].output = output
	}
	if proto.GetServerStreaming() {
		tree.methods[key].outputStream = true
	}
	if proto.Options != nil && protobuf.HasExtension(proto.Options, annotations.E_Http) {
		ext, err := protobuf.GetExtension(proto.Options, annotations.E_Http)
		if err == nil {
			if opts, ok := ext.(*annotations.HttpRule); ok {
				if endpoint := newEndpoint(opts); endpoint != nil {
					tree.methods[key].endpoints = append(tree.methods[key].endpoints, endpoint)
				}
				for _, opts := range opts.AdditionalBindings {
					if endpoint := newEndpoint(opts); endpoint != nil {
						tree.methods[key].endpoints = append(tree.methods[key].endpoints, endpoint)
					}
				}
			}
		}
	}
	return tree.methods[key]
}

func fillTreeWithMessage(tree *tree, key string, proto *descriptor.DescriptorProto, loc string, locs map[string]*descriptor.SourceCodeInfo_Location) *message {
	key = fmt.Sprintf("%s.%s", key, proto.GetName())
	tree.messages[key] = &message{key: key, comment: getComment(loc, locs), DescriptorProto: proto}

	// Oneofs
	for idx, proto := range proto.GetOneofDecl() {
		tree.messages[key].oneofs = append(tree.messages[key].oneofs, &oneof{
			index:                int32(idx),
			OneofDescriptorProto: proto,
		})
	}

	// Fields
	for idx, proto := range proto.GetField() {
		field := fillTreeWithField(tree, key, proto, fmt.Sprintf("%s,2,%d", loc, idx), locs)
		tree.messages[key].fields = append(tree.messages[key].fields, field)
	}

	// Nested
	for idx, proto := range proto.GetNestedType() {
		message := fillTreeWithMessage(tree, key, proto, fmt.Sprintf("%s,3,%d", loc, idx), locs)
		tree.messages[key].nested = append(tree.messages[key].nested, message)
	}

	// Enums
	for idx, proto := range proto.GetEnumType() {
		fillTreeWithEnum(tree, key, proto, fmt.Sprintf("%s,4,%d", loc, idx), locs)
	}

	return tree.messages[key]
}

func fillTreeWithField(tree *tree, parent string, proto *descriptor.FieldDescriptorProto, loc string, locs map[string]*descriptor.SourceCodeInfo_Location) *field {
	key := fmt.Sprintf("%s.%s", parent, proto.GetName())
	tree.fields[key] = &field{key: key, comment: getComment(loc, locs), FieldDescriptorProto: proto}
	if proto.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		tree.fields[key].repeated = true
	}
	if proto.OneofIndex != nil {
		if parent, ok := tree.messages[parent]; ok {
			for _, oneof := range parent.oneofs {
				if oneof.index == proto.GetOneofIndex() {
					oneof.fields = append(oneof.fields, tree.fields[key])
					tree.fields[key].isOneOf = true
				}
			}
		}
	}
	return tree.fields[key]
}

func fillTreeWithEnum(tree *tree, key string, proto *descriptor.EnumDescriptorProto, loc string, locs map[string]*descriptor.SourceCodeInfo_Location) *enum {
	key = fmt.Sprintf("%s.%s", key, proto.GetName())

	tree.enums[key] = &enum{key: key, comment: getComment(loc, locs), EnumDescriptorProto: proto}

	// Values
	for idx, proto := range proto.GetValue() {
		tree.enums[key].values = append(tree.enums[key].values, &enumValue{
			getComment(fmt.Sprintf("%s,2,%d", loc, idx), locs),
			proto,
		})
	}

	return tree.enums[key]
}

func getComment(loc string, locs map[string]*descriptor.SourceCodeInfo_Location) (comment string) {
	if loc, ok := locs[loc]; ok {
		var lines []string
		for _, line := range strings.Split(strings.TrimSuffix(loc.GetLeadingComments(), "\n"), "\n") {
			line = strings.TrimPrefix(line, " ")
			line = strings.Replace(line, "```", "", -1)
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
	}
	return ""
}
