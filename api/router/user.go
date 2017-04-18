package router

import (
	"github.com/TheThingsNetwork/ttn/api/models"
	"github.com/TheThingsNetwork/ttn/api/auth"
	"github.com/gin-gonic/gin"
	"errors"
)

// get info about current authorized user
func GetUser (c *gin.Context) {
	token, authorized := c.Get("token");
	if !authorized {
		c.AbortWithStatus(401)
		return
	}

	apps, err := models.ListApplications(token.(auth.Token))
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	claims, found := c.Get("claims")
	if !found {
		c.AbortWithError(500, errors.New("cannot find claims"))
		return
	}

	cclaims := claims.(map[string]interface{})
	email   := cclaims["email"].(string)

	c.JSON(200, models.User{
		Email: email,
		Apps:  apps,
	})

}

