module github.com/TheThingsNetwork/ttn/api

go 1.14

replace github.com/TheThingsNetwork/ttn/utils/errors => ../utils/errors

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/TheThingsNetwork/api v0.0.0-20200807125557-7bae06ae0e7b
	github.com/TheThingsNetwork/go-utils v0.0.0-20200807125606-b3493662e4bf
	github.com/TheThingsNetwork/ttn/utils/errors v0.0.0-20200807123328-b39cc6b19c87
	github.com/apex/log v1.1.2
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/juju/ratelimit v1.0.1
	github.com/prometheus/common v0.11.1 // indirect
	github.com/shirou/gopsutil v2.20.7+incompatible
	github.com/smartystreets/assertions v1.0.1
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	google.golang.org/grpc v1.31.0
)
