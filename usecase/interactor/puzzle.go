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

	vHintImage, hHintImage, err := i.imageProcessor.CropHintsFromImage(image, vHintQuad, hHintQuad)
	if err != nil {
		return nil, [2]entity.EncodedImage{}, err
	}

	vHintCells := i.imageProcessor.SplitHintToCells(vHintImage, size)
	hHintCells := i.imageProcessor.SplitHintToCells(hHintImage, size)

	vHint := make([][]int, size.VHintHeight)
	for j := range vHint {
		vHint[j] = make([]int, size.Width)
	}
	for j, rows := range vHintCells {
		for k, cell := range rows {
			recognizedNumber := 0
			preprocessedCells, err := i.imageProcessor.PreprocessAndSplitCellToDigits(cell, TRIM_PIXEL)
			if err != nil {
				return nil, [2]entity.EncodedImage{}, err
			}

			for _, preprocessedCell := range preprocessedCells {
				ocrResult, err := i.ocr.RecognizeNumberFromCell(preprocessedCell)
				if err != nil {
					return nil, [2]entity.EncodedImage{}, err
				}
				recognizedNumber = recognizedNumber*10 + ocrResult
				if recognizedNumber == -1 {
					continue
				}
			}

			vHint[j][k] = recognizedNumber
		}

	}

	hHint := make([][]int, size.Height)
	for j := range hHint {
		hHint[j] = make([]int, size.HHintWidth)
	}
	for j, rows := range hHintCells {
		for k, cell := range rows {
			recognizedNumber := 0
			preprocessedCells, err := i.imageProcessor.PreprocessAndSplitCellToDigits(cell, TRIM_PIXEL)
			if err != nil {
				return nil, [2]entity.EncodedImage{}, err
			}

			for _, preprocessedCell := range preprocessedCells {
				ocrResult, err := i.ocr.RecognizeNumberFromCell(preprocessedCell)
				if err != nil {
					return nil, [2]entity.EncodedImage{}, err
				}
				recognizedNumber = recognizedNumber*10 + ocrResult
				if recognizedNumber == -1 {
					continue
				}
			}

			vHint[j][k] = recognizedNumber
		}

	}

	draftPuzzle := &entity.Puzzle{Size: size, VHint: vHint, HHint: hHint}
	return draftPuzzle, [2]entity.EncodedImage{vHintImage, hHintImage}, nil
}

func (i PuzzleInteractor) Create(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
