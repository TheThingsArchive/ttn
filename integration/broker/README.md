# The Things Network Broker

[![Build Status](https://travis-ci.org/TheThingsNetwork/ttn.svg?branch=develop)](https://travis-ci.org/TheThingsNetwork/ttn) [![Slack Status](https://slack.thethingsnetwork.org/badge.svg)](https://slack.thethingsnetwork.org/)

![The Things Network](http://thethingsnetwork.org/static/ttn/media/The%20Things%20Uitlijning.svg)

The Things Network is a global open crowdsourced Internet of Things data network.

## Status

This image is a pre-1.0 version of The Things Network's Broker component. It is **under heavy development** and currently it's APIs and code are not yet stable.

## Tags

* [`latest`, `develop` (TheThingsNetwork/ttn - develop)](https://github.com/TheThingsNetwork/ttn/blob/develop/integration/broker/Dockerfile)

## Usage

```
docker pull thethingsnetwork/broker

docker run -p 3000:3000 -p 4000:4000 thethingsnetwork/broker --handlers-port 3000 --routers-port 4000
```
