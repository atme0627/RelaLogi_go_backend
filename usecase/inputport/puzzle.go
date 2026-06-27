package inputport

import (
	"context"

	"github.com/atme0627/RelaLogi_go_backend/entity"
)

type Puzzle interface {
	RecognizeFromImage(ctx context.Context, image entity.EncodedImage, vHintQuad entity.Quad, hHintQuad entity.Quad, size entity.PuzzleSize) (*entity.Puzzle, [2]entity.EncodedImage, error)
	Create(ctx context.Context) error
}
