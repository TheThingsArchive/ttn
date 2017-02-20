# API Reference

* Host: `<Region>.thethings.network`, where `<Region>` is last part of the handler you registered your application to, e.g. `eu`.
* Port: `1883` or `8883` for TLS
* PEM encoded CA certificate for TLS: [mqtt-ca.pem](https://console.thethingsnetwork.org/mqtt-ca.pem)
* Username: Application ID
* Password: Application Access Key

## Uplink Messages

**Topic:** `<AppID>/devices/<DevID>/up`

**Message:**

```js
{
  "app_id": "my-app-id",              // Same as in the topic
  "dev_id": "my-dev-id",              // Same as in the topic
  "hardware_serial": "0102030405060708", // In case of LoRaWAN: the DevEUI
  "port": 1,                          // LoRaWAN FPort
  "counter": 2,                       // LoRaWAN frame counter
  "is_retry": false,                  // Is set to true if this message is a retry (you could also detect this from the counter)
  "confirmed": false,                 // Is set to true if this message was a confirmed message
  "payload_raw": "AQIDBA==",          // Base64 encoded payload: [0x01, 0x02, 0x03, 0x04]
  "payload_fields": {},               // Object containing the results from the payload functions - left out when empty
  "metadata": {
    "time": "1970-01-01T00:00:00Z",   // Time when the server received the message
    "frequency": 868.1,               // Frequency at which the message was sent
    "modulation": "LORA",             // Modulation that was used - LORA or FSK
    "data_rate": "SF7BW125",          // Data rate that was used - if LORA modulation
    "bit_rate": 50000,                // Bit rate that was used - if FSK modulation
    "coding_rate": "4/5",             // Coding rate that was used
    "gateways": [
      {
        "gtw_id": "ttn-herengracht-ams", // EUI of the gateway
        "timestamp": 12345,              // Timestamp when the gateway received the message
        "time": "1970-01-01T00:00:00Z",  // Time when the gateway received the message - left out when gateway does not have synchronized time
        "channel": 0,                    // Channel where the gateway received the message
        "rssi": -25,                     // Signal strength of the received message
        "snr": 5,                        // Signal to noise ratio of the received message
        "rf_chain": 0,                   // RF chain where the gateway received the message
        "latitude": 52.1234,             // Latitude of the gateway reported in its status updates
        "longitude": 6.1234,             // Longitude of the gateway
        "altitude": 6                    // Altitude of the gateway
      },
      //...more if received by more gateways...
    ]
  }
}
```

Note: Some values may be omitted if they are `null`, `false`, `""` or `0`.

**Usage (Mosquitto):** `mosquitto_sub -h <Region>.thethings.network:1883 -d -t 'my-app-id/devices/my-dev-id/up'`

**Usage (Go client):**

```go
import (
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
)

func main() {
	ctx := apex.Stdout().WithField("Example", "Go Client")
	log.Set(ctx)

	client := mqtt.NewClient(ctx, "ttnctl", "my-app-id", "my-access-key", "<Region>.thethings.network:1883")
	if err := client.Connect(); err != nil {
		ctx.WithError(err).Fatal("Could not connect")
	}
	token := client.SubscribeDeviceUplink("my-app-id", "my-dev-id", func(client mqtt.Client, appID string, devID string, req types.UplinkMessage) {
		// Do something with the uplink message
	})
	token.Wait()
	if err := token.Error(); err != nil {
		ctx.WithError(err).Fatal("Could not subscribe")
	}
}
```

### Uplink Fields

Each uplink field will be published to its own topic `my-app-id/devices/my-dev-id/up/<field>`. The payload will be a string with the value in a JSON-style encoding. 

If your fields look like the following:

```js
{
  "water": true,
  "analog": [0, 255, 500, 1000],
  "gps": {
    "lat": 52.3736735,
    "lon": 4.886663
  },
  "text": "why are you using text?"
}
```

you will see this on MQTT:

* `my-app-id/devices/my-dev-id/up/water`: `true`
* `my-app-id/devices/my-dev-id/up/analog`: `[0, 255, 500, 1000]`
* `my-app-id/devices/my-dev-id/up/gps`: `{"lat":52.3736735,"lon":4.886663}`
* `my-app-id/devices/my-dev-id/up/gps/lat`: `52.3736735`
* `my-app-id/devices/my-dev-id/up/gps/lon`: `4.886663`
* `my-app-id/devices/my-dev-id/up/text`: `"why are you using text?"`

## Downlink Messages

**Topic:** `<AppID>/devices/<DevID>/down`

**Message:**

```js
{
  "port": 1,                 // LoRaWAN FPort
  "confirmed": false,        // Whether the downlink should be confirmed by the device
  "payload_raw": "AQIDBA==", // Base64 encoded payload: [0x01, 0x02, 0x03, 0x04]
}
```

**Usage (Mosquitto):** `mosquitto_pub -h <Region>.thethings.network:1883 -d -t 'my-app-id/devices/my-dev-id/down' -m '{"port":1,"payload_raw":"AQIDBA=="}'`

**Usage (Go client):**

_for setup, see **Uplink Messages**_

```go
token := client.PublishDownlink(types.DownlinkMessage{
  AppID:   "my-app-id",
  DevID:   "my-dev-id",
  FPort:   1,
  Payload: []byte{0x01, 0x02, 0x03, 0x04},
})
token.Wait()
if err := token.Error(); err != nil {
  ctx.WithError(err).Fatal("Could not publish")
}
```

### Downlink Fields

Instead of `payload_raw` you can also use `payload_fields` with an object of fields. This requires the application to be configured with an Encoder Payload Function which encodes the fields into a Buffer.

**Message:**

```js
{
  "port": 1,                 // LoRaWAN FPort
  "confirmed": false,        // Whether the downlink should be confirmed by the device
  "payload_fields": {
    "led": true
  }
}
```

**Usage (Mosquitto):** `mosquitto_pub -h <Region>.thethings.network:1883 -d -t 'my-app-id/devices/my-dev-id/down' -m '{"port":1,"payload_fields":{"led":true}}'`

**Usage (Go client):**

_for setup, see **Uplink Messages**_

```go
token := client.PublishDownlink(types.DownlinkMessage{
  AppID:   "my-app-id",
  DevID:   "my-dev-id",
  FPort:   1,
  Fields: map[string]interface{}{
    "led": true,
  },
})
token.Wait()
if err := token.Error(); err != nil {
  ctx.WithError(err).Fatal("Could not publish")
}
```

## Device Activations

**Topic:** `<AppID>/devices/<DevID>/events/activations`

**Message:**

```js
{
  "app_eui": "0102030405060708", // EUI of the application
  "dev_eui": "0102030405060708", // EUI of the device
  "dev_addr": "26001716",        // Assigned address of the device
  "metadata": {
    // Same as with Uplink Message
  }
}
```

**Usage (Mosquitto):** `mosquitto_sub -h <Region>.thethings.network:1883 -d -t 'my-app-id/devices/my-dev-id/events/activations'`

**Usage (Go client):**

_for setup, see **Uplink Messages**_

```go
token := client.SubscribeDeviceActivations("my-app-id", "my-dev-id", func(client Client, appID string, devID string, req Activation) {
  // Do something with the activation
})
token.Wait()
if err := token.Error(); err != nil {
  ctx.WithError(err).Fatal("Could not subscribe")
}
```

## Device Events

### Management Events

**Created:** `<AppID>/devices/<DevID>/events/create`
**Updated:** `<AppID>/devices/<DevID>/events/update`
**Deleted:** `<AppID>/devices/<DevID>/events/delete`

### Downlink Events

**Downlink Scheduled:** `<AppID>/devices/<DevID>/events/down/scheduled`  
payload: _null_

**Downlink Sent:** `<AppID>/devices/<DevID>/events/down/sent`  

```js
{
  "payload": "Base64 encoded LoRaWAN packet",
  "gateway_id": "some-gateway",
  "config": {
    "modulation": "LORA",
    "data_rate": "SF7BW125",
    "counter": 123,
    "frequency": 868300000,
    "power": 14
  }
}
```

**Downlink Acknowledgements:** `<AppID>/devices/<DevID>/events/down/acks`   
payload: _null_

### Error Events

The payload of error events is a JSON object with the error's description.

**Uplink Errors:** `<AppID>/devices/<DevID>/events/up/errors`  
**Downlink Errors:** `<AppID>/devices/<DevID>/events/down/errors`  
**Activation Errors:** `<AppID>/devices/<DevID>/events/activations/errors`  

Example: `{"error":"Activation DevNonce not valid: already used"}`
