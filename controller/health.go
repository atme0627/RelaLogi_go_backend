package controller

import (
	"context"
)

type HealthController struct {
}

type HealthResponse struct {
	Message string `json:"message"`
}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (c *HealthController) Get(ctx context.Context) (HealthResponse, error) {
	return HealthResponse{Message: "ok"}, nil
}
