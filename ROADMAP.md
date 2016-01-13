Roadmap
=======

## Milestone 1 - [Issues](https://github.com/TheThingsNetwork/ttn/milestones/Milestone%201)

Have a gateway simulator that is able to mock the behavior of a physical gateway. This will be used for testing and ensuring the correctness of other components.

## Milestone 2 - [Issues](https://github.com/TheThingsNetwork/ttn/milestones/Milestone%202)

Support for uplink messages (from a node to an application). A gateway can send received messages to a [Router](https://thethingsnetwork.github.io/docs/router/). The Router filters out messages that are part of other networks and routes "our" messages to to a [Broker](https://thethingsnetwork.github.io/docs/broker/). The Broker forwards the messages to a [Handler](https://thethingsnetwork.github.io/docs/handler/), which delivers them to the Application.

We will not support any MAC commands, nor device or application registration. The system will just forward messages using pre-configured server and end-device addresses.

## Milestone 3 - [Issues](https://github.com/TheThingsNetwork/ttn/milestones/Milestone%203)

Support application registration for personalization. Applications provide a list of personalized device addresses along with the network session keys.

## Milestone 4 - [Issues](https://github.com/TheThingsNetwork/ttn/milestones/Milestone%204)

Support for downlink messages (from an application to a node). Messages will be shipped as a response to an uplink transmission from a (Class A) node.

## Milestone 5 - [Issues](https://github.com/TheThingsNetwork/ttn/milestones/Milestone%205)

Support for Over-the-air Activation (OTAA). Devices with a globally unique end-device identifier (DevEUI), an application identifier (AppEUI) and an application key (AppKey) can send join-requests to the network.

We still not allow MAC commands from neither the end-device nor a network controller.

## Milestone 6 - [Issues](https://github.com/TheThingsNetwork/ttn/milestones/Milestone%206)

Support for LoRaWAN MAC commands.
