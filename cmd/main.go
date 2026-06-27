package main

import (
	"log"

	"github.com/atme0627/RelaLogi_go_backend/controller"
	"github.com/atme0627/RelaLogi_go_backend/transport/rest"
	"github.com/atme0627/RelaLogi_go_backend/transport/rest/handler"
)

func main() {
	healthController := controller.NewHealthController()
	healthHandler := handler.NewHandler(healthController)
	handlers := rest.Handlers{
		Health: healthHandler,
	}
	e := rest.NewEngine(handlers)
	log.Fatal(e.Run(":8080"))
}
