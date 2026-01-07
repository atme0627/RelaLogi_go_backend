package port

import "github.com/atme0627/RelaLogi_go_backend/entity"

type OCR interface {
	RecognizeNumbersFromCells(hintCells [][]entity.EncodedImage) ([][]int, error)
}
