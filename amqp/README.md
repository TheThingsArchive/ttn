# API Reference

TTN makes use of an AMQP Topic exchange. See [this](https://www.rabbitmq.com/tutorials/amqp-concepts.html) for details.

AMQP is not yet available in the public network.

* Host: The address of the handler on which your application is registered.
* Port: `5672`
* Exchange: `ttn.handler`

## Uplink Messages

**Routing key:** `<AppID>.devices.<DevID>.up`

Wildcards are allowed. For example `<AppID>.devices.*.up` to get uplink messages for all devices.

**Message**

```js
{
  "app_id": "my-app-id",                 // Same as in the topic
  "dev_id": "my-dev-id",                 // Same as in the topic
  "hardware_serial": "0102030405060708", // In case of LoRaWAN: the DevEUI
  "port": 1,                             // LoRaWAN FPort
  "counter": 2,                          // LoRaWAN frame counter
  "is_retry": false,                     // Is set to true if this message is a retry (you could also detect this from the counter)
  "confirmed": false,                    // Is set to true if this message was a confirmed message
  "payload_raw": "AQIDBA==",             // Base64 encoded payload: [0x01, 0x02, 0x03, 0x04]
  "payload_fields": {},                  // Object containing the results from the payload functions - left out when empty
  "metadata": {
    "time": "1970-01-01T00:00:00Z",      // Time when the server received the message
    "frequency": 868.1,                  // Frequency at which the message was sent
    "modulation": "LORA",                // Modulation that was used - LORA or FSK
    "data_rate": "SF7BW125",             // Data rate that was used - if LORA modulation
    "bit_rate": 50000,                   // Bit rate that was used - if FSK modulation
    "coding_rate": "4/5",                // Coding rate that was used
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
    ],
    "latitude": 52.2345,                 // Latitude of the device
    "longitude": 6.2345,                 // Longitude of the device
    "altitude": 2                        // Altitude of the device
  }
}
```

Note: Some values may be omitted if they are `null`, `false`, `""` or `0`.


**Usage (Go client):**

```go
package main

import (
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/amqp"
	"github.com/TheThingsNetwork/ttn/core/types"
)

func main() {

	ctx := apex.Stdout().WithField("Example", "Go AMQP client")
	c := amqp.NewClient(ctx, "guest", "guest", "localhost:5672")
	c.Connect()
	s := c.NewSubscriber("ttn.handler", "", false, true)
	s.Open()
	s.SubscribeDeviceUplink("my-app-id", "my-dev-id",
		func(_ amqp.Subscriber, appID string, devID string, req types.UplinkMessage) {
			ctx.Info("Uplink received")
			//...
		})
	//...
}
```

## Downlink Messages

**Routing key:** `<AppID>.devices.<DevID>.down`

**Message:**
```js
{
  "port": 1,                 // LoRaWAN FPort
  "confirmed": false,        // Whether the downlink should be confirmed by the device
  "payload_raw": "AQIDBA==", // Base64 encoded payload: [0x01, 0x02, 0x03, 0x04]
}
```

**Usage (RabbitMQ):**
`rabbitmqadmin publish exchange='ttn.handler' routing_key='my-app-id.devices.my-dev-id.down' payload='{"port":1,"payload_raw":"AQIDBA=="}'`

**Usage (Go client):**
```go
package main

import (
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/amqp"
	"github.com/TheThingsNetwork/ttn/core/types"
)

func main() {

	ctx := apex.Stdout().WithField("test", "test")
	c := amqp.NewClient(ctx, "guest", "guest", "localhost:5672")
	c.Connect()
	p := c.NewPublisher("ttn.handler")
  if err := p.Open(); err != nil {
    ctx.WithError(err).Error("Could not open publishing channel")
  }
  defer p.Close()
	d := types.DownlinkMessage{
    AppID:      "my-app-id",
    DevID:      "my-dev-id",
    FPort:      1,
    PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04}}
  p.PublishDownlink(d)
	//...
}
```

## Device Events

**Routing key:** 

* `<AppID>.devices.<DevID>.events.<event>`
* `0102030405060708.devices.abcdabcd12345678.events.activations`
* `*.devices.*.events.*`

**Message:**
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

**Usages (Go client):**

```go
package main

import (
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/amqp"
	"github.com/TheThingsNetwork/ttn/core/types"
)

func main() {

  ctx := apex.Stdout().WithField("test", "test")
  c := amqp.NewClient(ctx, "guest", "guest", "localhost:5672")
  if err := c.Connect(); err != nil {
    ctx.WithError(err).Error("Could not connect")
  }
  s := c.NewSubscriber("ttn.handler", "", true, false)
  if err := s.Open(); err != nil {
    ctx.WithError(err).Error("Could not open subcription channel")
  }
	err = s.SubscribeAppEvents("my-app-id", "some-event",
			func(_ Subscriber, appID string, eventType types.EventType, payload []byte) {
			  // Do your stuff
			})
	//...
}
```