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

Get all announcements for a specific service

- Request: [`GetAllRequest`](#discoverygetallrequest)
- Response: [`AnnouncementsResponse`](#discoverygetallrequest)

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
      "api_address": "http://eu.thethings.network:8084",
      "certificate": "-----BEGIN CERTIFICATE-----\n...",
      "description": "",
      "id": "ttn-handler-eu",
      "metadata": [
        {
          "key": "APP_ID",
          "value": "c29tZS1hcHAtaWQ="
        }
      ],
      "net_address": "eu.thethings.network:1904",
      "public": true,
      "public_key": "-----BEGIN PUBLIC KEY-----\n...",
      "service_name": "handler",
      "service_version": "2.0.0-dev-abcdef...",
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
  "api_address": "http://eu.thethings.network:8084",
  "certificate": "-----BEGIN CERTIFICATE-----\n...",
  "description": "",
  "id": "ttn-handler-eu",
  "metadata": [
    {
      "key": "APP_ID",
      "value": "c29tZS1hcHAtaWQ="
    }
  ],
  "net_address": "eu.thethings.network:1904",
  "public": true,
  "public_key": "-----BEGIN PUBLIC KEY-----\n...",
  "service_name": "handler",
  "service_version": "2.0.0-dev-abcdef...",
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
| `net_address` | `string` | Comma-separated network addresses in the form "[hostname]:[port]" (currently we only use the first) |
| `public_key` | `string` | ECDSA public key of this component |
| `certificate` | `string` | TLS Certificate (if TLS is enabled) |
| `api_address` | `string` | Contains the address where the HTTP API is exposed (if there is one) |
| `metadata` | _repeated_ [`Metadata`](#discoverymetadata) | Metadata for this component |

### `.discovery.AnnouncementsResponse`

A list of announcements

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `services` | _repeated_ [`Announcement`](#discoveryannouncement) |  |

### `.discovery.GetAllRequest`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `service_name` | `string` | The name of the service (router/broker/handler) |

### `.discovery.GetRequest`

The identifier of the service that should be returned

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `id` | `string` | The ID of the service |
| `service_name` | `string` | The name of the service (router/broker/handler) |

### `.discovery.Metadata`

Announcements have a list of Metadata

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `key` | [`Key`](#discoverymetadatakey) | The key indicates the metadata type |
| `value` | `bytes` | The value depends on the key type |

### `.discovery.MetadataRequest`

The metadata to add or remove from an announement

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `id` | `string` | The ID of the service that should be modified |
| `service_name` | `string` | The name of the service (router/broker/handler) that should be modified |
| `metadata` | [`Metadata`](#discoverymetadata) |  |

### `.google.protobuf.Empty`

A generic empty message that you can re-use to avoid defining duplicated
empty messages in your APIs.

## Used Enums

### `.discovery.Metadata.Key`

The Key indicates the metadata type

| Value | Description |
| ----- | ----------- |
| `OTHER` | OTHER indicates arbitrary metadata. We currently don't allow this. |
| `PREFIX` | The value for PREFIX consists of 1 byte denoting the number of bits, followed by the prefix and enough trailing bits to fill 4 octets. Only authorized brokers can announce PREFIX metadata. |
| `APP_EUI` | APP_EUI is used for announcing join handlers. The value for APP_EUI is the byte slice of the AppEUI. Only authorized join handlers can announce APP_EUI metadata (and we don't have any of those yet). |
| `APP_ID` | APP_ID is used for announcing that this handler is responsible for a certain AppID. The value for APP_ID is the byte slice of the AppID string. This metadata can only be added if the requesting client is authorized to manage this AppID. |

