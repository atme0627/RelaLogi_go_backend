package rest

import (
	"github.com/gin-gonic/gin"
)

func NewEngine(h Handlers) *gin.Engine {
	e := gin.Default()
	RegisterRoutes(e, h)
	return e
}
