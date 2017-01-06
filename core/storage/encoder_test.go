// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

type strtps []strtp

type jm struct {
	txt string
}

func (j jm) MarshalJSON() ([]byte, error) {
	return []byte("{" + j.txt + "}"), nil
}

type testStruct struct {
	unexported     string        `redis:"unexported"`
	NoRedis        string        ``
	DisRedis       string        `redis:"-"`
	String         string        `redis:"string"`
	Strings        []string      `redis:"strings"`
	EmptyString    string        `redis:"empty_string,omitempty"`
	EmptyStrings   []string      `redis:"empty_strings,omitempty"`
	AppEUI         types.AppEUI  `redis:"app_eui"`
	AppEUIPtr      *types.AppEUI `redis:"app_eui_ptr"`
	EmptyAppEUI    types.AppEUI  `redis:"empty_app_eui,omitempty"`
	EmptyAppEUIPtr *types.AppEUI `redis:"empty_app_eui_ptr,omitempty"`
	Time           time.Time     `redis:"time"`
	TimePtr        *time.Time    `redis:"time_ptr"`
	EmptyTime      time.Time     `redis:"empty_time,omitempty"`
	EmptyTimePtr   *time.Time    `redis:"empty_time_ptr,omitempty"`
	STime          Time          `redis:"stime"`
	STimePtr       *Time         `redis:"stime_ptr"`
	EmptySTime     Time          `redis:"empty_stime,omitempty"`
	EmptySTimePtr  *Time         `redis:"empty_stime_ptr,omitempty"`
	Str            strtp         `redis:"str"`
	Strs           []strtp       `redis:"strs"`
	JM             jm            `redis:"jm"`
	JMs            []jm          `redis:"jms"`
	Int            int           `redis:"int"`
	Uint           uint          `redis:"uint"`
}

func TestDefaultStructEncoder(t *testing.T) {
	a := New(t)

	var emptyAppEUI types.AppEUI
	var appEUI = types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}
	var now = time.Now()
	var emptyTime time.Time
	var stime = Time{now}
	var emptySTime Time

	out, err := buildDefaultStructEncoder("")(&testStruct{
		unexported:     "noop",
		NoRedis:        "noop",
		DisRedis:       "noop",
		String:         "string",
		Strings:        []string{"string1", "string2"},
		AppEUI:         appEUI,
		AppEUIPtr:      &appEUI,
		EmptyAppEUIPtr: &emptyAppEUI,
		Time:           now,
		TimePtr:        &now,
		STime:          stime,
		STimePtr:       &stime,
		EmptySTimePtr:  &emptySTime,
		EmptyTimePtr:   &emptyTime,
		JM:             jm{"cool"},
	})
	a.So(err, ShouldBeNil)
	a.So(out, ShouldNotContainKey, "unexported")
	a.So(out, ShouldNotContainKey, "empty_string")
	a.So(out, ShouldNotContainKey, "empty_strings")
	a.So(out, ShouldNotContainKey, "empty_app_eui")
	a.So(out, ShouldNotContainKey, "empty_app_eui_ptr")
	a.So(out, ShouldNotContainKey, "empty_time")
	a.So(out, ShouldNotContainKey, "empty_time_ptr")
	a.So(out, ShouldNotContainKey, "empty_stime")
	a.So(out, ShouldNotContainKey, "empty_stime_ptr")
	a.So(out["string"], ShouldEqual, "string")
	a.So(out["strings"], ShouldEqual, `["string1","string2"]`)
	a.So(out["jm"], ShouldEqual, "{cool}")

	out, err = buildDefaultStructEncoder("")(&testStruct{
		String:  "noop",
		Strings: []string{"string1", "string2"},
	}, "String")
	a.So(err, ShouldBeNil)
	a.So(out, ShouldNotContainKey, "strings")
}
