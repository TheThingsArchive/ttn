// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/health"
	"github.com/TheThingsNetwork/ttn/utils/security"
	"github.com/htdvisser/grpc-testing/test"
	. "github.com/smartystreets/assertions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestTLSConfig(t *testing.T) {
	a := New(t)
	svc := new(Announcement)

	{
		_, err := svc.TLSConfig()
		a.So(err, ShouldNotBeNil)
	}

	svc.NetAddress = "localhost:0"

	{
		tlsConfig, err := svc.TLSConfig()
		a.So(err, ShouldBeNil)
		a.So(tlsConfig, ShouldBeNil)
	}

	svc.Certificate = "invalid"

	{
		_, err := svc.TLSConfig()
		a.So(err, ShouldNotBeNil)
	}
}

func TestDialInsecure(t *testing.T) {
	a := New(t)

	testLogger := test.NewLogger()
	log.Set(testLogger)
	defer testLogger.Print(t)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	health.RegisterServer(s)
	go s.Serve(lis)

	svc := new(Announcement)

	_, err = svc.Dial(nil)
	a.So(err, ShouldNotBeNil)

	svc.NetAddress = lis.Addr().String()

	conn, err := svc.Dial(nil)
	a.So(err, ShouldBeNil)

	{
		ok, err := health.Check(conn)
		a.So(err, ShouldBeNil)
		a.So(ok, ShouldBeTrue)
	}

}

func TestDialSecure(t *testing.T) {
	a := New(t)

	tmp := os.TempDir()
	security.GenerateKeypair(tmp)
	security.GenerateCert(tmp, "localhost", "127.0.0.1", "::1")

	cert, _ := ioutil.ReadFile(path.Join(tmp, "server.cert"))
	key, _ := ioutil.ReadFile(path.Join(tmp, "server.key"))

	cer, err := tls.X509KeyPair(cert, key)
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}

	testLogger := test.NewLogger()
	log.Set(testLogger)
	defer testLogger.Print(t)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	health.RegisterServer(s)
	go s.Serve(lis)

	svc := new(Announcement)

	_, err = svc.Dial(nil)
	a.So(err, ShouldNotBeNil)

	svc.NetAddress = lis.Addr().String()
	svc.Certificate = string(cert)

	conn, err := svc.Dial(nil)
	a.So(err, ShouldBeNil)

	{
		ok, err := health.Check(conn)
		a.So(err, ShouldBeNil)
		a.So(ok, ShouldBeTrue)
	}

}
