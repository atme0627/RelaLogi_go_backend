package handler

import (
	"io"
	"net/http"

	"github.com/atme0627/RelaLogi_go_backend/apperror"
	"github.com/atme0627/RelaLogi_go_backend/controller"
	"github.com/atme0627/RelaLogi_go_backend/transport/rest/oapi"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
)

type PuzzleHandler struct {
	c *controller.PuzzleController
}

func NewPuzzleHandler(c *controller.PuzzleController) *PuzzleHandler {
	return &PuzzleHandler{
		c: c,
	}
}

func (h *PuzzleHandler) Register(r gin.IRoutes) {
	r.POST("/puzzles/recognize", h.recognize)
}

func (h *PuzzleHandler) recognize(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("puzzleImage")
	if err != nil {
		writeError(ctx, apperror.BadRequest("MISSING_PUZZLE_IMAGE", "パズル画像 (puzzleImage) が指定されていません。", err))
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		writeError(ctx, err)
		return
	}
	defer f.Close()

	imageBytes, err := io.ReadAll(f)
	if err != nil {
		writeError(ctx, err)
		return
	}

	s, ok := ctx.GetPostForm("hintParameter")
	if !ok {
		writeError(ctx, apperror.BadRequest("MISSING_HINT_PARAMETER", "ヒントパラメータ (hintParameter) が指定されていません。", nil))
		return
	}

	var hintParameter oapi.HintParameter
	err = json.Unmarshal([]byte(s), &hintParameter)
	if err != nil {
		writeError(ctx, apperror.BadRequest("INVALID_HINT_PARAMETER", "ヒントパラメータの形式が不正です。JSON を確認してください。", err))
		return
	}

	recognizedPuzzle, err := h.c.RecognizeFromImage(ctx.Request.Context(), imageBytes, hintParameter)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, recognizedPuzzle)
}
