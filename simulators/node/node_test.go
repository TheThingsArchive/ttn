// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package node

import (
	"math/rand"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewNode(t *testing.T) {
	rand.Seed(1234567) // Let's use a static seed for testing

	Convey("Given a Node", t, func() {
		ctx := GetLogger(t, "Node")
		node := New(100, ctx)

		messages := make(chan string, 3)

		Convey("When sending three messages", func() {
			node.NextMessage(messages)
			node.NextMessage(messages)
			node.NextMessage(messages)

			Convey("They should be published to the channel", func() {
				So(len(messages), ShouldEqual, 3)
			})
		})

	})

}
