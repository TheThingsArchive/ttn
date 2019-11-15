module github.com/TheThingsNetwork/ttn/core/proxy

go 1.11

replace github.com/TheThingsNetwork/ttn/utils/testing => ../../utils/testing

require (
	github.com/TheThingsNetwork/go-utils v0.0.0-20190516083235-bdd4967fab4e
	github.com/TheThingsNetwork/ttn/utils/testing v0.0.0-20190520084050-7adf4a69a7c3
	github.com/gogo/protobuf v1.2.1
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3
)
