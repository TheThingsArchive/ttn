# HTTP API Reference

The Handler HTTP API is a wrapper around the Handler's gRPC interface. We recommend everyone to use the gRPC interface if possible.

## Authorization

Authorization to the Handler HTTP API is done with the `Authorization` header in your requests.
This header should either contain a `Bearer` token with the JSON Web Token issued by the account server or a `Key` that is issued by the account server and can be exchanged to a JSON Web Token.

Example (`Bearer`):

```
Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhcHBzIjp7InRlc3QiOlsic2V0dGluZ3MiXX19.VGhpcyBpcyB0aGUgc2lnbmF0dXJl
```

Example (`Key`):

```
Authorization: Key ttn-account-preview.n4BAoKOGuK2hj7MXg_OVtpLO0BTJI8lLzt66UsvTlUvZPsi6FADOptnmSH3e3PuQzbLLEUhXxYhkxr34xyUqBQ
```

## Register Application

Request:

```
POST /applications

{
  "app_id": "the-app-id"
}
```

Response:

```
200 OK

{}
```

## Get Application

Request:

```
GET /applications/the-app-id
```

Response:

```
200 OK

{
  "app_id": "the-app-id",
  "decoder": "function Decoder(...) { ... }"
  "converter": "function Converter(...) { ... }"
  "validator": "function Validator(...) { ... }"
  "encoder": "function Encoder(...) { ... }"
}
```

## Get Devices For Application

Request:

```
GET /applications/the-app-id/devices
```

Response:

```
200 OK

{
  "devices": [
    {
      "app_id": "the-app-id",
      "dev_id": "the-dev-id",
      "lorawan_device": {
        "app_eui": "70B3D57EF0000001",
        "dev_eui": "70B3D57EF0000001",
        "app_id": "the-app-id",
        "dev_id": "the-dev-id",
        "dev_addr": "",
        "nwk_s_key": "",
        "app_s_key": "",
        "app_key": "01020304050607080102030405060708"
      }
    },
    ...
  ]
}
```

## Set Application (after it has been registered)

Request:

```
POST /applications/the-app-id
{
  "decoder": "function Decoder(...) { ... }"
  "converter": "function Converter(...) { ... }"
  "validator": "function Validator(...) { ... }"
  "encoder": "function Encoder(...) { ... }"
}
```

Response:

```
200 OK

{}
```

## Delete Application

Request:

```
DELETE /applications/the-app-id
```

Response:

```
200 OK

{}
```

## Get Device

```
GET /applications/the-app-id/devices/the-dev-id
```

Response:

```
200 OK

{
  "app_id": "the-app-id",
  "dev_id": "the-dev-id",
  "lorawan_device": {
    "app_eui": "70B3D57EF0000001",
    "dev_eui": "70B3D57EF0000001",
    "app_id": "the-app-id",
    "dev_id": "the-dev-id",
    "dev_addr": "",
    "nwk_s_key": "",
    "app_s_key": "",
    "app_key": "01020304050607080102030405060708"
  }
}
```

## Set Device

```
POST /applications/the-app-id/devices/the-dev-id

{
  "lorawan_device": {
    "app_eui": "70B3D57EF0000001",
    "dev_eui": "70B3D57EF0000001",
    "dev_addr": "26000001",
    "nwk_s_key": "01020304050607080102030405060708",
    "app_s_key": "01020304050607080102030405060708",
    "app_key": ""
  }
}
```

Response:

```
200 OK

{}
```

## Delete Device

Request:

```
DELETE /applications/the-app-id/devices/the-dev-id
```

Response:

```
200 OK

{}
```

## Types

### Application

```
app_id     string
decoder    string
converter  string
validator  string
encoder    string
```

### Device

```
app_id  string
dev_id  string

lorawan_device:
  app_eui              string
  dev_eui              string
  dev_addr             string
  nwk_s_key            string
  app_s_key            string
  app_key              string
  f_cnt_up             int
  f_cnt_down           int
  disable_f_cnt_check  bool
  uses32_bit_f_cnt     bool
  last_seen            int (unix-nanoseconds)
```
