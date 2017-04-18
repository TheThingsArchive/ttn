package router

import (
	"fmt"
	"errors"
	"github.com/TheThingsNetwork/ttn/api/models"
	"github.com/TheThingsNetwork/ttn/api/auth"
	"github.com/gin-gonic/gin"
)

func getDevice(c *gin.Context) {

	token,  _ := c.Get("token")
	appEUI, _ := c.Get("appEUI")
	devEUI, _ := c.Get("devEUI")

	dev, err := models.DeviceInfo(token.(auth.Token), appEUI.(models.EUI), devEUI.(models.EUI))
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	c.JSON(200, dev)
}

type registerDeviceFields struct {
	AppEUI  models.EUI  `json:"application"`
	DevEUI  models.EUI  `json:"device"`
	AppKey  models.SKey `json:"app_secret" binding:"required"`
	NwkKey  models.SKey `json:"nwk_secret"`
	Type    string      `json:"type" binding:"required"`
	Address []byte      `json:"address"`
}

func registerDevice(c *gin.Context) {

	token,  _ := c.Get("token")
	appEUI, _ := c.Get("appEUI")

	var body registerDeviceFields
	if c.BindJSON(&body) != nil {
		c.AbortWithStatus(400)
		return
	}

	var err error
	switch body.Type {
		case "dynamic":
			if body.AppKey != nil {
				err = models.RegisterOTAA(token.(auth.Token), appEUI.(models.EUI), body.DevEUI, body.AppKey)
			} else {
				err = models.RegisterOTAA(token.(auth.Token), appEUI.(models.EUI), body.DevEUI)
			}
		case "personal":
			switch {
				case body.AppKey == nil && body.NwkKey == nil:
					err = models.RegisterABP(token.(auth.Token), appEUI.(models.EUI), body.Address)
				case body.AppKey != nil && body.NwkKey != nil:
					err = models.RegisterABP(token.(auth.Token), appEUI.(models.EUI), body.Address, body.AppKey, body.NwkKey)
				default:
					c.AbortWithError(400, errors.New("invalid device keys, either give both app_secret and nwk_secret, or give none."))
					return
			}
		default:
			c.AbortWithError(400, errors.New(fmt.Sprintf("invalid device type %s", body.Type)))
			return
	}

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, gin.H{})
}
