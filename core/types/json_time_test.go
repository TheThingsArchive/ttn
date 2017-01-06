// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/assertions"
)

func TestJSONTime(t *testing.T) {
	a := New(t)

	a.So(BuildTime(0), ShouldResemble, JSONTime(time.Time{}))

	data, err := json.Marshal(JSONTime{})
	a.So(err, ShouldBeNil)
	a.So(string(data), ShouldResemble, `""`)

	data, err = json.Marshal(BuildTime(0))
	a.So(err, ShouldBeNil)
	a.So(string(data), ShouldResemble, `""`)

	data, err = json.Marshal(BuildTime(1465831736000000000))
	a.So(err, ShouldBeNil)
	a.So(string(data), ShouldResemble, `"2016-06-13T15:28:56Z"`)

	var time JSONTime
	err = json.Unmarshal([]byte(`"2016-06-13T15:28:56Z"`), &time)
	a.So(err, ShouldBeNil)
	a.So(time, ShouldResemble, BuildTime(1465831736000000000))
}
