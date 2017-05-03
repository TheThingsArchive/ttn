// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"strings"

	"github.com/TheThingsNetwork/ttn/utils/ttndoc"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	plugin "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/gogo/protobuf/vanity"
	"github.com/gogo/protobuf/vanity/command"
)

func main() {
	req := command.Read()
	files := req.GetProtoFile()
	files = vanity.FilterFiles(files, vanity.NotGoogleProtobufDescriptorProto)

	for _, opt := range []func(*descriptor.FileDescriptorProto){
		//vanity.TurnOffGoGettersAll,
		//vanity.TurnOffGoEnumPrefixAll,
		vanity.TurnOffGoStringerAll,
		vanity.TurnOnVerboseEqualAll,
		//vanity.TurnOnFaceAll,
		//vanity.TurnOnGoStringAll,
		//vanity.TurnOnPopulateAll,
		vanity.TurnOnStringerAll,
		vanity.TurnOnEqualAll,
		//vanity.TurnOnDescriptionAll,
		//vanity.TurnOnTestGenAll,
		//vanity.TurnOnBenchGenAll,
		vanity.TurnOnMarshalerAll,
		vanity.TurnOnUnmarshalerAll,
		//vanity.TurnOnStable_MarshalerAll,
		vanity.TurnOnSizerAll,
		//vanity.TurnOffGoEnumStringerAll,
		//vanity.TurnOnEnumStringerAll,
		//vanity.TurnOnUnsafeUnmarshalerAll,
		//vanity.TurnOnUnsafeMarshalerAll,
		//vanity.TurnOffGoExtensionsMapAll,
		vanity.TurnOffGoUnrecognizedAll,
		//vanity.TurnOffGogoImport,
		//vanity.TurnOnCompareAll,
	} {
		vanity.ForEachFile(files, opt)
	}

	doc := ttndoc.New()
	doc.FilterFiles = req.FileToGenerate

	vanity.ForEachFile(files, func(msg *descriptor.FileDescriptorProto) {
		doc.AddFile(msg)
	})

	out := command.Generate(req)

	doc = doc.FilterDocumented()

	file := req.FileToGenerate[0]
	dir := file[:strings.LastIndex(file, "/")]
	fileBase := strings.Split(strings.TrimPrefix(file, dir), ".")[0]
	outputFile := filepath.Join(dir, fileBase+".md")
	if len(req.FileToGenerate) > 1 {
		outputFile = filepath.Join(dir, "APIDOC.md")
	}
	if len(doc.Services) == 1 {
		for service := range doc.Services { // It's actually just 1
			outputFile = filepath.Join(dir, service[strings.LastIndex(service, ".")+1:]+".md")
		}
	}

	md, err := doc.Render()
	if err != nil {
		errStr := err.Error()
		out.Error = &errStr
	} else if md != "" {
		out.File = append(out.File, &plugin.CodeGeneratorResponse_File{
			Name:    &outputFile,
			Content: &md,
		})
	}

	command.Write(out)
}
