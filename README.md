The Things Network Core Architecture
====================================

## How to Contribute

- Open an issue and explain on what you're working
- Work in a separate branch forked from `develop`
- Bring your feature to life and make a PR
- Do not commit on `master` anymore

## Folder structure

So far: 
```
-| components
-----| broker
-----| handler
-----| networkcontroller
-----| router

-| adapters
-----| router-broker-http/brokAdapter
-----| router-gateway-udp/gateAdapter
-----| broker-ns-local/nsAdapter
-----| broker-handler-http/handAdapter
-----| broker-router-http/routAdapter
-----| ns-broker-local/brokAdapter
-----| handler-broker-http/brokAdapter

-| lorawan
-----| mac
-----| gateway
---------| protocol

-| simulators
-----| gateway
```

## Development Plan

See the [development plan](DEVELOPMENT_PLAN.md)

## Authors

See the [author's list](AUTHORS)

## License

See the [license file](LICENSE)
