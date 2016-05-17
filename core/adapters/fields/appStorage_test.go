// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fields

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func createStorage() AppStorage {
	storage, err := ConnectRedis("localhost:6379", 0)
	if err != nil {
		panic(err)
	}
	return storage
}

func TestConnect(t *testing.T) {
	a := New(t)

	c, err := ConnectRedis("localhost:6379", 0)
	a.So(err, ShouldBeNil)
	defer c.Close()

	_, err = ConnectRedis("", 0)
	a.So(err, ShouldNotBeNil)
}

func TestGetFunctions(t *testing.T) {
	a := New(t)

	eui, _ := types.ParseAppEUI("8000000000000001")

	storage := createStorage()
	defer storage.Close()

	fetchedFunctions, err := storage.GetFunctions(eui)
	a.So(err, ShouldBeNil)
	a.So(fetchedFunctions, ShouldBeNil)
}

func TestSetFunctions(t *testing.T) {
	a := New(t)

	eui, _ := types.ParseAppEUI("8000000000000001")
	functions := &Functions{
		Decoder:   `function (payload) { return { size: payload.length; } }`,
		Converter: `function (data) { return data; }`,
		Validator: `function (data) { return data.size % 2 == 0; }`,
	}

	storage := createStorage()
	defer storage.Close()
	defer storage.Reset()

	err := storage.SetFunctions(eui, functions)
	a.So(err, ShouldBeNil)

	fetchedFunctions, err := storage.GetFunctions(eui)
	a.So(err, ShouldBeNil)
	a.So(fetchedFunctions.Decoder, ShouldEqual, functions.Decoder)
	a.So(fetchedFunctions.Converter, ShouldEqual, functions.Converter)
	a.So(fetchedFunctions.Validator, ShouldEqual, functions.Validator)
}
