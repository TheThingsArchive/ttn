# Authentication

Currently, there are two methods of authenticating to the gRPC and HTTP APIs:

- Bearer token: OAuth 2.0 Bearer JSON Web Tokens (preferred)
- Access keys: Application access keys (only for `ApplicationManager` API)

## Bearer Token

This authentication method is the preferred method of authenticating. 

### gRPC

You can authenticate to the gRPC endpoint by supplying a `token` field in the Metadata. The value of this field should be the JSON Web Token. 

**Example (Go):**

```go
md := metadata.Pairs(
  "token", "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhcHBzIjp7InRlc3QiOlsic2V0dGluZ3MiXX19.VGhpcyBpcyB0aGUgc2lnbmF0dXJl",
)
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

### HTTP

For HTTP Endpoints, you should supply the `Authorization` header: `Authorization: Bearer <token>`.

**Example:**

```
Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhcHBzIjp7InRlc3QiOlsic2V0dGluZ3MiXX19.VGhpcyBpcyB0aGUgc2lnbmF0dXJl
```

## Access Key

With this authentication method, the server will exchange an Access Key for a Bearer Token internally.

### gRPC

You can authenticate to the gRPC endpoint by supplying a `key` field in the Metadata. The value of this field should be the Application access key. 

**Example (Go):**

```go
md := metadata.Pairs(
  "key", "ttn-account-v2.n4BAoKOGuK2hj7MXg_OVtpLO0BTJI8lLzt66UsvTlUvZPsi6FADOptnmSH3e3PuQzbLLEUhXxYhkxr34xyUqBQ",
)
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

### HTTP

For HTTP Endpoints, you should supply the `Authorization` header: `Authorization: Key <key>`.

**Example:**

```
Authorization: Key ttn-account-v2.n4BAoKOGuK2hj7MXg_OVtpLO0BTJI8lLzt66UsvTlUvZPsi6FADOptnmSH3e3PuQzbLLEUhXxYhkxr34xyUqBQ
```
