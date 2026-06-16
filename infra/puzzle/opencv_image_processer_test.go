package infra

import (
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
