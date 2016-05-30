package application

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func getTestApplication() (application *Application, dmap map[string]string) {
	return &Application{
			AppEUI:            types.AppEUI{8, 7, 6, 5, 4, 3, 2, 1},
			HandlerID:         "handlerID",
			HandlerNetAddress: "handlerNetAddress:1234",
		}, map[string]string{
			"app_eui":             "0807060504030201",
			"handler_id":          "handlerID",
			"handler_net_address": "handlerNetAddress:1234",
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
