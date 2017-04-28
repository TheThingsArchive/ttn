package component

import (
	"context"
	"net"
	"net/http"
	"testing"

	"io/ioutil"

	"time"

	"github.com/TheThingsNetwork/ttn/api/pool"
	assertions "github.com/smartystreets/assertions"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestStatus(t *testing.T) {
	a := assertions.New(t)
	c := new(Component)

	a.So(getStatus(c), assertions.ShouldEqual, StatusUnhealthy)
	a.So(c.GetStatus(), assertions.ShouldEqual, StatusUnhealthy)

	c.SetStatus(StatusHealthy)
	a.So(getStatus(c), assertions.ShouldEqual, StatusHealthy)
	a.So(c.GetStatus(), assertions.ShouldEqual, StatusHealthy)

	c.SetStatus(StatusUnhealthy)

	a.So(getStatus(c), assertions.ShouldEqual, StatusUnhealthy)
	a.So(c.GetStatus(), assertions.ShouldEqual, StatusUnhealthy)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	c.RegisterHealthServer(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("test HealthServer exited with error: %s", err)
		}
	}()

	conn, err := pool.Global.DialInsecure(lis.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	pb := healthpb.NewHealthClient(conn)

	viper.Set("health-port", 10700)
	initStatus(c)
	viper.Set("health-port", "")

	time.Sleep(100 * time.Millisecond)

	checkStatus := func(expectedString string, expectedPb healthpb.HealthCheckResponse_ServingStatus) {
		res, err := http.Get("http://localhost:10700/healthz")
		a.So(err, assertions.ShouldBeNil)
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		a.So(string(body), assertions.ShouldEqual, expectedString)

		pbRes, err := pb.Check(context.Background(), &healthpb.HealthCheckRequest{Service: statusName})
		a.So(err, assertions.ShouldBeNil)
		a.So(pbRes.Status, assertions.ShouldEqual, expectedPb)
	}

	checkStatus("Status is UNHEALTHY", healthpb.HealthCheckResponse_NOT_SERVING)

	setStatus(c, StatusHealthy)

	checkStatus("Status is HEALTHY", healthpb.HealthCheckResponse_SERVING)
}
