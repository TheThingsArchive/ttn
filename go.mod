module github.com/TheThingsNetwork/ttn

go 1.14

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
	github.com/TheThingsNetwork/api v0.0.0-20200324103623-039923721bb6
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
	github.com/apex/log v1.1.2
	github.com/bluele/gcache v0.0.0-20190518031135-bc40bd653833
	github.com/brocaar/lorawan v0.0.0-20170626123636-a64aca28516d
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fatih/color v1.9.0 // indirect
	github.com/fatih/structs v1.1.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.3.5
	github.com/gosuri/uitable v0.0.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.3
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/smartystreets/assertions v1.0.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.6
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.2
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/tj/go-elastic v0.0.0-20171221160941-36157cbbebc2
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59 // indirect
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20200323114720-3f67cca34472
	google.golang.org/grpc v1.28.0
	gopkg.in/ini.v1 v1.55.0 // indirect
	gopkg.in/redis.v5 v5.2.9
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.2.8
)
