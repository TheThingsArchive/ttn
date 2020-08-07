module github.com/TheThingsNetwork/ttn/mqtt

go 1.14

replace github.com/TheThingsNetwork/ttn/core/types => ../core/types

replace github.com/TheThingsNetwork/ttn/utils/random => ../utils/random

replace github.com/TheThingsNetwork/ttn/utils/testing => ../utils/testing

replace github.com/brocaar/lorawan => github.com/ThethingsIndustries/legacy-lorawan-lib v0.0.0-20190212122748-b905ab327304

require (
	github.com/TheThingsNetwork/go-utils v0.0.0-20200807125606-b3493662e4bf
	github.com/TheThingsNetwork/ttn/core/types v0.0.0-20200807123328-b39cc6b19c87
	github.com/TheThingsNetwork/ttn/utils/random v0.0.0-20200807123328-b39cc6b19c87
	github.com/TheThingsNetwork/ttn/utils/testing v0.0.0-20190516092602-86414c703ee1
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/smartystreets/assertions v1.0.1
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sys v0.0.0-20200806125547-5acd03effb82 // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/genproto v0.0.0-20200806141610-86f49bd18e98 // indirect
	google.golang.org/grpc v1.31.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)
