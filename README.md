# The Things Network Stack

[![Build Status](https://travis-ci.org/TheThingsNetwork/ttn.svg?branch=master)](https://travis-ci.org/TheThingsNetwork/ttn) [![Coverage Status](https://coveralls.io/repos/github/TheThingsNetwork/ttn/badge.svg?branch=master)](https://coveralls.io/github/TheThingsNetwork/ttn?branch=master) [![GoDoc](https://godoc.org/github.com/TheThingsNetwork/ttn?status.svg)](https://godoc.org/github.com/TheThingsNetwork/ttn)

![The Things Network](https://thethings.blob.core.windows.net/ttn/logo.svg)

The Things Network is a global open crowdsourced Internet of Things data network.

## Getting Started With The Things Network

When you get started with The Things Network, you'll probably have some questions. Here are some things you can do to find the answer to them:

- Check out our [website](https://www.thethingsnetwork.org/)
- Read the [official documentation](https://www.thethingsnetwork.org/docs/)
- Register on the [forum](https://www.thethingsnetwork.org/forum/) and search around
- Join [Slack](https://slack.thethingsnetwork.org) and ask us what you want to know
- Read background information on the [wiki](https://www.thethingsnetwork.org/wiki/)

## Installing and Running The Things Network Stack

Although we're all about building an open, public network, we understand that some people rather have everything privately on their own servers. On our website, you'll find some articles describing how you can [set up a private routing environment](https://www.thethingsnetwork.org/article/setting-up-a-private-routing-environment) and how you can [deploy this environment using Docker](https://www.thethingsnetwork.org/article/deploying-a-private-routing-environment-with-docker-compose).

## Development

First, you'll have to prepare your development environment. Follow the steps below to set up your development machine.

1. Make sure you have [Go](https://golang.org) installed (version 1.8 or later).
2. Set up your [Go environment](https://golang.org/doc/code.html#GOPATH)
3. Install the [protobuf compiler (`protoc`)](https://github.com/google/protobuf/releases)
4. Install `make`. On Linux install `build-essential`. On macOS, `make` comes with XCode or the developer tools. On Windows you can get `make` from [https://gnuarmeclipse.github.io/windows-build-tools/](https://gnuarmeclipse.github.io/windows-build-tools/)
5. Make sure you have [Redis](http://redis.io/download) and [RabbitMQ](https://www.rabbitmq.com/download.html) **installed** and **running**.  
  On a fresh installation you might need to install the [MQTT plugin for RabbitMQ](https://www.rabbitmq.com/mqtt.html).  
  If you're on Linux, you probably know how to do that. On a Mac, just run `brew bundle`. The Windows installer will setup and start RabbitMQ as a service. Use the `RabbitMQ Command Prompt (sbin dir)` to run commands, i.e. to enable plugins.
6. Declare a RabbitMQ exchange `ttn.handler` of type `topic`. Using [the management plugin](http://www.rabbitmq.com/management.html), declare the exchange in the web interface `http://server-name:15672` or using the management cli, run `rabbitmqadmin declare exchange name=ttn.handler type=topic auto_delete=false durable=true`. If your handler's user has sufficient permissions on RabbitMQ, it will attempt to create the exchange if not present.

Next, you can clone this repository and set up the TTN part:

1. Fork this repository
2. Clone your fork: `git clone --branch develop https://github.com/YOURUSERNAME/ttn.git $GOPATH/src/github.com/TheThingsNetwork/ttn`
3. `cd $GOPATH/src/github.com/TheThingsNetwork/ttn`
4. Install the dependencies for development: `make dev-deps`
5. Run the tests: `make test`
6. Run `make build` to build both `ttn` and `ttnctl` from source. 
7. Run `make dev` to install the go binaries into `$GOPATH/bin/`
    * Optionally on Linux or Mac you can use `make link` to link them to `$GOPATH/bin/` (In order to run the commands, you should have `export PATH="$GOPATH/bin:$PATH"` in your profile).
8. Configure your `ttnctl` with the settings in `.env/ttnctl.yml.dev-example` by copying that file to `~/.ttnctl.yml`.
9. Trust the CA certificate of your local discovery server by copying `.env/discovery/server.cert` to `~/.ttnctl/ca.cert`.

You can check your `ttnctl` configuration by running `ttnctl config`. It should look like this:

```
  INFO Using config:

         config file: /home/your-user/.ttnctl.yml
            data dir: /home/your-user/.ttnctl

         auth-server: https://account.thethingsnetwork.org
   discovery-address: localhost:1900
           router-id: dev
          handler-id: dev
        mqtt-address: localhost:1883
```

**NOTE:** From now on you should run all commands from the `$GOPATH/src/github.com/TheThingsNetwork/ttn` directory.

## License

Source code for The Things Network is released under the MIT License, which can be found in the [LICENSE](LICENSE) file. A list of authors can be found in the [AUTHORS](AUTHORS) file.
