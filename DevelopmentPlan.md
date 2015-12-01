Development Plan
================

## Milestone 1
Have a fake gateway able to mock the behavior of an existing real gateway. This will be used
mainly for testing and ensuring the correctness of other components.

- [ ] Fake gateway
    - [x] Types, packages and data structures in use
    - [x] Emit udp packets towards a server
    - [ ] Handle reception acknowledgement from server
    - [ ] Generate and serialize json rxpk object(s) 
    - [ ] Generate and serialize json stat object(s)
    - [ ] Simulate fake end-devices activity 

```go
type gateway struct {
    id       []string
    routers  []string
    faking   boolean
    started  boolean
    status   Stat
}

type RXPK struct {
	Chan uint      `json:"chan"` // Concentrator "IF" channel used for RX (unsigned integer)
	Codr string    `json:"codr"` // LoRa ECC coding rate identifier
    Data string    `json:"data"` // Base64 encoded RF packet payload, padded
	Datr string    `json:"datr"` // FSK datarate (unsigned in bit per second) || LoRa datarate identifier
	Freq float64   `json:"freq"` // RX Central frequency in MHx (unsigned float, Hz precision)
	Lsnr float64   `json:"lsnr"` // LoRa SNR ratio in dB (signed float, 0.1 dB precision)
	Modu string    `json:"modu"` // Modulation identifier "LORA" or "FSK"
	Rfch uint      `json:"rfch"` // Concentrator "RF chain" used for RX (unsigned integer)
	Rssi int       `json:"rssi"` // RSSI in dBm (signed integer, 1 dB precision)
	Size uint      `json:"size"` // RF packet payload size in bytes (unsigned integer)
	Stat int       `json:"stat"` // CRC status: 1 - OK, -1 = fail, 0 = no CRC
    Time time.Time `json:"time"` // UTC time of pkt RX, us precision, ISO 8601 'compact' format
	Tmst uint      `json:"tmst"` // Internal timestamp of "RX finished" event (32b unsigned)
}

type TXPK struct {
    Codr string     `json:"codr"` // LoRa ECC coding rate identifier
    Data string     `json:"data"` // Base64 encoded RF packet payload, padding optional
    Datr string     `json:"datr"` // LoRa datarate identifier (eg. SF12BW500) || FSK Datarate (unsigned, in bits per second)
    Fdev uint       `json:"fdev"` // FSK frequency deviation (unsigned integer, in Hz) 
    Freq float64    `json:"freq"` // TX central frequency in MHz (unsigned float, Hz precision)
    Imme bool       `json:"imme"` // Send packet immediately (will ignore tmst & time)
    Ipol bool       `json:"ipol"` // Lora modulation polarization inversion
    Modu string     `json:"modu"` // Modulation identifier "LORA" or "FSK"
    Ncrc bool       `json:"ncrc"` // If true, disable the CRC of the physical layer (optional)
    Powe uint       `json:"powe"` // TX output power in dBm (unsigned integer, dBm precision)
    Prea uint       `json:"prea"` // RF preamble size (unsigned integer)
    Rfch uint       `json:"rfch"` // Concentrator "RF chain" used for TX (unsigned integer)
    Size uint       `json:"size"` // RF packet payload size in bytes (unsigned integer)
    Time time.Time  `json:"time"` // Send packet at a certain time (GPS synchronization required)
    Tmst uint       `json:"tmst"` // Send packet on a certain timestamp value (will ignore time)
}

type Stat struct {
	Ackr float64    `json:"ackr"` // Percentage of upstream datagrams that were acknowledged
	Alti int        `json:"alti"` // GPS altitude of the gateway in meter RX (integer)
	Dwnb uint       `json:"dwnb"` // Number of downlink datagrams received (unsigned integer)
	Lati float64    `json:"lati"` // GPS latitude of the gateway in degree (float, N is +)
	Long float64    `json:"long"` // GPS latitude of the gateway in dgree (float, E is +)
	Rxfw uint       `json:"rxfw"` // Number of radio packets forwarded (unsigned integer)
	Rxnb uint       `json:"rxnb"` // Number of radio packets received (unsigned integer)
	Rxok uint       `json:"rxok"` // Number of radio packets received with a valid PHY CRC
    Time time.Time  `json:"time"` // UTC 'system' time of the gateway, ISO 8601 'expanded' format
	Txnb uint       `json:"txnb"` // Number of packets emitted (unsigned integer)
}

type Option struct {
    key string
    value {}interface
}

type Packet struct {
    version     byte
    token       [2]byte
    identifier  byte
    payload     []byte
}

type Gateway interface {
    Start () error
    EmitData (data []byte, options ...Option) error
    EmitStat (options ...Option) error
    Mimic (errors <-chan error)

    generateRXPK(nb int, options ...Option) []RXPK
    generateStat(options ...Option) Stat
    createPushData(rxpk []RXPK, stat Stat) error, []byte
    pull(routers ...string)
    decodeResponse(response []byte) error, Packet
}

Create (id string, routers ...string) error, *Gateway
```


## Milestone 2
Handle an uplink process that can forward packet coming from a gateway to a simple end-server
(fake handler). We handle no mac command and we does not care about registration yet. The
system will just forward messages using pre-configured end-device addresses.


- [ ] Basic Router  
    - [ ] Detail the list of features


- [ ] Basic Broker
    - [ ] Detail the list of features


- [ ] Minimalist Dumb Network-Server
    - [ ] Detail the list of features

## Milestone 3
Handle OTAA and downlink accept message. We still not allow mac commands from neither the
end-device nor a network server. Also, no messages can be sent by an application or whatever.
The only downlink message we accept is the join-accept / join-reject message sent during an
OTAA.

- [ ] Extend Router
    - [ ] Detail the list of features


- [ ] Extend Broker
    - [ ] Detail the list of features


- [ ] Extend Network-server
    - [ ] Detail the list of features


- [ ] Minimalist Handler
    - [ ] Detail the list of features

## Milestone 4
Allow transmission of downlink messages from an application. Messages will be shipped as
response after an uplink transmission from a node.

- [ ] Extend Broker
    - [ ] Detail the list of features


- [ ] Extend Handler
    - [ ] Detail the list of features


- [ ] Fake minismalist Application server
    - [ ] Detail the list of features

## Milestone 5
Handle more complexe commands and their corresponding acknowledgement. 

- [ ] Extend Network server
    - [ ] Detail the list of features
