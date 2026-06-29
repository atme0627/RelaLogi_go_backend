package infra

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"testing"

	"github.com/atme0627/RelaLogi_go_backend/entity"
)

func Test_getRectifiedImageQuad(t *testing.T) {
	t.Parallel()
	openCVImageProcesser := OpenCVImageProcessor{}
	tests := map[string]struct {
		in       entity.Quad
		expected entity.Quad
	}{
		"正常系 {(0, 0), (1, 0), (2, 1), (1, 1)} -> {(0, 0), (1, 0), (1, 1), (0, 1)}": {
			entity.Quad{
				entity.Point{0, 0},
				entity.Point{1, 0},
				entity.Point{2, 1},
				entity.Point{1, 1},
			},
			entity.Quad{
				entity.Point{0, 0},
				entity.Point{1, 0},
				entity.Point{1, 1},
				entity.Point{0, 1},
			},
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			actual := openCVImageProcesser.getRectifiedImageQuad(tt.in)
			if actual != tt.expected {
				t.Errorf("expected: %v, actual: %v", tt.expected, actual)
			}
		})
	}
}

func Test_CropHintsFromImage(t *testing.T) {
	t.Parallel()
	openCVImageProcesser := OpenCVImageProcessor{}
	sampleImageByte, err := os.ReadFile("testdata/warped_two_rects.png")
	if err != nil {
		t.Fatal(err)
	}
	sampleImage, err := entity.NewEncodedImage(sampleImageByte, "image/png")
	if err != nil {
		t.Fatal(err)
	}

	type in struct {
		encodedImage entity.EncodedImage
		hintQuad     entity.Quad
	}

	type expected struct {
		height int
		width  int
	}

	tests := map[string]struct {
		in       in
		expected expected
	}{
		"正常系1": {
			in{
				encodedImage: sampleImage,
				hintQuad:     entity.Quad{entity.Point{107, 257}, entity.Point{273, 278}, entity.Point{263, 382}, entity.Point{98, 360}},
			},
			expected{
				height: 104,
				width:  167,
			},
		},
		"正常系2": {
			in{
				encodedImage: sampleImage,
				hintQuad:     entity.Quad{entity.Point{285, 156}, entity.Point{435, 175}, entity.Point{422, 298}, entity.Point{273, 278}},
			},
			expected{
				height: 123,
				width:  151,
			},
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			actual, err := openCVImageProcesser.CropHintsFromImage(tt.in.encodedImage, tt.in.hintQuad)
			if err != nil {
				t.Fatal(err)
			}
			encodedConfig, _, err := image.DecodeConfig(bytes.NewReader(actual.Bytes))
			if err != nil {
				t.Fatal(err)
			}

			if encodedConfig.Width != tt.expected.width || encodedConfig.Height != tt.expected.height {
				t.Errorf("expected: (width: %d, height: %d), actual: (width: %d, height: %d)", tt.expected.width, tt.expected.height, encodedConfig.Width, encodedConfig.Height)
			}
		})
	}
}

func Test_SplitHintToCells(t *testing.T) {
	t.Parallel()
	openCVImageProcesser := OpenCVImageProcessor{}
	sampleImageByte, err := os.ReadFile("testdata/checkerboard_3x2.png")
	if err != nil {
		t.Fatal(err)
	}
	sampleImage, err := entity.NewEncodedImage(sampleImageByte, "image/png")
	if err != nil {
		t.Fatal(err)
	}

	type in struct {
		encodedImage entity.EncodedImage
		height       int
		width        int
	}

	type expected struct {
		height int
		width  int
	}

	tests := map[string]struct {
		in       in
		expected expected
	}{
		// 角セル[0][0]は左/上はクランプ、右/下に overlap(セル50pxの10%=5px)ぶん広がる→55x55
		"正常系": {
			in{
				encodedImage: sampleImage,
				height:       2,
				width:        3,
			},
			expected{
				height: 55,
				width:  55,
			},
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			actual := openCVImageProcesser.SplitHintToCells(tt.in.encodedImage, tt.in.height, tt.in.width)
			decodedConfig, _, err := image.DecodeConfig(bytes.NewReader(actual[0][0].Bytes))
			if err != nil {
				t.Fatal(err)
			}
			if decodedConfig.Width != tt.expected.width || decodedConfig.Height != tt.expected.height {
				t.Errorf("expected: (width: %d, height: %d), actual: (width: %d, height: %d)", tt.expected.width, tt.expected.height, decodedConfig.Width, decodedConfig.Height)
			}
		})
	}
}

func Test_PreprocessAndSplitCellToDigits(t *testing.T) {
	t.Parallel()
	openCVImageProcesser := OpenCVImageProcessor{}

	type in struct {
		imageName string
	}

	type expected struct {
		componentCount int
	}

	tests := map[string]struct {
		in       in
		expected expected
	}{
		"正常系: 14": {
			in{
				imageName: "testdata/number14.png",
			},
			expected{
				componentCount: 2,
			},
		},
		"正常系: 2": {
			in{
				imageName: "testdata/number2.png",
			},
			expected{
				componentCount: 1,
			},
		},
		"正常系: 空白1": {
			in{
				imageName: "testdata/blankCell1.png",
			},
			expected{
				componentCount: 0,
			},
		},
		"正常系: 空白2": {
			in{
				imageName: "testdata/blankCell2.png",
			},
			expected{
				componentCount: 0,
			},
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			targetImageByte, err := os.ReadFile(tt.in.imageName)
			if err != nil {
				t.Fatal(err)
			}
			targetImage, err := entity.NewEncodedImage(targetImageByte, "image/png")
			if err != nil {
				t.Fatal(err)
			}

			actual, err := openCVImageProcesser.PreprocessAndSplitCellToDigits(targetImage)
			if err != nil {
				t.Fatal(err)
			}

			for i, d := range actual {
				err := os.WriteFile(fmt.Sprintf("testdata/output/%s_preprocessed_%d.png", testName, i), d.Bytes, 0644)
				if err != nil {
					return
				}
			}

			if len(actual) != tt.expected.componentCount {
				t.Errorf("expected count: %d, actual count: %d", tt.expected.componentCount, len(actual))
			}

		})
	}
}
