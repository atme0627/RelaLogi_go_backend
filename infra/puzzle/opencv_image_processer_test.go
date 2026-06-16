package infra

import (
	"bytes"
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

func TestOpenCVImageProcessor_CropHintsFromImage(t *testing.T) {
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
		EncodedImage entity.EncodedImage
		vHintQuad    entity.Quad
		hHintQuad    entity.Quad
	}

	type expected struct {
		vHeight int
		vWidth  int
		hHeight int
		hWidth  int
	}

	tests := map[string]struct {
		in       in
		expected expected
	}{
		"正常系": {
			in{
				EncodedImage: sampleImage,
				vHintQuad:    entity.Quad{entity.Point{107, 257}, entity.Point{273, 278}, entity.Point{263, 382}, entity.Point{98, 360}},
				hHintQuad:    entity.Quad{entity.Point{285, 156}, entity.Point{435, 175}, entity.Point{422, 298}, entity.Point{273, 278}},
			},
			expected{
				vHeight: 104,
				vWidth:  167,
				hHeight: 123,
				hWidth:  151,
			},
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			vActual, hActual, err := openCVImageProcesser.CropHintsFromImage(tt.in.EncodedImage, tt.in.vHintQuad, tt.in.hHintQuad)
			if err != nil {
				t.Fatal(err)
			}
			vEncodedConfig, _, err := image.DecodeConfig(bytes.NewReader(vActual.Bytes))
			if err != nil {
				t.Fatal(err)
			}
			hEncodedConfig, _, err := image.DecodeConfig(bytes.NewReader(hActual.Bytes))
			if err != nil {
				t.Fatal(err)
			}
			if vEncodedConfig.Width != tt.expected.vWidth || vEncodedConfig.Height != tt.expected.vHeight {
				t.Errorf("expected: (width: %d, height: %d), actual: (width: %d, height: %d)", tt.expected.vWidth, tt.expected.vHeight, vEncodedConfig.Width, vEncodedConfig.Height)
			}
			if hEncodedConfig.Width != tt.expected.hWidth || hEncodedConfig.Height != tt.expected.hHeight {
				t.Errorf("expected: (width: %d, height: %d), actual: (width: %d, height: %d)", tt.expected.hWidth, tt.expected.hHeight, hEncodedConfig.Width, hEncodedConfig.Height)
			}
		})
	}
}
