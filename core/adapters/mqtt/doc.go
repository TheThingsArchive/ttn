// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

/*
Package mqtt provides an MQTT adapter between the network and an MQTT broker. The adapter listens on
2 topics:

    +/devices/+/down
    +/devices/+/activations

and may publish on two of them:

    +/devices/+/up
    +/devices/+/activations

Above, the first wildcard refers to an Application Unique Identifier (AppEUI) which is a
hex-encoded string of 16 characters (representing an 8-bytes long AppEUI). The second wildcard
refers to a Device Unique Identifier (DevEUI) which is also an hex-encoded string of 16
characters with one exception.

For ABP (Activation By Personalization), the activation should be made to:

    +/devices/personalized/activations


Serialization Format

For each topic, a MessagePack-JSON serialization format is expected, with the following top-level
json structure:


ABP :: +/devices/personalized/activations

    {"dev_addr": "% 4 bytes hex-encoded %", "apps_key": "% 16 bytes hex-encoded %", "nwks_key": "% 16 bytes hex-encoded %"}
    {"dev_addr":"01020304","apps_key":"01020304050607080900010203040506","nwks_key":"01020304050607080900010203040506"}

OTAA :: +/devices/+/activations

... TODO

Downlink :: +/devices/+/down

    {"payload": % sequence of bytes, decimal format %}
    {"payload":[112,97,116,97,116,101]}

Uplink :: +/devices/+/up

	{
         "payload": %sequence of bytes, decimal format%,
         "metadata": [{
				"coding_rate":	% string %,
				"data_rate":	% string %,
				"frequency":	% float %,
				"timestamp":	% uint %,
				"rssi":		% int %,
				"lsnr":		% float %,
				"altitude":	% int %,
				"latitude":	% float %,
				"longitude":	% float %
		}, ... ]
	}

	{
         "payload": [112,97,116,97,116,101],
         "metadata": [{
				"coding_rate":	"4/6",
				"data_rate":	"SF8BW125",
				"frequency":	"866.345",
				"timestamp":	"123698454",
				"rssi":		"-37",
				"lsnr":		"5.3",
				"altitude":	"56",
				"latitude":	"-14.678",
				"longitude":	"33.120182"
		}]
	}
*/
package mqtt
