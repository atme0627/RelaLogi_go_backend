package main

import (
	"net/http"

	"github.com/atme0627/RelaLogi_go_backend/framework"
)

func main() {
	mux := framework.InitRoute()
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
	}
}
