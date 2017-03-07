# Discovery API Reference

The Discovery service is used to discover services within The Things Network.

## Methods

### `Announce`

Announce a component to the Discovery server.
A call to `Announce` does not processes the `metadata` field, so you can safely leave this field empty.
Adding or removing Metadata should be done with the `AddMetadata` and `DeleteMetadata` methods.

- Request: [`Announcement`](#discoveryannouncement)
- Response: [`Empty`](#discoveryannouncement)

### `GetAll`

Get all announcements for a specific service type

- Request: [`GetServiceRequest`](#discoverygetservicerequest)
- Response: [`AnnouncementsResponse`](#discoverygetservicerequest)

#### HTTP Endpoint

- `GET` `/announcements/{service_name}`(`service_name` can be left out of the request body)

#### JSON Request Format

```json
{
  "service_name": "handler"
}
```

#### JSON Response Format

```json
{
  "services": [
    {
      "amqp_address": "",
      "api_address": "http://eu.thethings.network:8084",
      "certificate": "-----BEGIN CERTIFICATE-----\n...",
      "description": "",
      "id": "ttn-handler-eu",
      "metadata": [
        {
          "app_eui": "",
          "app_id": "some-app-id",
          "dev_addr_prefix": "AAAAAAA="
        }
      ],
      "mqtt_address": "",
      "net_address": "eu.thethings.network:1904",
      "public": true,
      "public_key": "-----BEGIN PUBLIC KEY-----\n...",
      "service_name": "handler",
      "service_version": "2.0.0-abcdef...",
      "url": ""
    }
  ]
}
```

### `Get`

Get a specific announcement

- Request: [`GetRequest`](#discoverygetrequest)
- Response: [`Announcement`](#discoverygetrequest)

#### HTTP Endpoint

- `GET` `/announcements/{service_name}/{id}`(`service_name`, `id` can be left out of the request body)

#### JSON Request Format

```json
{
  "id": "ttn-handler-eu",
  "service_name": "handler"
}
```

#### JSON Response Format

```json
{
  "amqp_address": "",
  "api_address": "http://eu.thethings.network:8084",
  "certificate": "-----BEGIN CERTIFICATE-----\n...",
  "description": "",
  "id": "ttn-handler-eu",
  "metadata": [
    {
      "app_eui": "",
      "app_id": "some-app-id",
      "dev_addr_prefix": "AAAAAAA="
    }
  ],
  "mqtt_address": "",
  "net_address": "eu.thethings.network:1904",
  "public": true,
  "public_key": "-----BEGIN PUBLIC KEY-----\n...",
  "service_name": "handler",
  "service_version": "2.0.0-abcdef...",
  "url": ""
}
```

### `AddMetadata`

Add metadata to an announement

- Request: [`MetadataRequest`](#discoverymetadatarequest)
- Response: [`Empty`](#discoverymetadatarequest)

### `DeleteMetadata`

Delete metadata from an announcement

- Request: [`MetadataRequest`](#discoverymetadatarequest)
- Response: [`Empty`](#discoverymetadatarequest)

## Messages

### `.discovery.Announcement`

The Announcement of a service (also called component)

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `id` | `string` | The ID of the component |
| `service_name` | `string` | The name of the component (router/broker/handler) |
| `service_version` | `string` | Service version in the form "[version]-[commit] ([build date])" |
| `description` | `string` | Description of the component |
| `url` | `string` | URL with documentation or more information about this component |
| `public` | `bool` | Indicates whether this service is part of The Things Network (the public community network) |
| `net_address` | `string` | Comma-separated network addresses in the form "domain1:port,domain2:port,domain3:port" (currently we only use the first) |
| `public_key` | `string` | ECDSA public key of this component |
| `certificate` | `string` | TLS Certificate for gRPC on net_address (if TLS is enabled) |
| `api_address` | `string` | Contains the address where the HTTP API is exposed (if there is one). Format: "http(s)://domain(:port)"; default http port is 80, default https port is 443. |
| `mqtt_address` | `string` | Contains the address where the MQTT API is exposed (if there is one). Format: "domain(:port)"; if no port supplied, mqtt is on 1883, mqtts is on 8883. |
| `amqp_address` | `string` | Contains the address where the AMQP API is exposed (if there is one). Format: "domain(:port)"; if no port supplied, amqp is on 5672, amqps is on 5671. |
| `metadata` | _repeated_ [`Metadata`](#discoverymetadata) | Metadata for this component |

### `.discovery.AnnouncementsResponse`

A list of announcements

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `services` | _repeated_ [`Announcement`](#discoveryannouncement) |  |

### `.discovery.GetRequest`

The identifier of the service that should be returned

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `id` | `string` | The ID of the service |
| `service_name` | `string` | The name of the service (router/broker/handler) |

### `.discovery.GetServiceRequest`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `service_name` | `string` | The name of the service (router/broker/handler) |

### `.discovery.Metadata`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| **metadata** | **oneof 3** | one of the following 3 |
| `dev_addr_prefix` | `bytes` | DevAddr prefix that is routed by this Broker 5 bytes; the first byte is the prefix length, the following 4 bytes are the address. Only authorized Brokers can announce PREFIX metadata. |
| `app_id` | `string` | AppID that is registered to this Handler This metadata can only be added if the requesting client is authorized to manage this AppID. |
| `app_eui` | `bytes` | AppEUI that is registered to this Join Handler Only authorized Join Handlers can announce APP_EUI metadata (and we don't have any of those yet). |

### `.discovery.MetadataRequest`

The metadata to add or remove from an announement

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `id` | `string` | The ID of the service that should be modified |
| `service_name` | `string` | The name of the service (router/broker/handler) that should be modified |
| `metadata` | [`Metadata`](#discoverymetadata) | Metadata to add or remove |

### `.google.protobuf.Empty`

A generic empty message that you can re-use to avoid defining duplicated
empty messages in your APIs.

