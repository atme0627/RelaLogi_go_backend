package framework

import (
	"net/http"

	"github.com/atme0627/RelaLogi_go_backend/controller"
)

type Server struct {
	c *controller.Controller
}

func NewServer(c *controller.Controller) *Server {
	return &Server{c: c}
}
func InitRoute() *http.ServeMux {
	server := NewServer(controller.New())
	mux := http.NewServeMux()
	mux.Handle("GET /health", http.HandlerFunc(server.c.Health))
	return mux
}
