package interactor

import (
	"context"

	"github.com/atme0627/RelaLogi_go_backend/entity"
	"github.com/atme0627/RelaLogi_go_backend/usecase/port/puzzle"
)

type PuzzleInteractor struct {
	imageProcesser port.ImageProcesser
	ocr            port.OCR
}

func New(imageProcesser port.ImageProcesser, ocr port.OCR) *PuzzleInteractor {
	return &PuzzleInteractor{imageProcesser: imageProcesser, ocr: ocr}
}

func (i PuzzleInteractor) FromImage(ctx context.Context, image entity.EncodedImage, vHintQuad entity.Quad, hHintQuad entity.Quad, size entity.PuzzleSize) (*entity.Puzzle, [2]entity.EncodedImage, error) {
	vHintImage, hHintImage, err := i.imageProcesser.CropHintsFromImage(image, vHintQuad, hHintQuad)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}

	vHintCells := i.imageProcesser.SplitHintToCells(vHintImage, size)
	hHintCells := i.imageProcesser.SplitHintToCells(hHintImage, size)

	vHint, err := i.ocr.RecognizeNumbersFromCells(vHintCells)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}
	hHint, err := i.ocr.RecognizeNumbersFromCells(hHintCells)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}

	draftPuzzle := &entity.Puzzle{Size: size, VHint: vHint, HHint: hHint}
	return draftPuzzle, [2]entity.EncodedImage{vHintImage, hHintImage}, nil
}

func (i PuzzleInteractor) Create(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
