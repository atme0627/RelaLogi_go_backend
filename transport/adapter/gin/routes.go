package gin

import (
	"github.com/atme0627/RelaLogi_go_backend/transport/adapter/gin/handler"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Health *handler.HealthHandler
}

func RegisterRoutes(r *gin.Engine, h Handlers) {
	api := r.Group("/api")
	h.Health.Register(api)
}
