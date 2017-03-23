# API Reference

The Things Network's backend servers.

**Options**

```
      --allow-insecure             Allow insecure fallback if TLS unavailable
      --auth-token string          The JWT token to be used for the discovery server
      --config string              config file (default "$HOME/.ttn.yml")
      --description string         The description of this component
      --discovery-address string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string       Location of Elasticsearch server for logging
      --health-port int            The port number where the health server should be started
      --id string                  The id of this component
      --key-dir string             The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string            Location of the log file
      --no-cli-logs                Disable CLI logs
      --public                     Announce this component as part of The Things Network (public community network)
      --tls                        Use TLS (default true)
```


## ttn broker



**Usage:** `ttn broker`

**Options**

```
      --deduplication-delay int          Deduplication delay (in ms) (default 200)
      --networkserver-address string     Networkserver host and port (default "localhost:1903")
      --networkserver-cert string        Networkserver certificate to use
      --networkserver-token string       Networkserver token to use
      --server-address string            The IP address to listen for communication (default "0.0.0.0")
      --server-address-announce string   The public IP address to announce (default "localhost")
      --server-port int                  The port for communication (default 1902)
```

### ttn broker gen-cert

ttn gen-cert generates a TLS Certificate

**Usage:** `ttn broker gen-cert`

### ttn broker gen-keypair

ttn gen-keypair generates a public/private keypair

**Usage:** `ttn broker gen-keypair`

### ttn broker register-prefix

ttn broker register prefix registers a prefix to this Broker

**Usage:** `ttn broker register-prefix [prefix ...]`

## ttn discovery



**Usage:** `ttn discovery`

**Options**

```
      --cache                             Add a cache in front of the database
      --http-address string               The IP address where the gRPC proxy should listen (default "0.0.0.0")
      --http-port int                     The port where the gRPC proxy should listen (default 8080)
      --master-auth-servers stringSlice   Auth servers that are allowed to manage this network (default [ttn-account-v2])
      --redis-address string              Redis server and port (default "localhost:6379")
      --redis-db int                      Redis database
      --redis-password string             Redis password
      --server-address string             The IP address to listen for communication (default "0.0.0.0")
      --server-port int                   The port for communication (default 1900)
```

### ttn discovery gen-cert

ttn gen-cert generates a TLS Certificate

**Usage:** `ttn discovery gen-cert`

### ttn discovery gen-keypair

ttn gen-keypair generates a public/private keypair

**Usage:** `ttn discovery gen-keypair`

## ttn handler



**Usage:** `ttn handler`

**Options**

```
      --amqp-address string              AMQP host and port. Leave empty to disable AMQP
      --amqp-address-announce string     AMQP address to announce (takes value of server-address-announce if empty while enabled)
      --amqp-exchange string             AMQP exchange (default "ttn.handler")
      --amqp-password string             AMQP password (default "guest")
      --amqp-username string             AMQP username (default "guest")
      --broker-id string                 The ID of the TTN Broker as announced in the Discovery server (default "dev")
      --http-address string              The IP address where the gRPC proxy should listen (default "0.0.0.0")
      --http-port int                    The port where the gRPC proxy should listen (default 8084)
      --mqtt-address string              MQTT host and port. Leave empty to disable MQTT
      --mqtt-address-announce string     MQTT address to announce (takes value of server-address-announce if empty while enabled)
      --mqtt-password string             MQTT password
      --mqtt-username string             MQTT username
      --redis-address string             Redis host and port (default "localhost:6379")
      --redis-db int                     Redis database
      --redis-password string            Redis password
      --server-address string            The IP address to listen for communication (default "0.0.0.0")
      --server-address-announce string   The public IP address to announce (default "localhost")
      --server-port int                  The port for communication (default 1904)
```

### ttn handler gen-cert

ttn gen-cert generates a TLS Certificate

**Usage:** `ttn handler gen-cert`

### ttn handler gen-keypair

ttn gen-keypair generates a public/private keypair

**Usage:** `ttn handler gen-keypair`

## ttn networkserver



**Usage:** `ttn networkserver`

**Options**

```
      --net-id int                       LoRaWAN NetID (default 19)
      --redis-address string             Redis server and port (default "localhost:6379")
      --redis-db int                     Redis database
      --redis-password string            Redis password
      --server-address string            The IP address to listen for communication (default "0.0.0.0")
      --server-address-announce string   The public IP address to announce (default "localhost")
      --server-port int                  The port for communication (default 1903)
```

### ttn networkserver authorize

ttn networkserver authorize generates a token that Brokers should use to connect

**Usage:** `ttn networkserver authorize [id]`

**Options**

```
      --valid int   The number of days the token is valid
```

### ttn networkserver gen-cert

ttn gen-cert generates a TLS Certificate

**Usage:** `ttn networkserver gen-cert`

### ttn networkserver gen-keypair

ttn gen-keypair generates a public/private keypair

**Usage:** `ttn networkserver gen-keypair`

## ttn router



**Usage:** `ttn router`

**Options**

```
      --mqtt-address-announce string     MQTT address to announce
      --server-address string            The IP address to listen for communication (default "0.0.0.0")
      --server-address-announce string   The public IP address to announce (default "localhost")
      --server-port int                  The port for communication (default 1901)
      --skip-verify-gateway-token        Skip verification of the gateway token
```

### ttn router gen-cert

ttn gen-cert generates a TLS Certificate

**Usage:** `ttn router gen-cert`

### ttn router gen-keypair

ttn gen-keypair generates a public/private keypair

**Usage:** `ttn router gen-keypair`

## ttn selfupdate

ttn selfupdate updates the current ttn to the latest version

**Usage:** `ttn selfupdate`

## ttn version

ttn version gets the build and version information of ttn

**Usage:** `ttn version`

