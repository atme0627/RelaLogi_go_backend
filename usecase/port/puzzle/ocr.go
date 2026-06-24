package port

import "github.com/atme0627/RelaLogi_go_backend/entity"

type OCR interface {
	RecognizeNumberFromCell(hintCell entity.EncodedImage) (int, error)
}
