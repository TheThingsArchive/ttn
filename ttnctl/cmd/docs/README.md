# API Reference

Control The Things Network from the command line.

## ttnctl applications

Manage applications

### Synopsis


ttnctl applications can be used to manage applications.

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications create

Create a new application

### Synopsis


ttnctl applications create can be used to create a new application.

```
ttnctl applications create [AppID] [Description]
```

### Examples

```
$ ttnctl applications create test "Test application"
  INFO Created Application
  INFO Selected Current Application

```

### Options

```
      --app-eui stringSlice   LoRaWAN AppEUI to register with application
      --skip-register         Do not register application with the Handler
      --skip-select           Do not select this application (also adds --skip-register)
```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications delete

Delete an application

### Synopsis


ttnctl devices delete can be used to delete an application.

```
ttnctl applications delete
```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications info

Get information about an application

### Synopsis


ttnctl applications info can be used to info applications.

```
ttnctl applications info [AppID]
```

### Examples

```
$ ttnctl applications info
  INFO Found application

AppID:   test
Name:    Test application
EUIs:
       - 0000000000000000

Access Keys:
       - Name: default key
         Key:  FZYr01cUhdhY1KBiMghUl+/gXyqXhrF6y+1ww7+DzHg=
         Rights: messages:up:r, messages:down:w

Collaborators:
       - Name: yourname
         Rights: settings, delete, collaborators

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications list

List applications

### Synopsis


ttnctl applications list can be used to list applications.

```
ttnctl applications list
```

### Examples

```
$ ttnctl applications list
  INFO Found one application:

 	ID  	Description     	EUIs	Access Keys	Collaborators
1	test	Test application	1   	1          	1

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications pf

Show the payload functions

### Synopsis


ttnctl applications pf shows the payload functions for decoding,
converting and validating binary payload.

```
ttnctl applications pf
```

### Examples

```
$ ttnctl applications pf
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Found Application
  INFO Decoder function
function Decoder(bytes) {
  return {
    payload: bytes,
  };
}
  INFO No converter function
  INFO No validator function
  INFO No encoder function

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications pf set

Set payload functions of an application

### Synopsis


ttnctl pf set can be used to get or set payload functions of an application.
The functions are read from the supplied file or from STDIN.

```
ttnctl applications pf set [decoder/converter/validator/encoder] [file.js]
```

### Examples

```
$ ttnctl applications pf set decoder
  INFO Discovering Handler...
  INFO Connecting with Handler...
function Decoder(bytes) {
  // Here you can decode the payload into json.
  // bytes is of type Buffer.
  // todo: return an object
  return {
    payload: bytes,
  };
}
########## Write your Decoder here and end with Ctrl+D (EOF):
function Decoder(bytes) {
  return {
    payload: bytes,
  };
}
  INFO Updated application                      AppID=test

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications register

Register this application with the handler

### Synopsis


ttnctl register can be used to register this application with the handler.

```
ttnctl applications register
```

### Examples

```
$ ttnctl applications register
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Registered application                   AppID=test

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications select

select the application to use

### Synopsis


ttnctl applications select can be used to select the application to use in next commands.

```
ttnctl applications select
```

### Examples

```
$ ttnctl applications select
  INFO Found one application "test", selecting that one.
  INFO Found one EUI "0000000000000000", selecting that one.
  INFO Updated configuration

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications unregister

Unregister this application from the handler

### Synopsis


ttnctl unregister can be used to unregister this application from the handler.

```
ttnctl applications unregister
```

### Examples

```
$ ttnctl applications unregister
Are you sure you want to unregister application test?
> yes
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Unregistered application                 AppID=test

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl devices

Manage devices

### Synopsis


ttnctl devices can be used to manage devices.

### Options

```
      --app-eui string   The app EUI to use
      --app-id string    The app ID to use
```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl devices create

Create a new device

### Synopsis


ttnctl devices create can be used to create a new device.

```
ttnctl devices create [Device ID] [DevEUI] [AppKey]
```

### Examples

```
$ ttnctl devices create test
  INFO Using Application                        AppEUI=70B3D57EF0000024 AppID=test
  INFO Generating random DevEUI...
  INFO Generating random AppKey...
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Created device                           AppEUI=70B3D57EF0000024 AppID=test AppKey=EBD2E2810A4307263FE5EF78E2EF589D DevEUI=0001D544B2936FCE DevID=test

```

### Options inherited from parent commands

```
      --app-eui string              The app EUI to use
      --app-id string               The app ID to use
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl devices delete

Delete a device

### Synopsis


ttnctl devices delete can be used to delete a device.

```
ttnctl devices delete [Device ID]
```

### Examples

```
$ ttnctl devices delete test
  INFO Using Application                        AppID=test
Are you sure you want to delete device test from application test?
> yes
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Deleted device                           AppID=test DevID=test

```

### Options inherited from parent commands

```
      --app-eui string              The app EUI to use
      --app-id string               The app ID to use
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl devices info

Get information about a device

### Synopsis


ttnctl devices info can be used to get information about a device.

```
ttnctl devices info [Device ID]
```

### Examples

```
$ ttnctl devices info test
  INFO Using Application                        AppEUI=70B3D57EF0000024 AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Found device

  Application ID: test
       Device ID: test
       Last Seen: never

    LoRaWAN Info:

     AppEUI: 70B3D57EF0000024
     DevEUI: 0001D544B2936FCE
    DevAddr: 26001ADA
     AppKey: <nil>
    AppSKey: D8DD37B4B709BA76C6FEC62CAD0CCE51
    NwkSKey: 3382A3066850293421ED8D392B9BF4DF
     FCntUp: 0
   FCntDown: 0
    Options:

```

### Options

```
      --format string   Formatting: hex/msb/lsb (default "hex")
```

### Options inherited from parent commands

```
      --app-eui string              The app EUI to use
      --app-id string               The app ID to use
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl devices list

List al devices for the current application

### Synopsis


ttnctl devices list can be used to list all devices for the current application.

```
ttnctl devices list
```

### Examples

```
$ ttnctl devices list
  INFO Using Application                        AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...

DevID	AppEUI          	DevEUI          	DevAddr 	Up/Down
test 	70B3D57EF0000024	0001D544B2936FCE	26001ADA	0/0

  INFO Listed 1 devices                         AppID=test

```

### Options inherited from parent commands

```
      --app-eui string              The app EUI to use
      --app-id string               The app ID to use
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl devices personalize

Personalize a device

### Synopsis


ttnctl devices personalize can be used to personalize a device (ABP).

```
ttnctl devices personalize [Device ID] [NwkSKey] [AppSKey]
```

### Examples

```
$ ttnctl devices personalize test
  INFO Using Application                        AppEUI=70B3D57EF0000024 AppID=test
  INFO Generating random NwkSKey...
  INFO Generating random AppSKey...
  INFO Discovering Handler...                   Handler=ttn-handler-eu
  INFO Connecting with Handler...               Handler=eu.thethings.network:1904
  INFO Requesting DevAddr for device...
  INFO Personalized device                      AppID=test AppSKey=D8DD37B4B709BA76C6FEC62CAD0CCE51 DevAddr=26001ADA DevID=test NwkSKey=3382A3066850293421ED8D392B9BF4DF

```

### Options inherited from parent commands

```
      --app-eui string              The app EUI to use
      --app-id string               The app ID to use
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl devices set

Set properties of a device

### Synopsis


ttnctl devices set can be used to set properties of a device.

```
ttnctl devices set [Device ID]
```

### Examples

```
$ ttnctl devices set test --fcnt-up 0 --fcnt-down 0
  INFO Using Application                        AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Updated device                           AppID=test DevID=test

```

### Options

```
      --32-bit-fcnt          Use 32 bit FCnt
      --app-eui string       Set AppEUI
      --app-key string       Set AppKey
      --app-s-key string     Set AppSKey
      --dev-addr string      Set DevAddr
      --dev-eui string       Set DevEUI
      --disable-fcnt-check   Disable FCnt check
      --fcnt-down int        Set FCnt Down (default -1)
      --fcnt-up int          Set FCnt Up (default -1)
      --nwk-s-key string     Set NwkSKey
```

### Options inherited from parent commands

```
      --app-id string               The app ID to use
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl downlink

Send a downlink message to a device

### Synopsis


ttnctl downlink can be used to send a downlink message to a device.

```
ttnctl downlink [DevID] [Payload]
```

### Examples

```
$ ttnctl downlink test aabc
  INFO Connecting to MQTT...
  INFO Connected to MQTT
  INFO Enqueued downlink                        AppID=test DevID=test

$ ttnctl downlink test --json '{"led":"on"}'
  INFO Connecting to MQTT...
  INFO Connected to MQTT
  INFO Enqueued downlink                        AppID=test DevID=test

```

### Options

```
      --fport int   FPort for downlink (default 1)
      --json        Provide the payload as JSON
```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl gateway

Manage gateways

### Synopsis


ttnctl applications can be used to manage gateways.

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl gateway status

Get status of a gateway

### Synopsis


ttnctl gateway status can be used to get status of gateways.

```
ttnctl gateway status [gatewayID]
```

### Examples

```
$ ttnctl gateway status eui-0000024b08060030
  INFO Discovering Router...
  INFO Connecting with Router...
  INFO Connected to Router
  INFO Received status

           Last seen: 2016-09-20 08:25:27.94138808 +0200 CEST
           Timestamp: 0
       Reported time: 2016-09-20 08:25:26 +0200 CEST
     GPS coordinates: (52.372791 4.900300)
                 Rtt: not available
                  Rx: (in: 0; ok: 0)
                  Tx: (in: 0; ok: 0)

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl subscribe

Subscribe to events for this application

### Synopsis


ttnctl subscribe can be used to subscribe to events for this application.

```
ttnctl subscribe
```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl uplink

Simulate an uplink message to the network

### Synopsis


ttnctl uplink simulates an uplink message to the network

```
ttnctl uplink [DevAddr] [NwkSKey] [AppSKey] [FCnt] [Payload]
```

### Options

```
      --confirmed   Use confirmed uplink (this also sets --downlink)
      --downlink    Also start downlink (unstable)
```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl user

Show the current user

### Synopsis


ttnctl user shows the current logged on user's profile

```
ttnctl user
```

### Examples

```
$ ttnctl user
  INFO Found user profile:

            Username: yourname
                Name: Your Name
               Email: your@email.org

  INFO Login credentials valid until Sep 20 09:04:12

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl user login

Login

### Synopsis


ttnctl user login allows you to login to the account server.

```
ttnctl user login [client code]
```

### Examples

```
First get an access code from your TTN Profile by going to
https://account.thethingsnetwork.org and clicking "ttnctl access code".

$ ttnctl user login 2keK3FTu6e0327cq4ni0wRTMT2mTS-m_FLzFBlNQadwa
  INFO Successfully logged in as yourname (your@email.org)

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl user logout

Logout the current user

### Synopsis


ttnctl user logout logs out the current user

```
ttnctl user logout
```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl user register

Register

### Synopsis


ttnctl user register allows you to register a new user in the account server

```
ttnctl user register [username] [e-mail]
```

### Examples

```
$ ttnctl user register yourname your@email.org
Password: <entering password>
  INFO Registered user
  WARN You might have to verify your email before you can login

```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl version

Get build and version information

### Synopsis


ttnctl version gets the build and version information of ttnctl

```
ttnctl version
```

### Options inherited from parent commands

```
      --config string               config file (default is $HOME/.ttnctl.yaml)
      --discovery-server string     The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --mqtt-broker string          The address of the MQTT broker (default "eu.thethings.network:1883")
      --ttn-account-server string   The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --ttn-handler string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --ttn-router string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


