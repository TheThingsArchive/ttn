// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"strings"
)

var exampleValues = map[string]interface{}{
	".discovery.*.id":                         "ttn-handler-eu",
	".discovery.*.service_name":               "handler",
	".discovery.Announcement.api_address":     "http://eu.thethings.network:8084",
	".discovery.Announcement.certificate":     "-----BEGIN CERTIFICATE-----\n...",
	".discovery.Announcement.net_address":     "eu.thethings.network:1904",
	".discovery.Announcement.public_key":      "-----BEGIN PUBLIC KEY-----\n...",
	".discovery.Announcement.public":          true,
	".discovery.Announcement.service_version": "2.0.0-abcdef...",
	".discovery.Metadata.dev_addr_prefix":     "AAAAAAA=",
	".discovery.Metadata.app_id":              "some-app-id",
	".handler.*.app_id":                       "some-app-id",
	".handler.*.dev_id":                       "some-dev-id",
	".handler.*.fields":                       `{"light":100}`,
	".handler.*.payload":                      "ZA==",
	".handler.*.port":                         1,
	".handler.Application.converter":          "function Converter(decoded, port) {...",
	".handler.Application.decoder":            "function Decoder(bytes, port) {...",
	".handler.Application.encoder":            "Encoder(object, port) {...",
	".handler.Application.validator":          "Validator(converted, port) {...",
	".handler.DryDownlinkMessage.payload":     "",
	".handler.LogEntry.fields":                `["TTN",123]`,
	".handler.LogEntry.function":              "decoder",
	".handler.Device.description":             "Some description of the device",
	".handler.Device.latitude":                52.375,
	".handler.Device.longitude":               4.887,
	".lorawan.Device.activation_constraints":  "local",
	".lorawan.Device.app_eui":                 "0102030405060708",
	".lorawan.Device.app_id":                  "some-app-id",
	".lorawan.Device.app_key":                 "01020304050607080102030405060708",
	".lorawan.Device.app_s_key":               "01020304050607080102030405060708",
	".lorawan.Device.dev_addr":                "01020304",
	".lorawan.Device.dev_eui":                 "0102030405060708",
	".lorawan.Device.dev_id":                  "some-dev-id",
	".lorawan.Device.nwk_s_key":               "01020304050607080102030405060708",
	".lorawan.Device.uses32_bit_f_cnt":        true,
}

func (m *message) MapExample(tree *tree) map[string]interface{} {
	example := make(map[string]interface{})
	for _, field := range m.fields {
		typ := strings.ToLower(strings.TrimPrefix(field.GetType().String(), "TYPE_"))
		var val interface{}

		if exampleValue, ok := exampleValues[field.key]; ok {
			val = exampleValue
		} else if exampleValue, ok := exampleValues[field.key[:strings.Index(field.key[1:], ".")+1]+".*."+field.GetName()]; ok {
			val = exampleValue
		} else {
			switch typ {
			case "message":
				if message, ok := tree.messages[field.GetTypeName()]; ok {
					val = message.MapExample(tree)
				}
			case "enum":
				if enums, ok := tree.enums[field.GetTypeName()]; ok {
					val = enums.MapExample(tree)
				}
			case "string":
				val = ""
			case "bool":
				val = false
			case "bytes":
				val = ""
			case "int64", "int32", "uint64", "uint32", "sint64", "sint32", "fixed64", "fixed32", "sfixed32", "sfixed64":
				val = 0
			case "double", "float":
				val = 0.0
			default:
			}
		}
		if field.repeated {
			example[field.GetName()] = []interface{}{val}
		} else {
			example[field.GetName()] = val
		}
	}
	return example
}

func (m *enum) MapExample(tree *tree) string {
	if len(m.values) == 0 {
		return ""
	}
	return m.values[len(m.values)-1].GetName()
}

func (m *message) JSONExample(tree *tree) string {
	example := m.MapExample(tree)
	exampleBytes, _ := json.MarshalIndent(example, "", "  ")
	return string(exampleBytes)
}
