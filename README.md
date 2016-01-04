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
-----| rtr-brk-http
----------> Adapter

-----| gtw-rtr-udp
----------> Adapter

-----| brk-nwc-local
----------> Adapter

-----| brk-hdl-http
----------> Adapter

-----| brk-rtr-http
----------> Adapter

-----| ns-brk-local
----------> Adapter

-----| hdl-brk-http
----------> Adapter

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

-| testing
-----| mock_components
----------> Router
----------> Broker
----------> Networkcontroller
----------> Handler

-----| mock_adapters
----------| rtr-brk-mock
---------------> Adapter

----------| gtw-rtr-mock
---------------> Adapter

----------| brk-nwc-mock
---------------> Adapter

----------| brk-hdl-mock
---------------> Adapter

----------| brk-rtr-mock
---------------> Adapter

----------| ns-brk-mock
---------------> Adapter

----------| hdl-brk-mock
---------------> Adapter
```

## Development Plan

See the [development plan](DEVELOPMENT_PLAN.md)

## Authors

See the [author's list](AUTHORS)

## License

See the [license file](LICENSE)
