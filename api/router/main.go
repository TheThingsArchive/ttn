package router

import (
	"os"
	"log"
	"errors"
	"github.com/gin-gonic/gin"
  "github.com/TheThingsNetwork/ttn/api/auth"
	"github.com/TheThingsNetwork/ttn/api/models"
  "github.com/spf13/viper"
)

var (
	out = log.New(os.Stdout, "", 0)
	err = log.New(os.Stderr, "error: ", 0)
)

// handle errors by showing the reason in response
func errorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) > 0 {
		c.JSON(-1, c.Errors)
	}
}

// read app eui param and store it, aborting when the EUI is invalid
func appEUI(c *gin.Context) {
	reui, err := models.ReadEUI(c.Param("appeui"))
	if err != nil {
		c.AbortWithError(400, errors.New("invalid application eui"))
	} else {
		c.Set("appEUI", models.EUI(reui))
		c.Next()
	}
}

// read dev eui param and store it, aborting when the EUI is invalid
func devEUI(c *gin.Context) {
	reui, err := models.ReadEUI(c.Param("deveui"))
	if err != nil {
		c.AbortWithError(400, errors.New("invalid device eui"))
	} else {
		c.Set("devEUI", models.EUI(reui))
		c.Next()
	}
}


// start the gin router
func Start() {
	authServer := viper.GetString("ttn-account-server")

	rtr := gin.Default()
	rtr.Use(errorHandler)

	// redirect trailing slashes: foo/ -> foo
	rtr.RedirectTrailingSlash = true

	api := rtr.Group("/api")
	v1  := api.Group("/v1")
	v1.Use(auth.RequireAuth(authServer))


	me := v1.Group("/me")
	me.GET("/", GetUser)

	apps := v1.Group("/applications")
	apps.GET("/", listApplications)
	apps.GET("/:appeui", appEUI, getApplication)
	apps.POST("/", createApplication)
	apps.POST("/:appeui/authorize", appEUI, authorizeApplication)
	apps.DELETE("/:appeui", appEUI, deleteApplication)

	devs := apps.Group("/:appeui/devices", appEUI)
	devs.GET("/", appEUI, getDevicesForApp)
	devs.GET("/:deveui", devEUI, getDevice)
	devs.POST("/", registerDevice)

	// TODO:
	// devs.POST("/devices/:eui/downlink", sendDownLink)
	// devs.GET("/:deveui/subscribe", subscribeApplication)
	// devs.GET("/:deveui/subscribe", subscribeDevice)

	rtr.Run()
}
