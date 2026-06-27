package handler

import (
	"net/http"

	"github.com/atme0627/RelaLogi_go_backend/controller"
	"github.com/atme0627/RelaLogi_go_backend/transport/rest/response"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	c *controller.HealthController
}

func NewHandler(c *controller.HealthController) *HealthHandler {
	return &HealthHandler{c: c}
}

func (h *HealthHandler) Register(r gin.IRoutes) {
	r.GET("/health", h.get)
}

func (h *HealthHandler) get(ctx *gin.Context) {
	resp, err := h.c.Get(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: "server error"})
	}
	ctx.JSON(http.StatusOK, resp)
}
