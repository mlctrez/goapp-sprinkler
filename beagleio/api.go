package beagleio

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HttpApi struct {
	Api *Api
}

type Pin struct {
	Number int
	Path   string
	State  string
}

func (h *HttpApi) Routes(engine *gin.Engine) {
	engine.GET("/api/pins", func(context *gin.Context) {
		context.JSON(http.StatusOK, h.currentState())
	})
	engine.POST("/api/pins/:pin/:state", func(context *gin.Context) {
		_ = h.Api.InitializePins()
		h.Api.ChangePin(context.Param("pin"), context.Param("state"))
		context.JSON(http.StatusOK, h.currentState())
	})

}

func (h *HttpApi) currentState() map[string][]Pin {
	var pins []Pin
	for i, path := range h.Api.GpioPaths {
		pin := Pin{Number: i, Path: path}
		pin.State = h.Api.ReadPin(i)
		pins = append(pins, pin)
	}
	res := map[string][]Pin{"pins": pins}
	return res
}
