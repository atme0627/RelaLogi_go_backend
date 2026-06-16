package infra

import "github.com/atme0627/RelaLogi_go_backend/entity"

type OpenCVImageProcessor struct {
}

func (o OpenCVImageProcessor) CropHintsFromImage(image entity.EncodedImage, vHintQuad entity.Quad, hHintQuad entity.Quad) (entity.EncodedImage, entity.EncodedImage, error) {
	//TODO implement me
	panic("implement me")
}

func (o OpenCVImageProcessor) SplitHintToCells(hint entity.EncodedImage, size entity.PuzzleSize) [][]entity.EncodedImage {
	//TODO implement me
	panic("implement me")
}
