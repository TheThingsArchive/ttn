# The Things Network Router

[![Build Status](https://travis-ci.org/TheThingsNetwork/ttn.svg?branch=develop)](https://travis-ci.org/TheThingsNetwork/ttn) [![Slack Status](https://slack.thethingsnetwork.org/badge.svg)](https://slack.thethingsnetwork.org/)

![The Things Network](http://thethingsnetwork.org/static/ttn/media/The%20Things%20Uitlijning.svg)

The Things Network is a global open crowdsourced Internet of Things data network.

## Status

This image is a pre-1.0 version of The Things Network's Router component. It is **under heavy development** and currently it's APIs and code are not yet stable.

## Tags

* [`latest`, `develop` (TheThingsNetwork/ttn - develop)](https://github.com/TheThingsNetwork/ttn/blob/develop/integration/router/Dockerfile)

## Usage

```
docker pull thethingsnetwork/router

docker run -p 8647:8647/udp -p 8627:8627 thethingsnetwork/router --udp-port 8647 --tcp-port 8627 --brokers "1.2.3.4:8672,broker.host.com:8672"
```
