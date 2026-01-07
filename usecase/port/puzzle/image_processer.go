package port

import "github.com/atme0627/RelaLogi_go_backend/entity"

type ImageProcesser interface {
	CropHintsFromImage(image entity.EncodedImage, vHintQuad entity.Quad, hHintQuad entity.Quad) (entity.EncodedImage, entity.EncodedImage, error)
	SplitHintToCells(hint entity.EncodedImage, size entity.PuzzleSize) [][]entity.EncodedImage
}
