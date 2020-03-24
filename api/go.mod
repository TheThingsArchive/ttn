module github.com/TheThingsNetwork/ttn/api

go 1.14

replace github.com/TheThingsNetwork/ttn/utils/errors => ../utils/errors

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/TheThingsNetwork/api v0.0.0-20190516085542-c732802571cf
	github.com/TheThingsNetwork/go-utils v0.0.0-20190516083235-bdd4967fab4e
	github.com/TheThingsNetwork/ttn/utils/errors v0.0.0-20190516081709-034d40b328bd
	github.com/apex/log v1.1.2
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/juju/ratelimit v1.0.1
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/shirou/gopsutil v2.20.2+incompatible
	github.com/smartystreets/assertions v1.0.1
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20200323114720-3f67cca34472 // indirect
	google.golang.org/grpc v1.28.0
)
