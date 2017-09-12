// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package security

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestKeyPairFuncs(t *testing.T) {
	a := New(t)

	location := os.TempDir()

	err := GenerateKeypair(location)
	a.So(err, ShouldBeNil)

	_, err = LoadKeypair(location + "/derp")
	a.So(err, ShouldNotBeNil)

	key, err := LoadKeypair(location)
	a.So(err, ShouldBeNil)
	a.So(key, ShouldNotBeNil)

	ioutil.WriteFile(location+"/server.key", []byte{}, 0644)

	_, err = LoadKeypair(location)
	a.So(err, ShouldNotBeNil)
}

func TestCertFuncs(t *testing.T) {
	a := New(t)

	location := os.TempDir()

	GenerateKeypair(location)

	err := GenerateCert(location, "test cert", "localhost")
	a.So(err, ShouldBeNil)

	_, err = LoadCert(location + "/derp")
	a.So(err, ShouldNotBeNil)

	cert, err := LoadCert(location)
	a.So(err, ShouldBeNil)
	a.So(cert, ShouldNotBeNil)
}
