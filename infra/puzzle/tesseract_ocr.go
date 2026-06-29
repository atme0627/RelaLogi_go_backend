package infra

import (
	"github.com/atme0627/RelaLogi_go_backend/entity"
	"github.com/otiai10/gosseract/v2"
)

type TesseractOCR struct {
	client *gosseract.Client
}

func NewTesseractOCR() *TesseractOCR {
	client := gosseract.NewClient()

	err := client.SetLanguage("eng")
	if err != nil {
		return nil
	}

	err = client.SetPageSegMode(gosseract.PSM_SINGLE_CHAR)
	if err != nil {
		return nil
	}

	return &TesseractOCR{
		client: client,
	}
}

func (o *TesseractOCR) RecognizeNumberFromCell(hintCell entity.EncodedImage) (int, error) {
	err := o.client.SetImageFromBytes(hintCell.Bytes)
	if err != nil {
		return -1, err
	}

	ocrResult, err := o.client.Text()
	if err != nil {
		return -1, err
	}

	if ocrResult == "" || len(ocrResult) != 1 {
		return -1, nil
	}

	if '0' <= ocrResult[0] && ocrResult[0] <= '9' {
		return int(ocrResult[0] - '0'), nil
	}

	//別のアルファベットに誤認識した場合の対処
	switch ocrResult[0] {
	case 'o', 'O':
		return 0, nil
	case '|', 'l', 'I':
		return 1, nil
	case 'z', 'Z':
		return 2, nil
	case 's', 'S':
		return 5, nil
	case 'b':
		return 6, nil
	case 'B':
		return 8, nil
	case 'g':
		return 9, nil
	}
	return -1, nil
}

func (o *TesseractOCR) Close() error {
	err := o.client.Close()
	if err != nil {
		return err
	}
	return nil
}
