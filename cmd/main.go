package main

import (
	"log"

	"github.com/atme0627/RelaLogi_go_backend/controller"
	"github.com/atme0627/RelaLogi_go_backend/framework/adapter/gin"
	"github.com/atme0627/RelaLogi_go_backend/framework/adapter/gin/handler"
)

func main() {
	healthController := controller.NewHealthController()
	healthHandler := handler.NewHandler(healthController)
	handlers := gin.Handlers{
		Health: healthHandler,
	}
	e := gin.NewEngine(handlers)
	log.Fatal(e.Run(":8080"))
}
