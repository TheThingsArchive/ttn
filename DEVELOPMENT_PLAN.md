Development Plan
================

## Milestone 1
Have a fake gateway able to mock the behavior of a physical gateway. This will be used
mainly for testing and ensuring the correctness of other components.

- [ ] Fake gateway
    - [x] Types, packages and data structures in use
    - [x] Emit udp packets towards a server
    - [ ] Handle behavior described by the semtech protocol
    - [x] Serialize json rxpk/stat object(s) 
    - [x] Generate json rxpl/stat object(s)
    - [ ] Simulate fake end-devices activity 
    - [ ] Update gateway statistics accordingly

## Milestone 2
Handle an uplink process that can forward packet coming from a gateway to a simple end-server
(fake handler). We handle no MAC commands and we does not care about registration yet. The
system will just forward messages using pre-configured end-device addresses.


- [ ] Basic Router  
    - [ ] Core
        - [x] Lookup for device address (only local)
        - [x] Invalidate broker periodically (only local)
        - [x] Acknowledge packet from gateway
        - [x] Forward packet to brokers
        - [ ] Create Mock router
        - [x] Create Mock DownAdapter
        - [x] Create Mock UpAdatper
        - [ ] Test'em all
        - [ ] Reemit errored packet
        - [ ] Switch from local in-memory storage to Reddis
    - [ ] UpAdapter
        - [x] Listen and forward incoming packets to Core router
        - [x] Keep track of existing UDP connections
        - [x] Send ack through existing UDP connection
    - [ ] DownAdapter
        - [x] Listen and forward incoming packet to Core router
        - [x] Broadcast a packet to several brokers
        - [x] Send packet to given brokers (same as above ?)


- [ ] Basic Broker
    - [ ] Detail the list of features


- [ ] Minimalist Dumb Network-Controller
    - [ ] Detail the list of features

## Milestone 3
Support application registration for personalization. Applications provide a list of
personalized device addresses along with the network session keys

- [ ] Extend Router
    - [ ] Detail the list of features


- [ ] Extend Broker
    - [ ] Detail the list of features


- [ ] Extend Network-Controller
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


- [ ] Fake minimalist Application server
    - [ ] Detail the list of features

## Milestone 5
Handle OTAA and downlink accept message. We still not allow MAC commands from neither the
end-device nor a network controller. Also, no downlink payload can be sent by an application:
the only downlink message we accept is the join-accept / join-reject message sent during an

- [ ] Extend Router
    - [ ] Detail the list of features


- [ ] Extend Broker
    - [ ] Detail the list of features


- [ ] Extend Network-Controller
    - [ ] Detail the list of features


- [ ] Minimalist Handler
    - [ ] Detail the list of features



## Milestone 6
Handle more complexe commands and their corresponding acknowledgement. 

- [ ] Extend Network Controller
    - [ ] Detail the list of features
