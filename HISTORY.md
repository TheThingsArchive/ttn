# History

## 2.4.0 (2017-02-03)

The 2.4.0 release contains the following API Changes:

- Add `SimulateUplink` RPC to Application Manager API. Can be used to test (for example) integrations or external applications.
- Add "confirmed" field to DownlinkMessage
- Add "is_retry" field to UplinkMessage

Other changes:

- Use `gogoproto` instead of regular `proto` (and use it in jsonpb marshaling for HTTP API)
- Better handling of confirmed uplink
- Add support for confirmed downlink
- Try to create AMQP exchange instead of crashing
- Add TLS support to MQTT Client
- Add validation for DataRates
- Add more metadata to log messages

## 2.3.0 (2017-01-20)

The 2.3.0 release contains the following API Changes:

- Added pagination with `limit` and `offset` in gRPC metadata or HTTP query string
- Handler Devices "List" operation returns total number of devices in gRPC Header

Other changes:

- Use `TheThingsNetwork/go-utils/log` instead of `apex/log`
- Use `TheThingsNetwork/go-utils/encoding` for Redis Maps and optimize list operations
- Don't crash on unexpected Otto panics
- Testing payload functions in `ttnctl`

## 2.2.0 (2017-01-16)

The 2.2.0 release contains the following API Changes:

- Add Trace to Uplink/Downlink/Activations
- Add AppID,DevID,HardwareSerial to MQTT Uplink
- Add Latitude,Longitude,Altitude to Device (Handler) and Uplink
- Add Create/Update/Delete events for Devices

Furthermore it adds the following functionality:

- Restarting Gateway-Router streams if auth info changes

## 2.1.0 (2017-01-10)

The 2.1.0 release contains the following API Changes:

- Add MQTT/AMQP address to the Discovery announcement.
- Add trusted flag for gateways in metadata and status.

Furthermore it adds the following functionality:

- Monitoring (NOC) of the Broker components
- Support for the China 779-787 and Europe 433 bands
- The Korea 920-923 channel plan

## 2.0.0 (2016-12-14)

With the 2.0.0 release we now declare the v2 systems "out of preview".

## v2-preview (2016-08-21)

The v2-preview is a rewrite of almost all backend code, following a clear separation of concerns between components. Instead of using hard to read EUIs for applications and devices, you can now work with IDs that you can choose yourself. We added many new features and are sure that you'll love them.

## v1-staging (2016-04-18)

With the "staging" release we introduced device management, downlink messages, over the air activation, message encryption and MQTT feeds for receiving messages.

![the command-line interface for managing devices](https://ttn.blob.core.windows.net/upload/ttnctl-staging.png)

## v0-croft (2016-08-20)

The day before the official launch of The Things Network, we sent our first text with an application built on The Things Network.

![iPhone showing the first message sent over The Things Network](https://ttn.blob.core.windows.net/upload/slack_for_ios_upload_1024.jpg)
