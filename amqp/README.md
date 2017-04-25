AMQP 0.9.1 API Reference
========================

If you don't know what AMQP is you can read [this](https://blogs.vmware.com/vfabric/2013/02/choosing-your-messaging-protocol-amqp-mqtt-or-stomp.html)
and [this](https://www.rabbitmq.com/tutorials/amqp-concepts.html).

TTN use the Topic AMQP exchange. The default exchange name is `amq.topic`

__Public Network__

*Building*

## Routing keys behavior

Routing keys act like filters. When a message arrive on the AMQP exchange
the key is read and the message forwarded to all matching queue with the
a matching routing queue.

Routing keys make use of the wildcard `*` to match multiple queues.

Example:
* Match a device uplink: `<App_ID>.devices.<Dev_ID>.up`
* Match all App devices uplink: `<App_ID>.devices.*.up`

## Uplink


**Routing key:** `<App_ID>.devices.<Dev_ID>.up`

**Message**
```json
{
  "app_id": "fiware-dev",              // Same as in the topic
  "dev_id": "f1",              // Same as in the topic
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
    ],
    "latitude": 52.2345,              // Latitude of the device
    "longitude": 6.2345,              // Longitude of the device
    "altitude": 2                     // Altitude of the device
  }
}
```
Note: Some values may be omitted if they are null, false, "" or 0.

#### **Usages**

* **Publish**
  * **ttn:** `ttn devices simulate <dev_id> '{"app_id":"<App_ID>",...}'`  
  _Note_: You will have to register your application in the handler first and
create the device.

  * **RabbitMQ:** `rabbitmqadmin publish exchange=ttn.handler routing_key=<app-id>.devices.<dev-id>.up payload='{"app_id":"<App_iD"",...}'`  
 *Only work if your are running rabbitmq on your host*
 
* **Subscribe**

  * **go:**

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
	s := c.NewSubscriber("amq.topic", "", false, true)
	s.Open()
	s.SubscribeDeviceUplink("app-id", "dev-id",
		func(_ amqp.Subscriber, appID string, devID string, req types.UplinkMessage) {
			ctx.Info("Uplink received")
			//...
		})
	//...
}
```

## Downlink

**Routing key:** `<App_ID>.devices.<Dev_ID>.down`

**Message:**
```json
{
  "port": 1,                 // LoRaWAN FPort
  "confirmed": false,        // Whether the downlink should be confirmed by the device
  "payload_raw": "AQIDBA==", // Base64 encoded payload: [0x01, 0x02, 0x03, 0x04]
}
```

**Usages**

* **rabbitMQ:** `'rabbitmqadmin publish exchange=ttn.handler routing_key=app-id.devices.dev-id.down payload={"port":1,"payload_raw":"AQIDBA=="}'`

* **go:**
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
	p := c.NewPublisher("amq.topic")
  if err := p.Open(); err != nil {
    ctx.WithError(err).Error("Could not open publishing channel")
  }
  defer p.Close()
	d := types.DownlinkMessage{
    AppID:      "app-id",
    DevID:      "dev-id",
    FPort:      1,
    PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04}}
  p.PublishDownlink(d)
	//...
}
```

## Events

**Routing key:** `<App_ID>.devices.<Dev_ID>.<event>`

**Message:**
```json
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

**Usages (GO)**

*Subscription:*
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
  s := c.NewSubscriber("amq.topic", "", true, false)
  if err := s.Open(); err != nil {
    ctx.WithError(err).Error("Could not open subcription channel")
  }
	err = s.SubscribeAppEvents("app-id", "some-event",
			func(_ Subscriber, appID string, eventType types.EventType, payload []byte) {
			  // Do your stuff
			})
	//...
}
```

*Publish:*
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
  p := c.NewPublisher("amq.topic")
  if err := p.Open(); err != nil {
    ctx.WithError(err).Error("Could not open publishing channel")
  }
  defer p.Close()
  p.PublishAppEvent("app-id", "some-event", "payload")
	//...
}
```
