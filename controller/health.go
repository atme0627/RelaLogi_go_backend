package controller

import (
	"context"
)

type HealthController struct {
}

type HealthResponse struct {
	Message string `json:"message"`
}

func New() *HealthController {
	return &HealthController{}
}

func (c *HealthController) get(ctx *context.Context) (HealthResponse, error) {
	return HealthResponse{Message: "ok"}, nil
}
