// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestDecodeToType(t *testing.T) {
	a := New(t)
	a.So(decodeToType(reflect.String, "abc"), ShouldEqual, "abc")
	a.So(decodeToType(reflect.Bool, "true"), ShouldEqual, true)
	a.So(decodeToType(reflect.Bool, "false"), ShouldEqual, false)
	a.So(decodeToType(reflect.Int, "-10"), ShouldEqual, -10)
	a.So(decodeToType(reflect.Int8, "-10"), ShouldEqual, -10)
	a.So(decodeToType(reflect.Int16, "-10"), ShouldEqual, -10)
	a.So(decodeToType(reflect.Int32, "-10"), ShouldEqual, -10)
	a.So(decodeToType(reflect.Int64, "-10"), ShouldEqual, -10)
	a.So(decodeToType(reflect.Uint, "10"), ShouldEqual, 10)
	a.So(decodeToType(reflect.Uint8, "10"), ShouldEqual, 10)
	a.So(decodeToType(reflect.Uint16, "10"), ShouldEqual, 10)
	a.So(decodeToType(reflect.Uint32, "10"), ShouldEqual, 10)
	a.So(decodeToType(reflect.Uint64, "10"), ShouldEqual, 10)
	a.So(decodeToType(reflect.Float64, "12.34"), ShouldEqual, 12.34)
	a.So(decodeToType(reflect.Float32, "12.34"), ShouldEqual, 12.34)
	a.So(decodeToType(reflect.Struct, "blabla"), ShouldBeNil)
}

type noIdea struct {
	something complex128
}

type strtp string

type testTextUnmarshaler struct {
	Data string
}

func (um *testTextUnmarshaler) UnmarshalText(text []byte) error {
	um.Data = string(text)
	return nil
}

type testCustomType [2]byte

type testJSONUnmarshaler struct {
	Customs []testCustomType
}

type jsonStruct struct {
	String string `json:"string"`
}

func (jum *testJSONUnmarshaler) UnmarshalJSON(text []byte) error {
	jum.Customs = []testCustomType{}
	txt := strings.Trim(string(text), "[]")
	txtlist := strings.Split(txt, ",")
	for _, txtitem := range txtlist {
		txtitem = strings.Trim(txtitem, `"`)
		b, _ := hex.DecodeString(txtitem)
		var o testCustomType
		copy(o[:], b[:])
		jum.Customs = append(jum.Customs, o)
	}
	return nil
}

func TestUnmarshalToType(t *testing.T) {
	a := New(t)

	var str string
	strOut, err := unmarshalToType(reflect.TypeOf(str), "data")
	a.So(err, ShouldBeNil)
	a.So(strOut, ShouldEqual, "data")

	var strtp strtp
	strtpOut, err := unmarshalToType(reflect.TypeOf(strtp), "data")
	a.So(err, ShouldBeNil)
	a.So(strtpOut, ShouldEqual, "data")

	var um testTextUnmarshaler
	umOut, err := unmarshalToType(reflect.TypeOf(um), "data")
	a.So(err, ShouldBeNil)
	a.So(umOut.(testTextUnmarshaler), ShouldResemble, testTextUnmarshaler{"data"})

	var jum testJSONUnmarshaler
	jumOut, err := unmarshalToType(reflect.TypeOf(jum), `["abcd","1234","def0"]`)
	a.So(err, ShouldBeNil)
	a.So(jumOut.(testJSONUnmarshaler), ShouldResemble, testJSONUnmarshaler{[]testCustomType{
		testCustomType{0xab, 0xcd},
		testCustomType{0x12, 0x34},
		testCustomType{0xde, 0xf0},
	}})

	var js jsonStruct
	jsOut, err := unmarshalToType(reflect.TypeOf(js), `{"string": "String"}`)
	a.So(err, ShouldBeNil)
	a.So(jsOut.(jsonStruct), ShouldResemble, jsonStruct{"String"})

	_, err = unmarshalToType(reflect.TypeOf(js), `this is no json`)
	a.So(err, ShouldNotBeNil)

	var eui types.DevEUI
	euiOut, err := unmarshalToType(reflect.TypeOf(eui), "0102abcd0304abcd")
	a.So(err, ShouldBeNil)
	a.So(euiOut.(types.DevEUI), ShouldEqual, types.DevEUI{0x01, 0x02, 0xab, 0xcd, 0x03, 0x04, 0xab, 0xcd})

	var euiPtr *types.DevEUI
	euiPtrOut, err := unmarshalToType(reflect.TypeOf(euiPtr), "0102abcd0304abcd")
	a.So(err, ShouldBeNil)
	a.So(euiPtrOut.(*types.DevEUI), ShouldResemble, &types.DevEUI{0x01, 0x02, 0xab, 0xcd, 0x03, 0x04, 0xab, 0xcd})
}
