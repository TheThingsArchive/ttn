package router

import (
	"github.com/TheThingsNetwork/ttn/api/models"
	"github.com/TheThingsNetwork/ttn/api/auth"
	"github.com/gin-gonic/gin"
	"errors"
	"net/http"
)

func listApplications (c *gin.Context) {
	token, authorized := c.Get("token")

	// todo: is this even necessary?
	if !authorized {
		c.AbortWithStatus(401)
		return
	}

	apps, err := models.ListApplications(token.(auth.Token))
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, apps)
}


// get application by eui
func getApplication (c *gin.Context) {
	token, authorized := c.Get("token");
	if !authorized {
		c.AbortWithStatus(401)
		return
	}

	// EUI as string
	eui, _ := c.Get("appEUI")

	app, err  := models.GetApplication(token.(auth.Token), eui.(models.EUI))
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	c.JSON(200, app)
}

func getDevicesForApp (c *gin.Context) {
	token, authorized := c.Get("token");
	if !authorized {
		c.AbortWithStatus(401)
		return
	}

	// EUI as string
	eui, _ := c.Get("appEUI")

	// get the app
	app, err := models.GetApplication(token.(auth.Token), eui.(models.EUI))
	if err != nil {
		c.AbortWithError(404, err)
		return
	}

	devices, err := models.DevicesForApp(token.(auth.Token), app.EUI)
	if err != nil {
		devices = make([]models.Device, 0)
		// c.JSON(200, de)
		// c.AbortWithError(500, err)
		// return
	}

	c.JSON(200, devices)
}

type createAppFields struct {
	Name string     `json:"name" binding:"required"`
	Eui  models.EUI `json:"eui"`
}

// creates a new application
func createApplication (c *gin.Context) {
	token, authorized := c.Get("token")
	if !authorized {
		c.AbortWithStatus(401)
		return
	}

	var body createAppFields
	if c.BindJSON(&body) != nil {
		c.AbortWithStatus(400)
		return
	}

	app, err := models.CreateApplication(token.(auth.Token), body.Eui, body.Name)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	claims, found := c.Get("claims")
	if !found {
		c.AbortWithError(500, errors.New("Unexpected: cannot read claims"))
		return
	}

	cclaims := claims.(map[string]interface{})
	email   := cclaims["email"].(string)
	app.Owner = email
	app.Valid = true

	c.JSON(http.StatusCreated, app)
}

type authorizeAppFields struct {
	Email string `json:"name" binding:"required"`
}

func authorizeApplication(c *gin.Context) {

	// get token
	token, authorized := c.Get("token")
	if !authorized {
		c.AbortWithStatus(401)
		return
	}

	// read eui param
	eui, _ := c.Get("appEUI")

	// read body
	var body authorizeAppFields
	if c.BindJSON(&body) != nil {
		c.AbortWithStatus(400)
		return
	}

	// perfrom authorization
	err := models.AuthorizeUserToApplication(token.(auth.Token), eui.(models.EUI), body.Email)

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, gin.H{})
}


func deleteApplication(c *gin.Context) {
	// get token
	token, authorized := c.Get("token")
	if !authorized {
		c.AbortWithStatus(401)
		return
	}

	eui, _ := c.Get("appEUI")
	err := models.DeleteApplication(token.(auth.Token), eui.(models.EUI))
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, gin.H{})
}

