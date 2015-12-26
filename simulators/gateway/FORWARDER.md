Forwarder
-------

### Behavior

From the [packet
forwarder](https://github.com/TheThingsNetwork/packet_forwarder/blob/master/PROTOCOL.TXT)
defined by Semtech we can extract the following behavior for the forwarder:

- The forwarder is able to send two types of packet:
    - `PUSH_DATA` packets with a payload coming from a device
    - `PULL_DATA` packets use to maintain an established connection

- The forwarder is able to receive three types of packet:
    - `PUSH_ACK` packets which acknowledge a `PUSH_DATA`
    - `PULL_ACK` packets which acknowledge a `PULL_DATA`
    - `PULL_RESP` packets used to transfer a response down to the devices

For the moment, we simply log down any missing acknowledgement. In a second time, we'll
consider re-emitting the corresponding packets.

A forwarder instance does not presume of any activity, it is sleeping by default and is rather
stimulated by an external agent. 

We want the forwarder to log and store each downlink packet. However, any external agent can
still flush the forwarder to retrieve all stored packet and clear the forwarder internal buffer.
This way, the forwarder is nothing more than a forwarder while the handling logic is under the
control of a separated entity. 

When a downlink datagram is received it is stored unless it does not reflect a valid semtech
Packet (i.e., a `PUSH_ACK`, `PULL_ACK` or `PULL_RESP` with valid data). Any other data received
by the forwarder is ignored.

### Interfaces

```go
// New create a forwarder instance bound to a set of routers. 
func NewForwarder (id string, routers ...io.ReadWriteCloser) (*Forwarder, error)  

// Flush spits out all downlink packet received by the forwarder since the last flush.
func (fwd *Forwarder) Flush() []semtech.Packet

// Forward dispatch a packet to all connected routers. 
func (fwd *Forwarder) Forward(packet semtech.Packet) error

// Stats computes and return the forwarder statistics since it was created
func (fwd Forwarder) Stats() semtech.Stat

// Stop terminate the forwarder activity. Closing all routers connections
func (fwd *Forwarder) Stop() error
```

### Stats

 Name   |   Type   | Function
--------|----------|--------------------------------------------------------------
 *time* | *string* | *UTC 'system' time of the gateway, ISO 8601 'expanded' format*
 *lati* | *number* | *GPS latitude of the gateway in degree (float, N is +)*
 *long* | *number* | *GPS latitude of the gateway in degree (float, E is +)*
 *alti* | *number* | *GPS altitude of the gateway in meter RX (integer)*
 rxnb | number | Number of radio packets received (unsigned integer)
 rxok | number | Number of radio packets received with a valid PHY CRC
 rxfw | number | Number of radio packets forwarded (unsigned integer)
 ackr | number | Percentage of upstream datagrams that were acknowledged
 dwnb | number | Number of downlink datagrams received (unsigned integer)
 txnb | number | Number of packets emitted (unsigned integer)

##### rxnb
Incremented each time a packet is received from a device. In other words, any call to Forward
with a non-nil packet should incremented that stat.

##### rxok 
Incremented each time a packet is received from a device. Because the forwarder only simulate
what a real gateway would do, we do not consider a full device packet with a CRC and a PHY
payload. We only consider the payload and thus, rxok and rxnb should be seemingly the same.

##### rxfw 
This conveys a successful packet forwarding. It should be incremented once per packet received
from devices that has successfully been forwarded to routers (regardless of any ack from them).

##### ackr
Computed using the number of forwarded packet that has been acknowledged and the total number
of forwarded packet.

##### dwnb
Incremented each time a packet is received from a router.

##### txnb 
Incremented for each packet forwarded but also for each `PULL_DATA` packets sent. 
