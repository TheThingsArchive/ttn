module github.com/TheThingsNetwork/ttn/core/types

go 1.14

replace github.com/TheThingsNetwork/ttn/utils/errors => ../../utils/errors

replace github.com/brocaar/lorawan => github.com/ThethingsIndustries/legacy-lorawan-lib v0.0.0-20190212122748-b905ab327304

require (
	github.com/TheThingsNetwork/ttn/utils/errors v0.0.0-20200807123328-b39cc6b19c87
	github.com/brocaar/lorawan v0.0.0-20200726141338-ee070f85d494
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/jacobsa/crypto v0.0.0-20190317225127-9f44e2d11115 // indirect
	github.com/jacobsa/oglematchers v0.0.0-20150720000706-141901ea67cd // indirect
	github.com/jacobsa/oglemock v0.0.0-20150831005832-e94d794d06ff // indirect
	github.com/jacobsa/ogletest v0.0.0-20170503003838-80d50a735a11 // indirect
	github.com/jacobsa/reqtrace v0.0.0-20150505043853-245c9e0234cb // indirect
	github.com/smartystreets/assertions v1.0.1
	github.com/smartystreets/goconvey v1.6.4 // indirect
)
