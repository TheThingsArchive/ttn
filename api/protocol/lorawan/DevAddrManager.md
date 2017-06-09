# DevAddrManager API Reference

The Device Addresses in the network are issued by the NetworkServer

## Methods

### `GetPrefixes`

Get all prefixes that are in use or available

- Request: [`PrefixesRequest`](#lorawanprefixesrequest)
- Response: [`PrefixesResponse`](#lorawanprefixesrequest)

### `GetDevAddr`

Request a device address

- Request: [`DevAddrRequest`](#lorawandevaddrrequest)
- Response: [`DevAddrResponse`](#lorawandevaddrrequest)

## Messages

### `.lorawan.DevAddrRequest`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `usage` | _repeated_ `string` | The usage constraints (see activation_constraints in lorawan.proto) |

### `.lorawan.DevAddrResponse`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `dev_addr` | `bytes` |  |

### `.lorawan.PrefixesRequest`

### `.lorawan.PrefixesResponse`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `prefixes` | _repeated_ [`PrefixMapping`](#lorawanprefixesresponseprefixmapping) | The prefixes that are in use or available |

### `.lorawan.PrefixesResponse.PrefixMapping`

| Field Name | Type | Description |
| ---------- | ---- | ----------- |
| `prefix` | `string` | The prefix that can be used |
| `usage` | _repeated_ `string` | Usage constraints of this prefix (see activation_constraints in lorawan.proto) |

