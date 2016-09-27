## ttn

The Things Network's backend servers

### Synopsis


ttn launches The Things Network's backend servers

### Options

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn broker

The Things Network broker

### Synopsis


The Things Network broker

```
ttn broker
```

### Options

```
      --deduplication-delay int          Deduplication delay (in ms) (default 200)
      --networkserver-address string     Networkserver host and port (default "localhost:1903")
      --networkserver-cert string        Networkserver certificate to use
      --networkserver-token string       Networkserver token to use
      --server-address string            The IP address to listen for communication (default "0.0.0.0")
      --server-address-announce string   The public IP address to announce (default "localhost")
      --server-port int                  The port for communication (default 1902)
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn broker genkeys

Generate keys and certificate

### Synopsis


ttn genkeys generates keys and a TLS certificate for this component

```
ttn broker genkeys
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn broker register-prefix

Register a prefix to this Broker

### Synopsis


ttn broker register prefix registers a prefix to this Broker

```
ttn broker register-prefix [prefix ...]
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn discovery

The Things Network discovery

### Synopsis


The Things Network discovery

```
ttn discovery
```

### Options

```
      --redis-address string    Redis server and port (default "localhost:6379")
      --redis-db int            Redis database
      --server-address string   The IP address to listen for communication (default "0.0.0.0")
      --server-port int         The port for communication (default 1900)
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn handler

The Things Network handler

### Synopsis


The Things Network handler

```
ttn handler
```

### Options

```
      --mqtt-broker string               MQTT broker host and port (default "localhost:1883")
      --mqtt-password string             MQTT password
      --mqtt-username string             MQTT username (default "handler")
      --redis-address string             Redis host and port (default "localhost:6379")
      --redis-db int                     Redis database
      --server-address string            The IP address to listen for communication (default "0.0.0.0")
      --server-address-announce string   The public IP address to announce (default "localhost")
      --server-port int                  The port for communication (default 1904)
      --ttn-broker string                The ID of the TTN Broker as announced in the Discovery server (default "dev")
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn handler genkeys

Generate keys and certificate

### Synopsis


ttn genkeys generates keys and a TLS certificate for this component

```
ttn handler genkeys
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn networkserver

The Things Network networkserver

### Synopsis


The Things Network networkserver

```
ttn networkserver
```

### Options

```
      --net-id int                       LoRaWAN NetID (default 19)
      --redis-address string             Redis server and port (default "localhost:6379")
      --redis-db int                     Redis database
      --server-address string            The IP address to listen for communication (default "0.0.0.0")
      --server-address-announce string   The public IP address to announce (default "localhost")
      --server-port int                  The port for communication (default 1903)
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn networkserver authorize

Generate a token that Brokers should use to connect

### Synopsis


ttn networkserver authorize generates a token that Brokers should use to connect

```
ttn networkserver authorize [id]
```

### Options

```
      --valid int   The number of days the token is valid
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn networkserver genkeys

Generate keys and certificate

### Synopsis


ttn genkeys generates keys and a TLS certificate for this component

```
ttn networkserver genkeys
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn router

The Things Network router

### Synopsis


The Things Network router

```
ttn router
```

### Options

```
      --server-address string            The IP address to listen for communication (default "0.0.0.0")
      --server-address-announce string   The public IP address to announce (default "localhost")
      --server-port int                  The port for communication (default 1901)
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn router genkeys

Generate keys and certificate

### Synopsis


ttn genkeys generates keys and a TLS certificate for this component

```
ttn router genkeys
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


## ttn version

Get build and version information

### Synopsis


ttn version gets the build and version information of ttn

```
ttn version
```

### Options inherited from parent commands

```
      --auth-token string         The JWT token to be used for the discovery server
      --config string             config file (default "$HOME/.ttn.yaml")
      --description string        The description of this component
      --discovery-server string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --elasticsearch string      Location of Elasticsearch server for logging
      --health-port int           The port number where the health server should be started
      --id string                 The id of this component
      --key-dir string            The directory where public/private keys are stored (default "$HOME/.ttn")
      --log-file string           Location of the log file
      --no-cli-logs               Disable CLI logs
      --tls                       Use TLS
```


