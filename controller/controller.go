package controller

import (
	"net/http"
)

type Controller struct {
}

func New() *Controller {
	return &Controller{}
}

// 一旦、net/httpへの依存は気にしない。
func (c *Controller) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status":"ok"}`))
	if err != nil {
		return
	}
}
