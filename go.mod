module github.com/TheThingsNetwork/ttn

go 1.11

replace github.com/TheThingsNetwork/ttn/api => ./api

replace github.com/TheThingsNetwork/ttn/core/proxy => ./core/proxy

replace github.com/TheThingsNetwork/ttn/core/types => ./core/types

replace github.com/TheThingsNetwork/ttn/mqtt => ./mqtt

replace github.com/TheThingsNetwork/ttn/utils/errors => ./utils/errors

replace github.com/TheThingsNetwork/ttn/utils/random => ./utils/random

replace github.com/TheThingsNetwork/ttn/utils/security => ./utils/security

replace github.com/TheThingsNetwork/ttn/utils/testing => ./utils/testing

replace github.com/TheThingsNetwork/ttn/utils/toa => ./utils/toa

replace github.com/brocaar/lorawan => github.com/ThethingsIndustries/legacy-lorawan-lib v0.0.0-20190212122748-b905ab327304

replace github.com/robertkrimen/otto => github.com/ThethingsIndustries/otto v0.0.0-20181129100957-6ddbbb60554a

require (
	github.com/TheThingsNetwork/api v0.0.0-20190522113053-d844e8c040fc
	github.com/TheThingsNetwork/go-account-lib v0.0.0-20190516094738-77d15a3f8875
	github.com/TheThingsNetwork/go-cayenne-lib v1.0.0
	github.com/TheThingsNetwork/go-utils v0.0.0-20190813113035-8715cf82e887
	github.com/TheThingsNetwork/ttn/api v0.0.0-20190516093004-b66899428ed5
	github.com/TheThingsNetwork/ttn/core/proxy v0.0.0-20190520085727-78600a8e394e
	github.com/TheThingsNetwork/ttn/core/types v0.0.0-20190517101034-52d38c791f1e
	github.com/TheThingsNetwork/ttn/mqtt v0.0.0-20190516093004-b66899428ed5
	github.com/TheThingsNetwork/ttn/utils/errors v0.0.0-20190516093004-b66899428ed5
	github.com/TheThingsNetwork/ttn/utils/random v0.0.0-20190516093004-b66899428ed5
	github.com/TheThingsNetwork/ttn/utils/security v0.0.0-20190516093004-b66899428ed5
	github.com/TheThingsNetwork/ttn/utils/testing v0.0.0-20190520084050-7adf4a69a7c3
	github.com/TheThingsNetwork/ttn/utils/toa v0.0.0-20190520085727-78600a8e394e
	github.com/apex/log v1.1.0
	github.com/bluele/gcache v0.0.0-20190301044115-79ae3b2d8680
	github.com/brocaar/lorawan v0.0.0-20170626123636-a64aca28516d
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fatih/color v1.7.0 // indirect
	github.com/fatih/structs v1.1.0
	github.com/gogo/protobuf v1.2.1
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/gosuri/uitable v0.0.3
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mwitkow/go-grpc-middleware v1.0.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.2
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/tj/go-elastic v0.0.0-20171221160941-36157cbbebc2
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f // indirect
	golang.org/x/net v0.0.0-20191002035440-2ec189313ef0
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	google.golang.org/appengine v1.6.0 // indirect
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c
	google.golang.org/grpc v1.25.1
	gopkg.in/redis.v5 v5.2.9
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.2.3
)
