package port

import "github.com/atme0627/RelaLogi_go_backend/entity"

type ImageProcessor interface {
	CropHintsFromImage(image entity.EncodedImage, vHintQuad entity.Quad, hHintQuad entity.Quad) (entity.EncodedImage, entity.EncodedImage, error)
	SplitHintToCells(hint entity.EncodedImage, height int, width int) [][]entity.EncodedImage
	PreprocessAndSplitCellToDigits(cell entity.EncodedImage, trimPixel int) ([]entity.EncodedImage, error)
}
