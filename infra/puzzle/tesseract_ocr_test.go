package infra

import (
	"os"
	"testing"

	"github.com/atme0627/RelaLogi_go_backend/entity"
)

func Test_RecognizeNumberFromCell(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		imagePath string
		expected  int
	}{
		"正常系1: 1": {
			imagePath: "testdata/preprocessed_1.png",
			expected:  1,
		},
		"正常系1: 2": {
			imagePath: "testdata/preprocessed_2.png",
			expected:  2,
		},
		"正常系1: 4": {
			imagePath: "testdata/preprocessed_4.png",
			expected:  4,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			tesseractOCR := NewTesseractOCR()
			targetImageByte, err := os.ReadFile(tt.imagePath)
			if err != nil {
				t.Fatal(err)
			}
			targetImage, err := entity.NewEncodedImage(targetImageByte, "image/png")
			if err != nil {
				t.Fatal(err)
			}
			actual, err := tesseractOCR.RecognizeNumberFromCell(targetImage)
			if err != nil {
				t.Fatal(err)
			}
			err = tesseractOCR.Close()
			if err != nil {
				t.Fatal(err)
			}

			if actual != tt.expected {
				t.Errorf("expected: %d, actual: %d", tt.expected, actual)
			}
		})
	}
}
