# ApplicationManager API Reference

ApplicationManager manages application and device registrations on the Handler

To protect our quality of service, you can make up to 5000 calls to the
ApplicationManager API per hour. Once you go over the rate limit, you will
receive an error response.

## Methods

### `RegisterApplication`

Applications should first be registered to the Handler with the `RegisterApplication` method

- Request: [`ApplicationIdentifier`](#handlerapplicationidentifier)
- Response: [`Empty`](#handlerapplicationidentifier)

#### HTTP Endpoint

- `POST` `/applications`

#### JSON Request Format

```json
{
  "app_id": "some-app-id"
}
```

#### JSON Response Format

```json
{}
```

### `GetApplication`

GetApplication returns the application with the given identifier (app_id)

- Request: [`ApplicationIdentifier`](#handlerapplicationidentifier)
- Response: [`Application`](#handlerapplicationidentifier)

#### HTTP Endpoint

- `GET` `/applications/{app_id}`(`app_id` can be left out of the request body)

#### JSON Request Format

```json
{
  "app_id": "some-app-id"
}
```

#### JSON Response Format

```json
{
  "app_id": "some-app-id",
  "converter": "function Converter(decoded, port) {...",
  "decoder": "function Decoder(bytes, port) {...",
  "encoder": "Encoder(object, port) {...",
  "validator": "Validator(converted, port) {..."
}
```

### `SetApplication`

SetApplication updates the settings for the application. All fields must be supplied.

- Request: [`Application`](#handlerapplication)
- Response: [`Empty`](#handlerapplication)

#### HTTP Endpoints

- `POST` `/applications/{app_id}`(`app_id` can be left out of the request body)
- `PUT` `/applications/{app_id}`(`app_id` can be left out of the request body)

#### JSON Request Format

```json
{
  "app_id": "some-app-id",
  "converter": "function Converter(decoded, port) {...",
  "decoder": "function Decoder(bytes, port) {...",
  "encoder": "Encoder(object, port) {...",
  "validator": "Validator(converted, port) {..."
}
```

#### JSON Response Format

```json
{}
```

### `DeleteApplication`

DeleteApplication deletes the application with the given identifier (app_id)

- Request: [`ApplicationIdentifier`](#handlerapplicationidentifier)
- Response: [`Empty`](#handlerapplicationidentifier)

#### HTTP Endpoint

- `DELETE` `/applications/{app_id}`(`app_id` can be left out of the request body)

#### JSON Request Format

```json
{
  "app_id": "some-app-id"
}
```

#### JSON Response Format

```json
{}
```

### `GetDevice`

GetDevice returns the device with the given identifier (app_id and dev_id)

- Request: [`DeviceIdentifier`](#handlerdeviceidentifier)
- Response: [`Device`](#handlerdeviceidentifier)

#### HTTP Endpoint

- `GET` `/applications/{app_id}/devices/{dev_id}`(`app_id`, `dev_id` can be left out of the request body)

#### JSON Request Format

```json
{
  "app_id": "some-app-id",
  "dev_id": "some-dev-id"
}
```

#### JSON Response Format

```json
{
  "altitude": 0,
  "app_id": "some-app-id",
  "dev_id": "some-dev-id",
  "latitude": 0,
  "longitude": 0,
  "lorawan_device": {
    "activation_constraints": "local",
    "app_eui": "0102030405060708",
    "app_id": "some-app-id",
    "app_key": "01020304050607080102030405060708",
    "app_s_key": "01020304050607080102030405060708",
    "dev_addr": "01020304",
    "dev_eui": "0102030405060708",
    "dev_id": "some-dev-id",
    "disable_f_cnt_check": false,
    "f_cnt_down": 0,
    "f_cnt_up": 0,
    "last_seen": 0,
    "nwk_s_key": "01020304050607080102030405060708",
    "uses32_bit_f_cnt": true
  }
}
```

### `SetDevice`

SetDevice creates or updates a device. All fields must be supplied.

- Request: [`Device`](#handlerdevice)
- Response: [`Empty`](#handlerdevice)

#### HTTP Endpoints

- `POST` `/applications/{app_id}/devices/{dev_id}`(`app_id`, `dev_id` can be left out of the request body)
- `PUT` `/applications/{app_id}/devices/{dev_id}`(`app_id`, `dev_id` can be left out of the request body)
- `POST` `/applications/{app_id}/devices`(`app_id` can be left out of the request body)
- `PUT` `/applications/{app_id}/devices`(`app_id` can be left out of the request body)

#### JSON Request Format

```json
{
  "altitude": 0,
  "app_id": "some-app-id",
  "dev_id": "some-dev-id",
  "latitude": 0,
  "longitude": 0,
  "lorawan_device": {
    "activation_constraints": "local",
    "app_eui": "0102030405060708",
    "app_id": "some-app-id",
    "app_key": "01020304050607080102030405060708",
    "app_s_key": "01020304050607080102030405060708",
    "dev_addr": "01020304",
    "dev_eui": "0102030405060708",
    "dev_id": "some-dev-id",
    "disable_f_cnt_check": false,
    "f_cnt_down": 0,
    "f_cnt_up": 0,
    "last_seen": 0,
    "nwk_s_key": "01020304050607080102030405060708",
    "uses32_bit_f_cnt": true
  }
}
```

#### JSON Response Format

```json
{}
```

### `DeleteDevice`

DeleteDevice deletes the device with the given identifier (app_id and dev_id)

- Request: [`DeviceIdentifier`](#handlerdeviceidentifier)
- Response: [`Empty`](#handlerdeviceidentifier)

#### HTTP Endpoint

- `DELETE` `/applications/{app_id}/devices/{dev_id}`(`app_id`, `dev_id` can be left out of the request body)

#### JSON Request Format

```json
{
  "app_id": "some-app-id",
  "dev_id": "some-dev-id"
}
```

#### JSON Response Format

```json
{}
```

### `GetDevicesForApplication`

GetDevicesForApplication returns all devices that belong to the application with the given identifier (app_id)

- Request: [`ApplicationIdentifier`](#handlerapplicationidentifier)
- Response: [`DeviceList`](#handlerapplicationidentifier)

#### HTTP Endpoint

- `GET` `/applications/{app_id}/devices`(`app_id` can be left out of the request body)

#### JSON Request Format

```json
{
  "app_id": "some-app-id"
}
```

#### JSON Response Format

```json
{
  "devices": [
    {
      "altitude": 0,
      "app_id": "some-app-id",
      "dev_id": "some-dev-id",
      "latitude": 0,
      "longitude": 0,
      "lorawan_device": {
        "activation_constraints": "local",
        "app_eui": "0102030405060708",
        "app_id": "some-app-id",
        "app_key": "01020304050607080102030405060708",
        "app_s_key": "01020304050607080102030405060708",
        "dev_addr": "01020304",
        "dev_eui": "0102030405060708",
        "dev_id": "some-dev-id",
        "disable_f_cnt_check": false,
        "f_cnt_down": 0,
        "f_cnt_up": 0,
        "last_seen": 0,
        "nwk_s_key": "01020304050607080102030405060708",
        "uses32_bit_f_cnt": true
      }
    }
  ]
}
```

### `DryDownlink`

DryUplink simulates processing a downlink message and returns the result

- Request: [`DryDownlinkMessage`](#handlerdrydownlinkmessage)
- Response: [`DryDownlinkResult`](#handlerdrydownlinkmessage)

### `DryUplink`

DryUplink simulates processing an uplink message and returns the result

- Request: [`DryUplinkMessage`](#handlerdryuplinkmessage)
- Response: [`DryUplinkResult`](#handlerdryuplinkmessage)

### `SimulateUplink`

SimulateUplink simulates an uplink message

- Request: [`SimulatedUplinkMessage`](#handlersimulateduplinkmessage)
- Response: [`Empty`](#handlersimulateduplinkmessage)

## Messages

### `.google.protobuf.Empty`

A generic empty message that you can re-use to avoid defining duplicated
empty messages in your APIs.

### `.handler.Application`

The Application settings

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `app_id` | `string` |  |
| `decoder` | `string` | The decoder is a JavaScript function that decodes a byte array to an object. |
| `converter` | `string` | The converter is a JavaScript function that can be used to convert values in the object returned from the decoder. This can for example be useful to convert a voltage to a temperature. |
| `validator` | `string` | The validator is a JavaScript function that checks the validity of the object returned by the decoder or converter. If validation fails, the message is dropped. |
| `encoder` | `string` | The encoder is a JavaScript function that encodes an object to a byte array. |

### `.handler.ApplicationIdentifier`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `app_id` | `string` |  |

### `.handler.Device`

The Device settings

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `app_id` | `string` |  |
| `dev_id` | `string` |  |
| `lorawan_device` | [`Device`](#lorawandevice) |  |
| `latitude` | `float` |  |
| `longitude` | `float` |  |
| `altitude` | `int32` |  |

### `.handler.DeviceIdentifier`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `app_id` | `string` |  |
| `dev_id` | `string` |  |

### `.handler.DeviceList`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `devices` | _repeated_ [`Device`](#handlerdevice) |  |

### `.handler.DryDownlinkMessage`

DryDownlinkMessage is a simulated message to test downlink processing

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `payload` | `bytes` | The binary payload to use |
| `fields` | `string` | JSON-encoded object with fields to encode |
| `app` | [`Application`](#handlerapplication) | The Application containing the payload functions that should be executed |
| `port` | `uint32` | The port number that should be passed to the payload function |

### `.handler.DryDownlinkResult`

DryDownlinkResult is the result from a downlink simulation

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `payload` | `bytes` | The payload that was encoded |
| `logs` | _repeated_ [`LogEntry`](#handlerlogentry) | Logs that have been generated while processing |

### `.handler.DryUplinkMessage`

DryUplinkMessage is a simulated message to test uplink processing

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `payload` | `bytes` | The binary payload to use |
| `app` | [`Application`](#handlerapplication) | The Application containing the payload functions that should be executed |
| `port` | `uint32` | The port number that should be passed to the payload function |

### `.handler.DryUplinkResult`

DryUplinkResult is the result from an uplink simulation

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `payload` | `bytes` | The binary payload |
| `fields` | `string` | The decoded fields |
| `valid` | `bool` | Was validation of the message successful |
| `logs` | _repeated_ [`LogEntry`](#handlerlogentry) | Logs that have been generated while processing |

### `.handler.LogEntry`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `function` | `string` | The location where the log was created (what payload function) |
| `fields` | _repeated_ `string` | A list of JSON-encoded fields that were logged |

### `.handler.SimulatedUplinkMessage`

SimulatedUplinkMessage is a simulated uplink message

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `app_id` | `string` |  |
| `dev_id` | `string` |  |
| `payload` | `bytes` | The binary payload to use |
| `port` | `uint32` | The port number |

### `.lorawan.Device`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `app_eui` | `bytes` | The AppEUI is a unique, 8 byte identifier for the application a device belongs to. |
| `dev_eui` | `bytes` | The DevEUI is a unique, 8 byte identifier for the device. |
| `app_id` | `string` | The AppID is a unique identifier for the application a device belongs to. It can contain lowercase letters, numbers, - and _. |
| `dev_id` | `string` | The DevID is a unique identifier for the device. It can contain lowercase letters, numbers, - and _. |
| `dev_addr` | `bytes` | The DevAddr is a dynamic, 4 byte session address for the device. |
| `nwk_s_key` | `bytes` | The NwkSKey is a 16 byte session key that is known by the device and the network. It is used for routing and MAC related functionality. This key is negotiated during the OTAA join procedure, or statically configured using ABP. |
| `app_s_key` | `bytes` | The AppSKey is a 16 byte session key that is known by the device and the application. It is used for payload encryption. This key is negotiated during the OTAA join procedure, or statically configured using ABP. |
| `app_key` | `bytes` | The AppKey is a 16 byte static key that is known by the device and the application. It is used for negotiating session keys (OTAA). |
| `f_cnt_up` | `uint32` | FCntUp is the uplink frame counter for a device session. |
| `f_cnt_down` | `uint32` | FCntDown is the downlink frame counter for a device session. |
| `disable_f_cnt_check` | `bool` | The DisableFCntCheck option disables the frame counter check. Disabling this makes the device vulnerable to replay attacks, but makes ABP slightly easier. |
| `uses32_bit_f_cnt` | `bool` | The Uses32BitFCnt option indicates that the device keeps track of full 32 bit frame counters. As only the 16 lsb are actually transmitted, the 16 msb will have to be inferred. |
| `activation_constraints` | `string` | The ActivationContstraints are used to allocate a device address for a device (comma-separated). There are different prefixes for `otaa`, `abp`, `world`, `local`, `private`, `testing`. |
| `last_seen` | `int64` | When the device was last seen (Unix nanoseconds) |

