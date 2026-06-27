package controller

import (
	"context"
	"fmt"

	"github.com/atme0627/RelaLogi_go_backend/apperror"
	"github.com/atme0627/RelaLogi_go_backend/entity"
	"github.com/atme0627/RelaLogi_go_backend/transport/rest/oapi"
	"github.com/atme0627/RelaLogi_go_backend/usecase/inputport"
)

type PuzzleController struct {
	puzzle inputport.Puzzle
}

func NewPuzzleController(puzzle inputport.Puzzle) *PuzzleController {
	return &PuzzleController{
		puzzle: puzzle,
	}
}

func (c *PuzzleController) RecognizeFromImage(ctx context.Context, puzzleImageByte []byte, vHintSize oapi.GridSize, hHintSize oapi.GridSize, vHintRegion oapi.Quad, hHintRegion oapi.Quad) (oapi.Puzzle, error) {
	puzzleImage, err := entity.NewEncodedImage(puzzleImageByte, "image/png")
	if err != nil {
		return oapi.Puzzle{}, apperror.Internal("IMAGE_CROP_FAILED", "failed to create puzzle image: %w", err)
	}

	puzzleSize := entity.PuzzleSize{
		Width:       vHintSize.Cols,
		Height:      hHintSize.Rows,
		VHintHeight: vHintSize.Rows,
		HHintWidth:  hHintSize.Cols,
	}

	vHintQuad, err := toEntityQuad(vHintRegion)
	if err != nil {
		return oapi.Puzzle{}, fmt.Errorf("failed to convert hint quad: %w", err)
	}
	hHintQuad, err := toEntityQuad(hHintRegion)
	if err != nil {
		return oapi.Puzzle{}, fmt.Errorf("failed to convert hint quad: %w", err)
	}

	recognizedPuzzle, croppedImages, err := c.puzzle.RecognizeFromImage(ctx, puzzleImage, vHintQuad, hHintQuad, puzzleSize)
	if err != nil {
		return oapi.Puzzle{}, apperror.Internal("HINT_RECOGNIZE_FAILED", "failed to recognize puzzle: %w", err)
	}

	verticalHintGrid, horizontalHintGrid := toOapiHintGrid(*recognizedPuzzle)
	verticalHintImage := toDataURI(croppedImages[0])
	horizontalHintImage := toDataURI(croppedImages[1])
	return oapi.Puzzle{
		HorizontalHintGrid:  horizontalHintGrid,
		HorizontalHintImage: horizontalHintImage,
		VerticalHintGrid:    verticalHintGrid,
		VerticalHintImage:   verticalHintImage,
	}, nil
}
