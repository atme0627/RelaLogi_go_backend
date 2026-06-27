package controller

import (
	"context"

	"github.com/atme0627/RelaLogi_go_backend/transport/rest/oapi"
)

type HealthController struct {
}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (c *HealthController) Get(ctx context.Context) oapi.HealthResponse {
	return oapi.HealthResponse{Status: oapi.Ok}
}
