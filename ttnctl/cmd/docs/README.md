# API Reference

Control The Things Network from the command line.

**Options**

```
      --allow-insecure             Allow insecure fallback if TLS unavailable
      --auth-server string         The address of the OAuth 2.0 server (default "https://account.thethingsnetwork.org")
      --config string              config file (default is $HOME/.ttnctl.yml)
      --data string                directory where ttnctl stores data (default is $HOME/.ttnctl)
      --discovery-address string   The address of the Discovery server (default "discover.thethingsnetwork.org:1900")
      --handler-id string          The ID of the TTN Handler as announced in the Discovery server (default "ttn-handler-eu")
      --mqtt-address string        The address of the MQTT broker (default "eu.thethings.network:1883")
      --mqtt-password string       The password for the MQTT broker
      --mqtt-username string       The username for the MQTT broker
      --router-id string           The ID of the TTN Router as announced in the Discovery server (default "ttn-router-eu")
```


## ttnctl applications

ttnctl applications can be used to manage applications.

### ttnctl applications add

ttnctl applications add can be used to add a new application to your account.

**Usage:** `ttnctl applications add [AppID] [Description]`

**Options**

```
      --app-eui stringSlice   LoRaWAN AppEUI to register with application
      --skip-register         Do not register application with the Handler
      --skip-select           Do not select this application (also adds --skip-register)
```

**Example**

```
$ ttnctl applications add test "Test application"
  INFO Added Application
  INFO Selected Current Application
```

### ttnctl applications delete

ttnctl devices delete can be used to delete an application.

**Usage:** `ttnctl applications delete [AppID]`

### ttnctl applications info

ttnctl applications info can be used to info applications.

**Usage:** `ttnctl applications info [AppID]`

**Example**

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

### ttnctl applications list

ttnctl applications list can be used to list applications.

**Usage:** `ttnctl applications list`

**Example**

```
$ ttnctl applications list
  INFO Found one application:

 	ID  	Description     	EUIs	Access Keys	Collaborators
1	test	Test application	1   	1          	1
```

### ttnctl applications pf

ttnctl applications pf shows the payload functions for decoding,
converting and validating binary payload.

**Usage:** `ttnctl applications pf`

**Example**

```
$ ttnctl applications pf
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Found Application
  INFO Decoder function
function Decoder(bytes, port) {
  var decoded = {};
  if (port === 1) {
    decoded.led = bytes[0];
  }
  return decoded;
}
  INFO No converter function
  INFO No validator function
  INFO No encoder function
```

#### ttnctl applications pf set

ttnctl pf set can be used to get or set payload functions of an application.
The functions are read from the supplied file or from STDIN.

**Usage:** `ttnctl applications pf set [decoder/converter/validator/encoder] [file.js]`

**Options**

```
      --skip-test   skip payload function test
```

**Example**

```
$ ttnctl applications pf set decoder
  INFO Discovering Handler...
  INFO Connecting with Handler...
function Decoder(bytes, port) {
  // Decode an uplink message from a buffer
  // (array) of bytes to an object of fields.
  var decoded = {};

  // if (port === 1) {
  //   decoded.led = bytes[0];
  // }

  return decoded;
}
########## Write your Decoder here and end with Ctrl+D (EOF):
function Decoder(bytes, port) {
  var decoded = {};

  // if (port === 1) {
  //   decoded.led = bytes[0];
  // }

  return decoded;
}

Do you want to test the payload functions? (Y/n)
Y

Payload: 12 34
Port: 1

  INFO Function tested successfully

  INFO Updated application                      AppID=test
```

### ttnctl applications register

ttnctl applications register can be used to register this application with the handler.

**Usage:** `ttnctl applications register`

**Example**

```
$ ttnctl applications register
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Registered application                   AppID=test
```

### ttnctl applications select

ttnctl applications select can be used to select the application to use in next commands.

**Usage:** `ttnctl applications select [AppID [AppEUI]]`

**Example**

```
$ ttnctl applications select
  INFO Found one application "test", selecting that one.
  INFO Found one EUI "0000000000000000", selecting that one.
  INFO Updated configuration
```

### ttnctl applications unregister

ttnctl unregister can be used to unregister this application from the handler.

**Usage:** `ttnctl applications unregister`

**Example**

```
$ ttnctl applications unregister
Are you sure you want to unregister application test?
> yes
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Unregistered application                 AppID=test
```

## ttnctl config

ttnctl config prints the configuration that is used

**Usage:** `ttnctl config`

## ttnctl devices

ttnctl devices can be used to manage devices.

**Options**

```
      --app-eui string   The app EUI to use
      --app-id string    The app ID to use
```

### ttnctl devices delete

ttnctl devices delete can be used to delete a device.

**Usage:** `ttnctl devices delete [Device ID]`

**Example**

```
$ ttnctl devices delete test
  INFO Using Application                        AppID=test
Are you sure you want to delete device test from application test?
> yes
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Deleted device                           AppID=test DevID=test
```

### ttnctl devices info

ttnctl devices info can be used to get information about a device.

**Usage:** `ttnctl devices info [Device ID]`

**Options**

```
      --format string   Formatting: hex/msb/lsb (default "hex")
```

**Example**

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

### ttnctl devices list

ttnctl devices list can be used to list all devices for the current application.

**Usage:** `ttnctl devices list`

**Example**

```
$ ttnctl devices list
  INFO Using Application                        AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...

DevID	AppEUI          	DevEUI          	DevAddr
test 	70B3D57EF0000024	0001D544B2936FCE	26001ADA

  INFO Listed 1 devices                         AppID=test
```

### ttnctl devices personalize

ttnctl devices personalize can be used to personalize a device (ABP).

**Usage:** `ttnctl devices personalize [Device ID] [NwkSKey] [AppSKey]`

**Example**

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

### ttnctl devices register

ttnctl devices register can be used to register a new device.

**Usage:** `ttnctl devices register [Device ID] [DevEUI] [AppKey] [Lat,Long]`

**Example**

```
$ ttnctl devices register test
  INFO Using Application                        AppEUI=70B3D57EF0000024 AppID=test
  INFO Generating random DevEUI...
  INFO Generating random AppKey...
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Registered device                        AppEUI=70B3D57EF0000024 AppID=test AppKey=EBD2E2810A4307263FE5EF78E2EF589D DevEUI=0001D544B2936FCE DevID=test
```

#### ttnctl devices register on-join

ttnctl devices register on-join can be used to register a device template for on-join registrations.

**Usage:** `ttnctl devices register on-join [Device ID Prefix] [AppKey]`

### ttnctl devices set

ttnctl devices set can be used to set properties of a device.

**Usage:** `ttnctl devices set [Device ID]`

**Options**

```
      --16-bit-fcnt          Use 16 bit FCnt
      --32-bit-fcnt          Use 32 bit FCnt (default)
      --altitude int32       Set altitude
      --app-eui string       Set AppEUI
      --app-key string       Set AppKey
      --app-s-key string     Set AppSKey
      --description string   Set Description
      --dev-addr string      Set DevAddr
      --dev-eui string       Set DevEUI
      --disable-fcnt-check   Disable FCnt check
      --enable-fcnt-check    Enable FCnt check (default)
      --fcnt-down int        Set FCnt Down (default -1)
      --fcnt-up int          Set FCnt Up (default -1)
      --latitude float32     Set latitude
      --longitude float32    Set longitude
      --nwk-s-key string     Set NwkSKey
      --override             Override protection against breaking changes
```

**Example**

```
$ ttnctl devices set test --fcnt-up 0 --fcnt-down 0
  INFO Using Application                        AppID=test
  INFO Discovering Handler...
  INFO Connecting with Handler...
  INFO Updated device                           AppID=test DevID=test
```

### ttnctl devices simulate

ttnctl devices simulate can be used to simulate an uplink message for a device.

**Usage:** `ttnctl devices simulate [Device ID] [Payload]`

**Options**

```
      --port uint32   Port number (default 1)
```

## ttnctl downlink

ttnctl downlink can be used to send a downlink message to a device.

**Usage:** `ttnctl downlink [DevID] [Payload]`

**Options**

```
      --access-key string   The access key to use
      --confirmed           Confirmed downlink
      --fport int           FPort for downlink (default 1)
      --json                Provide the payload as JSON
```

**Example**

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

## ttnctl gateways

ttnctl gateways can be used to manage gateways.

### ttnctl gateways delete

ttnctl gateways delete can be used to delete a gateway

**Usage:** `ttnctl gateways delete [GatewayID]`

**Example**

```
$ ttnctl gateways delete test
  INFO Deleted gateway                          Gateway ID=test
```

### ttnctl gateways edit

ttnctl gateways edit can be used to edit settings of a gateway

**Usage:** `ttnctl gateways edit [GatewayID]`

**Options**

```
      --frequency-plan string   The frequency plan to use on the gateway
      --location string         The location of the gateway
```

**Example**

```
$ ttnctl gateways edit test --location 52.37403,4.88968 --frequency-plan EU
  INFO Edited gateway                          Gateway ID=test
```

### ttnctl gateways info

ttnctl gateways info can be used to get information about a gateway

**Usage:** `ttnctl gateways info [GatewayID]`

### ttnctl gateways list

ttnctl gateways list can be used to list the gateways you have access to

**Usage:** `ttnctl gateways list`

**Example**

```
$ ttnctl gateways list
 	ID  	Activated	Frequency Plan	Coordinates
1	test	true		US				(52.3740, 4.8896)
```

### ttnctl gateways register

ttnctl gateways register can be used to register a gateway

**Usage:** `ttnctl gateways register [GatewayID] [FrequencyPlan] [Location]`

**Example**

```
$ ttnctl gateways register test US 52.37403,4.88968
  INFO Registered gateway                          Gateway ID=test
```

### ttnctl gateways status

ttnctl gateways status can be used to get status of gateways.

**Usage:** `ttnctl gateways status [gatewayID]`

**Example**

```
$ ttnctl gateways status test
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

## ttnctl selfupdate

ttnctl selfupdate updates the current ttnctl to the latest version

**Usage:** `ttnctl selfupdate`

## ttnctl subscribe

ttnctl subscribe can be used to subscribe to events for this application.

**Usage:** `ttnctl subscribe`

**Options**

```
      --access-key string   The access key to use
```

## ttnctl user

ttnctl user shows the current logged on user's profile

**Usage:** `ttnctl user`

**Example**

```
$ ttnctl user
  INFO Found user profile:

            Username: yourname
                Name: Your Name
               Email: your@email.org

  INFO Login credentials valid until Sep 20 09:04:12
```

### ttnctl user login

ttnctl user login allows you to log in to your TTN account.

**Usage:** `ttnctl user login [access code]`

**Example**

```
First get an access code from your TTN profile by going to
https://account.thethingsnetwork.org and clicking "ttnctl access code".

$ ttnctl user login [paste the access code you requested above]
  INFO Successfully logged in as yourname (your@email.org)
```

### ttnctl user logout

ttnctl user logout logs out the current user

**Usage:** `ttnctl user logout`

### ttnctl user register

ttnctl user register allows you to register a new user in the account server

**Usage:** `ttnctl user register [username] [e-mail]`

**Example**

```
$ ttnctl user register yourname your@email.org
Password: <entering password>
  INFO Registered user
  WARN You might have to verify your email before you can login
```

## ttnctl version

ttnctl version gets the build and version information of ttnctl

**Usage:** `ttnctl version`

