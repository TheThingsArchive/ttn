Development Plan
================

## Milestone 1
Have a fake gateway able to mock the behavior of an existing real gateway. This will be used
mainly for testing and ensuring the correctness of other components.

- [ ] Fake gateway
    - [ ] Types, packages and data structures in use
    - [ ] Emit udp packets towards a server
    - [ ] Handle reception acknowledgement from server
    - [ ] Generate and serialize json rxpk object(s) 
    - [ ] Generate and serialize json stat object(s)
    - [ ] Simulate fake end-devices activity 


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
