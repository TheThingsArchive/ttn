The Things Network Core Architecture
====================================

[![Slack Status](https://slack.thethingsnetwork.org/badge.svg)](https://slack.thethingsnetwork.org/)

## How to Contribute

- Open an issue and explain on what you're working
- Work in a separate branch forked from `develop`
- Bring your feature to life and make a PR
- Do not commit on `master` anymore

## Folder structure

So far: 

```
-> Router
-> GatewayRouterAdapter
-> RouterBrokerAdapter
-> BrokerAddress
-> GatewayAddress

-| components
-----> Router
-----> Broker
-----> Networkcontroller
-----> Handler

-| adapters
-----| http
----------> Adapter
----------| pubsub
---------------> Adapter
----------| broadcast
---------------> Adapter
-----| semtech
---------------> Adapter

-| lorawan
-----| mac
-----| semtech
----------> Payload
----------> Packet
----------> DeviceAddress
----------> RXPK
----------> TXPK
----------> Stat

-| simulators
-----| gateway
----------> Forwarder
----------> Imitator
```

## Development Plan

See the [development plan](DEVELOPMENT_PLAN.md)

## Authors

See the [author's list](AUTHORS)

## License

See the [license file](LICENSE)
