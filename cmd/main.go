package main

import (
	"log"

	"github.com/atme0627/RelaLogi_go_backend/controller"
	"github.com/atme0627/RelaLogi_go_backend/transport/http"
	"github.com/atme0627/RelaLogi_go_backend/transport/http/handler"
)

func main() {
	healthController := controller.NewHealthController()
	healthHandler := handler.NewHandler(healthController)
	handlers := http.Handlers{
		Health: healthHandler,
	}
	e := http.NewEngine(handlers)
	log.Fatal(e.Run(":8080"))
}
