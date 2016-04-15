package router

// This is harder than anticipated,
// websockets only allow cookies (so no Authorization header)
// let's look at gorilla/websockets, they seem to have an upgrade
// mechanism in place.

// import (
// 	"github.com/TheThingsNetwork/ttn/api/models"
// 	"github.com/TheThingsNetwork/ttn/api/auth"
// 	"github.com/TheThingsNetwork/ttn/core"
// 	"golang.org/x/net/websocket"
// 	"github.com/gin-gonic/gin"
// )
//
// subscribe to all devices for an application
// func subscribeApplication(c *gin.Context) {
//
// 	token,  _ := c.Get("token")
//
//
// 	handler := websocket.Handler(func(ws *websocket.Conn) {
// 		// todo check errors
// 		done, err := models.Subscribe(token.(auth.Token), appEUI.(models.EUI), func(dataUp core.DataUpAppReq) {
// 			println(dataUp)
// 			ws.Write([]byte("KOEL"))
// 		})
//
// 		defer done()
// 	})
// 	handler.ServeHTTP(c.Writer, c.Req)
// 	// TODO: handle socket close
// }
//
//
// func subscribeDevice(c *gin.Context) {
//
// 	token,  _ := c.Get("token")
// 	appEUI, _ := c.Get("appEUI")
// 	appEUI, _ := c.Get("devEUI")
//
//
// 	handler := websocket.Handler(func(ws *websocket.Conn) {
// 		// todo check errors
// 		done := models.Subscribe(token.(auth.Token), appEUI.(models.EUI), func(dataUp core.DataUpAppReq) {
// 			println(dataUp)
// 			ws.Write([]byte("KOEL"))
// 		})
//
// 		defer done()
// 	})
// 	handler.ServeHTTP(c.Writer, c.Req)
// 	// TODO: handle socket close
// }
