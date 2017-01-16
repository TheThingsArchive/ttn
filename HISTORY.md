# History

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
