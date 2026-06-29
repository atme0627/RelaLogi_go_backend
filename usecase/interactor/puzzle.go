package interactor

import (
	"context"

	"github.com/atme0627/RelaLogi_go_backend/entity"
	"github.com/atme0627/RelaLogi_go_backend/usecase/gateway/puzzle"
)

type PuzzleInteractor struct {
	imageProcessor port.ImageProcessor
	ocr            port.OCR
}

func New(imageProcesser port.ImageProcessor, ocr port.OCR) *PuzzleInteractor {
	return &PuzzleInteractor{imageProcessor: imageProcesser, ocr: ocr}
}

func (i PuzzleInteractor) RecognizeFromImage(ctx context.Context, image entity.EncodedImage, vHintQuad entity.Quad, hHintQuad entity.Quad, size entity.PuzzleSize) (*entity.Puzzle, [2]entity.EncodedImage, error) {
	vHintImage, err := i.imageProcessor.CropHintsFromImage(image, vHintQuad)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}
	hHintImage, err := i.imageProcessor.CropHintsFromImage(image, hHintQuad)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}

	vHint, err := i.recognizeHintFromImage(vHintImage, size.VHintHeight, size.Width)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}
	hHint, err := i.recognizeHintFromImage(hHintImage, size.Height, size.HHintWidth)
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

func (i PuzzleInteractor) recognizeHintFromImage(hintImage entity.EncodedImage, height int, width int) ([][]int, error) {
	vHintCells := i.imageProcessor.SplitHintToCells(hintImage, height, width)

	result := make([][]int, height)
	for j := range result {
		result[j] = make([]int, width)
	}
	for j, rows := range vHintCells {
		for k, cell := range rows {
			recognizedNumber := 0
			preprocessedCells, err := i.imageProcessor.PreprocessAndSplitCellToDigits(cell)
			if err != nil {
				return nil, err
			}

			if len(preprocessedCells) == 0 {
				recognizedNumber = -1
			}

			for _, preprocessedCell := range preprocessedCells {
				ocrResult, err := i.ocr.RecognizeNumberFromCell(preprocessedCell)
				if err != nil {
					return nil, err
				}
				if ocrResult == -1 {
					recognizedNumber = -1
					break
				}
				recognizedNumber = recognizedNumber*10 + ocrResult
			}
			result[j][k] = recognizedNumber
		}
	}
	return result, nil
}
