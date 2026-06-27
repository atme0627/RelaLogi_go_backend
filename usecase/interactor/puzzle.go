package interactor

import (
	"context"

	"github.com/atme0627/RelaLogi_go_backend/entity"
	"github.com/atme0627/RelaLogi_go_backend/usecase/port/puzzle"
)

type PuzzleInteractor struct {
	imageProcessor port.ImageProcessor
	ocr            port.OCR
}

func New(imageProcesser port.ImageProcessor, ocr port.OCR) *PuzzleInteractor {
	return &PuzzleInteractor{imageProcessor: imageProcesser, ocr: ocr}
}

func (i PuzzleInteractor) FromImage(ctx context.Context, image entity.EncodedImage, vHintQuad entity.Quad, hHintQuad entity.Quad, size entity.PuzzleSize) (*entity.Puzzle, [2]entity.EncodedImage, error) {
	const TRIM_PIXEL = 2
	vHintImage, err := i.imageProcessor.CropHintsFromImage(image, vHintQuad)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}
	hHintImage, err := i.imageProcessor.CropHintsFromImage(image, hHintQuad)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}

	vHint, err := i.recognizeHintFromImage(ctx, vHintImage, vHintQuad, size.VHintHeight, size.Width, TRIM_PIXEL)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}
	hHint, err := i.recognizeHintFromImage(ctx, hHintImage, hHintQuad, size.Height, size.HHintWidth, TRIM_PIXEL)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}

	draftPuzzle := &entity.Puzzle{Size: size, VHint: vHint, HHint: hHint}
	err = i.ocr.Close()
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}
	return draftPuzzle, [2]entity.EncodedImage{vHintImage, hHintImage}, nil
}

func (i PuzzleInteractor) Create(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (i PuzzleInteractor) recognizeHintFromImage(ctx context.Context, hintImage entity.EncodedImage, hintQuad entity.Quad, height int, width int, trimPixel int) ([][]int, error) {
	vHintCells := i.imageProcessor.SplitHintToCells(hintImage, height, width)

	result := make([][]int, height)
	for j := range result {
		result[j] = make([]int, width)
	}
	for j, rows := range vHintCells {
		for k, cell := range rows {
			recognizedNumber := 0
			preprocessedCells, err := i.imageProcessor.PreprocessAndSplitCellToDigits(cell, trimPixel)
			if err != nil {
				return nil, err
			}

			for _, preprocessedCell := range preprocessedCells {
				ocrResult, err := i.ocr.RecognizeNumberFromCell(preprocessedCell)
				if err != nil {
					return nil, err
				}
				recognizedNumber = recognizedNumber*10 + ocrResult
				if recognizedNumber == -1 {
					continue
				}
			}
			result[j][k] = recognizedNumber
		}
	}
	return result, nil
}
