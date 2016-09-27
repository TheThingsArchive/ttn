# The Things Network Control Utility - `ttnctl`

`ttnctl` can be used to manage The Things Network from the command line. 

[Documentation](https://www.thethingsnetwork.org/docs/cli/)

## Configuration File

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

## Development

**Configuration for Development:** Copy `../.env/ttnctl.yaml.dev-example` to `~/.ttnctl.yaml`

## License

Source code for The Things Network is released under the MIT License, which can be found in the [LICENSE](../LICENSE) file. A list of authors can be found in the [AUTHORS](../AUTHORS) file.
