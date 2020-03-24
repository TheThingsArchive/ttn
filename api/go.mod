module github.com/TheThingsNetwork/ttn/api

go 1.14

replace github.com/TheThingsNetwork/ttn/utils/errors => ../utils/errors

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/TheThingsNetwork/api v0.0.0-20200324103623-039923721bb6
	github.com/TheThingsNetwork/go-utils v0.0.0-20190516083235-bdd4967fab4e
	github.com/TheThingsNetwork/ttn/utils/errors v0.0.0-20190516081709-034d40b328bd
	github.com/apex/log v1.1.2
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/juju/ratelimit v1.0.1
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/shirou/gopsutil v2.20.2+incompatible
	github.com/smartystreets/assertions v1.0.1
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8
	google.golang.org/grpc v1.28.0
)
