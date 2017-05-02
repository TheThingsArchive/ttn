// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
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
	command.Write(command.Generate(req))
}
