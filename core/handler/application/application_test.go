package application

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func getTestApplication() (application *Application, dmap map[string]string) {
	return &Application{
			AppEUI:        types.AppEUI{8, 7, 6, 5, 4, 3, 2, 1},
			DefaultAppKey: types.AppKey{8, 7, 6, 5, 4, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1},
			Decoder:       `function (payload) { return { size: payload.length; } }`,
			Converter:     `function (data) { return data; }`,
			Validator:     `function (data) { return data.size % 2 == 0; }`,
		}, map[string]string{
			"app_eui":         "0807060504030201",
			"default_app_key": "08070605040302010807060504030201",
			"decoder":         `function (payload) { return { size: payload.length; } }`,
			"converter":       `function (data) { return data; }`,
			"validator":       `function (data) { return data.size % 2 == 0; }`,
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
