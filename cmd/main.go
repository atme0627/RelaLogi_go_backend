package main

import (
	"log"

	"github.com/atme0627/RelaLogi_go_backend/controller"
	infra "github.com/atme0627/RelaLogi_go_backend/infra/puzzle"
	"github.com/atme0627/RelaLogi_go_backend/transport/rest"
	"github.com/atme0627/RelaLogi_go_backend/transport/rest/handler"
	"github.com/atme0627/RelaLogi_go_backend/usecase/interactor"
)

func main() {
	healthController := controller.NewHealthController()
	healthHandler := handler.NewHandler(healthController)

	infraImageProcessor := infra.NewOpenCVImageProcessor()
	infraOCR := infra.NewTesseractOCR()
	puzzleInteractor := interactor.New(infraImageProcessor, infraOCR)
	puzzleController := controller.NewPuzzleController(puzzleInteractor)
	puzzleHandler := handler.NewPuzzleHandler(puzzleController)

	handlers := rest.Handlers{
		Health: healthHandler,
		Puzzle: puzzleHandler,
	}
	e := rest.NewEngine(handlers)
	log.Fatal(e.Run(":8080"))
}
