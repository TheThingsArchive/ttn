# Authentication

Currently, there are two methods of authenticating to the gRPC APIs:

- Bearer token: OAuth 2.0 Bearer JSON Web Tokens (preferred)
- Access keys: Application access keys (only for `ApplicationManager` API)

## Bearer Token

This authentication method is the preferred method of authenticating. 

_TODO: Add reference to Account Server reference_

You can authenticate to the gRPC endpoint by supplying a `token` field in the Metadata. The value of this field should be the JSON Web Token. 

For HTTP Endpoints, you should supply the `Authorization` header: `Authorization: Bearer <token>`.

## Access Key

With this authentication method, the server will exchange an Access Key for a Bearer Token internally.

_TODO: Add reference to Account Server reference_

You can authenticate to the gRPC endpoint by supplying a `key` field in the Metadata. The value of this field should be the Application access key. 

For HTTP Endpoints, you should supply the `Authorization` header: `Authorization: Key <key>`.
