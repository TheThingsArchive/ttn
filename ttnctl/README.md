# The Things Network Control Utility - `ttnctl`

This document is a simple guide to `ttnctl`.

## Configuration

Configuration is done with:

* Command line arguments
* Environment variables
* Configuration file

The following configuration options can be set:

| CLI flag / yaml key   | Environment Var             | Description  |
|-----------------------|-----------------------------|--------------|
| `app-id`              | `TTNCTL_APP_ID`             | The application ID that should be used |
| `app-eui`             | `TTNCTL_APP_EUI`            | The LoRaWAN AppEUI that should be used |
| `debug`               | `TTNCTL_DEBUG`              | Print debug logs |
| `discovery-server`    | `TTNCTL_DISCOVERY_SERVER`   | The address and port of the discovery server |
| `ttn-router`          | `TTNCTL_TTN_ROUTER`         | The id of the router |
| `ttn-handler`         | `TTNCTL_TTN_HANDLER`        | The id of the handler |
| `mqtt-broker`         | `TTNCTL_MQTT_BROKER`        | The address and port of the MQTT broker |
| `ttn-account-server`  | `TTNCTL_TTN_ACCOUNT_SERVER` | The protocol, address (and port) of the account server |

**Configuration for Development:** Copy `../.env/ttnctl.yaml.dev-example` to `~/.ttnctl.yaml`

## Command Options

The arguments and flags for each command are shown when executing a command with the `--help` flag.

## Getting Started

* Create an account: `ttnctl user register [username] [e-mail]`
    * Note: You might have to verify your email before you can login.
* Get a client access code on the account server by clicking *ttnctl access
  code* on the home page.
* Login with the client code you received `ttnctl user login [client code]`
* List your applications: `ttnctl applications list`
* Create a new application: `ttnctl applications create [AppID] [Description]`
* Select the application you want to use from now on: `ttnctl applications select`
* Register the application with the Handler: `ttnctl applications register`
* List the devices in your application: `ttnctl devices list`
* Create a new device: `ttnctl devices create [Device ID]`
* Get info about the device: `ttnctl devices info [Device ID]`
* Personalize the device (optional): `ttnctl devices personalize [Device ID]`
* Set the next downlink for a device: `ttnctl downlink [Device ID] [Payload]`
* Subscribe to messages from your devices: `ttnctl subscribe`
* Get payload functions for your application: `ttnctl applications pf`
* Set payload functions for your application: `ttnctl applications pf set [decoder/converter/validator]`

## List of commands

```
ttnctl
|-- user
   |-- register
   |-- login
   |-- logout
|-- applications
   |-- create
   |-- delete
   |-- list
   |-- select
   |-- info
   |-- register
   |-- unregister
   |-- pf
      |-- set
|-- devices
   |-- create
   |-- delete
   |-- list
   |-- info
   |-- set
   |-- personalize
|-- downlink
|-- subscribe
|-- version
```
