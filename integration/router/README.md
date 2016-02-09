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

docker run -p 33000:33000/udp -p 4000:4000 thethingsnetwork/router --udp-port 33000 --tcp-port 4000 --brokers "10.10.10.40:3000,broker.host.com:3000"
```
