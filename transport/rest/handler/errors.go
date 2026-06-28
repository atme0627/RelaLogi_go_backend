package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/atme0627/RelaLogi_go_backend/apperror"
	"github.com/atme0627/RelaLogi_go_backend/transport/rest/oapi"
	"github.com/gin-gonic/gin"
)

var kindStatus = map[apperror.Kind]int{
	apperror.KindBadRequest: http.StatusBadRequest,
	apperror.KindInternal:   http.StatusInternalServerError,
}

func writeError(ctx *gin.Context, err error) {
	var ae *apperror.Error
	var e oapi.Error
	status := http.StatusInternalServerError
	if !errors.As(err, &ae) {
		e = oapi.Error{Code: "INTERNAL_SERVER_ERROR", Message: "internal server error"}
	} else {
		e = oapi.Error{Code: ae.Code, Message: ae.Message}
		v, ok := kindStatus[ae.Kind]
		if ok {
			status = v
		}
	}

	log.Printf("error: %v", err)

	ctx.JSON(status, e)
}
