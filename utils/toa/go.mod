module github.com/TheThingsNetwork/ttn/utils/toa

go 1.11

replace github.com/TheThingsNetwork/ttn/core/types => ../../core/types

require (
	github.com/TheThingsNetwork/ttn/core/types v0.0.0-20190517101034-52d38c791f1e
	github.com/pkg/errors v0.8.1 // indirect
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3
	google.golang.org/grpc v1.20.1 // indirect
)
