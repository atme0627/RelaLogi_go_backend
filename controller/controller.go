package controller

import (
	"fmt"
	"net/http"
)

type Controller struct {
}

func New() *Controller {
	return &Controller{}
}

// 一旦、net/httpへの依存は気にしない。
func (c *Controller) Health(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ok")
}
