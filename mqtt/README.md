# The Things Network MQTT

This package contains the code that is used to publish and subscribe to MQTT.
This README describes the topics and messages that are used 

## Uplink Messages

**Topic:** `<AppID>/devices/<DevID>/up`

**Message:**

```js
{
  "port": 1,                          // LoRaWAN FPort
  "counter": 2,                       // LoRaWAN frame counter
  "payload_raw": "AQIDBA==",          // Base64 encoded payload: [0x01, 0x02, 0x03, 0x04]
  "payload_fields": {},               // Object containing the results from the payload functions - left out when empty
  "metadata": {
    "time": "1970-01-01T00:00:00Z",   // Time when the server received the message
    "frequency": 868.1,               // Frequency at which the message was sent
    "modulation": "LORA",             // Modulation that was used - currently only LORA. In the future we will support FSK as well
    "data_rate": "SF7BW125",          // Data rate that was used - if LORA modulation
    "bit_rate": 50000,                // Bit rate that was used - if FSK modulation
    "coding_rate": "4/5",             // Coding rate that was used
    "gateways": [
      {
        "eui": "0102030405060708",      // EUI of the gateway
        "timestamp": 12345,             // Timestamp when the gateway received the message
        "time": "1970-01-01T00:00:00Z", // Time when the gateway received the message - left out when gateway does not have synchronized time 
        "channel": 0,                   // Channel where the gateway received the message
        "rssi": -25,                    // Signal strength of the received message
        "snr": 5,                       // Signal to noise ratio of the received message
        "rf_chain": 0,                  // RF chain where the gateway received the message
      },
      //...more if received by more gateways...
    ]
  }
}
```

Note: Some values may be omitted if they are `null`, `""` or `0`.

**Usage (Mosquitto):** `mosquitto_sub -d -t 'my-app-id/devices/my-dev-id/up'`

**Usage (Go client):**

```go
ctx := log.WithField("Example", "Go Client")
client := NewClient(ctx, "ttnctl", "my-app-id", "my-access-key", "staging.thethingsnetwork.org:1883")
if err := client.Connect(); err != nil {
  ctx.WithError(err).Fatal("Could not connect")
}
token := client.SubscribeDeviceUplink("my-app-id", "my-dev-id", func(client Client, appID string, devID string, req UplinkMessage) {
  // Do something with the uplink message
})
token.Wait()
if err := token.Error(); err != nil {
  ctx.WithError(err).Fatal("Could not subscribe")
}
```

## Downlink Messages

**Topic:** `<AppID>/devices/<DevID>/down`

**Message:**

```js
{
  "port": 1,                 // LoRaWAN FPort
  "payload_raw": "AQIDBA==", // Base64 encoded payload: [0x01, 0x02, 0x03, 0x04]
}
```

**Usage (Mosquitto):** `mosquitto_pub -d -t 'my-app-id/devices/my-dev-id/down' -m '{"port":1,"payload_raw":"AQIDBA=="}'`

**Usage (Go client):**

```go
ctx := log.WithField("Example", "Go Client")
client := NewClient(ctx, "ttnctl", "my-app-id", "my-access-key", "staging.thethingsnetwork.org:1883")
if err := client.Connect(); err != nil {
  ctx.WithError(err).Fatal("Could not connect")
}
token := client.PublishDownlink(DownlinkMessage{
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

## Device Activations

**Topic:** `<AppID>/devices/<DevID>/activations`

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

**Usage (Mosquitto):** `mosquitto_sub -d -t 'my-app-id/devices/my-dev-id/activations'`

**Usage (Go client):**

```go
ctx := log.WithField("Example", "Go Client")
client := NewClient(ctx, "ttnctl", "my-app-id", "my-access-key", "staging.thethingsnetwork.org:1883")
if err := client.Connect(); err != nil {
  ctx.WithError(err).Fatal("Could not connect")
}
token := client.SubscribeDeviceActivations("my-app-id", "my-dev-id", func(client Client, appID string, devID string, req Activation) {
  // Do something with the activation
})
token.Wait()
if err := token.Error(); err != nil {
  ctx.WithError(err).Fatal("Could not subscribe")
}
```
