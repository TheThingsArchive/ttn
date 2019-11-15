module github.com/TheThingsNetwork/ttn/mqtt

go 1.11

replace github.com/TheThingsNetwork/ttn/core/types => ../core/types

replace github.com/TheThingsNetwork/ttn/utils/random => ../utils/random

replace github.com/TheThingsNetwork/ttn/utils/testing => ../utils/testing

replace github.com/brocaar/lorawan => github.com/ThethingsIndustries/legacy-lorawan-lib v0.0.0-20190212122748-b905ab327304

require (
	github.com/TheThingsNetwork/go-utils v0.0.0-20190516083235-bdd4967fab4e
	github.com/TheThingsNetwork/ttn/core/types v0.0.0-20190516092602-86414c703ee1
	github.com/TheThingsNetwork/ttn/utils/random v0.0.0-20190516092602-86414c703ee1
	github.com/TheThingsNetwork/ttn/utils/testing v0.0.0-20190516092602-86414c703ee1
	github.com/eclipse/paho.mqtt.golang v1.1.0
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3
)
