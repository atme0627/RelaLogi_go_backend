package rest

import (
	"github.com/atme0627/RelaLogi_go_backend/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewEngine(h Handlers, cfg config.Config) *gin.Engine {
	e := gin.Default()
	if len(cfg.CorsAllowOrigins) > 0 {
		e.Use(corsMiddleware(cfg.CorsAllowOrigins))
	}
	RegisterRoutes(e, h)
	return e
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = allowedOrigins
	return cors.New(corsConfig)
}
