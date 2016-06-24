// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package application

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func getTestApplication() (application *Application, dmap map[string]string) {
	return &Application{
			AppID:     "AppID-1",
			Decoder:   `function (payload) { return { size: payload.length; } }`,
			Converter: `function (data) { return data; }`,
			Validator: `function (data) { return data.size % 2 == 0; }`,
		}, map[string]string{
			"app_id":    "AppID-1",
			"decoder":   `function (payload) { return { size: payload.length; } }`,
			"converter": `function (data) { return data; }`,
			"validator": `function (data) { return data.size % 2 == 0; }`,
		}
}

func TestToStringMap(t *testing.T) {
	a := New(t)
	application, expected := getTestApplication()
	dmap, err := application.ToStringStringMap(ApplicationProperties...)
	a.So(err, ShouldBeNil)
	a.So(dmap, ShouldResemble, expected)
}

func TestFromStringMap(t *testing.T) {
	a := New(t)
	application := &Application{}
	expected, dmap := getTestApplication()
	err := application.FromStringStringMap(dmap)
	a.So(err, ShouldBeNil)
	a.So(application, ShouldResemble, expected)
}
