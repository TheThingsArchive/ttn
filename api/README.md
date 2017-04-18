# ttn-http

This is a http wrapper for the The Things Network.

```
get info about me and my applications
GET   /me

list all applictions
GET   /applications

create new application
POST  /applications

get info about specific application
GET   /applications/:eui

authorize a a user for an application
POST  /applications/:eui/authorize

delete an application
DELETE /applications/:eui

get all devices for application
GET   /applications/:eui/devices

get info about a specific device
GET   /applications/:eui/devices/:deui

register a device
POST  /applications/:eui/devices

listen to all devices for application over websocket (todo)
GET /applications/:eui/devices/subscribe

listen a specific device over websocket (todo)
GET /applications/:eui/devices/:deui/subscribe

send message to a device (todo)
POST /applications/:eui/devices/:deui/message
```

All of these enpoints require an access token that you can get
from The Things Network account server ([https://account.thethingsnetwork.org][]).


More in depth documentation is forthcoming!
