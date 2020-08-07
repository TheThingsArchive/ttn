module github.com/TheThingsNetwork/ttn/core/proxy

go 1.14

replace github.com/TheThingsNetwork/ttn/utils/testing => ../../utils/testing

require (
	github.com/TheThingsNetwork/go-utils v0.0.0-20200807125606-b3493662e4bf
	github.com/TheThingsNetwork/ttn/utils/testing v0.0.0-20190520084050-7adf4a69a7c3
	github.com/gogo/protobuf v1.3.1
	github.com/grpc-ecosystem/grpc-gateway v1.14.3
	github.com/smartystreets/assertions v1.0.1
)
