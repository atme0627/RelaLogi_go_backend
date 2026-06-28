package entity

import "math"

type EncodedImage struct {
	Bytes     []byte
	MimeTypes string
}
type Point struct {
	X int
	Y int
}

func (p *Point) Distance(q Point) float64 {
	return math.Sqrt(float64((p.X-q.X)*(p.X-q.X) + (p.Y-q.Y)*(p.Y-q.Y)))
}

type Quad struct {
	P1 Point
	P2 Point
	P3 Point
	P4 Point
}

func (q *Quad) SortClockwiseFromTopLeft() {
	var topLeft, topRight, bottomRight, bottomLeft Point
	maxSum := math.MinInt64
	minSum := math.MaxInt64
	maxDiff := math.MinInt64
	minDiff := math.MaxInt64

	for _, p := range []Point{q.P1, q.P2, q.P3, q.P4} {
		sum := p.X + p.Y
		diff := p.Y - p.X
		if sum > maxSum {
			maxSum = sum
			bottomRight = p
		}
		if sum < minSum {
			minSum = sum
			topLeft = p
		}
		if diff > maxDiff {
			maxDiff = diff
			bottomLeft = p
		}
		if diff < minDiff {
			minDiff = diff
			topRight = p
		}
	}

	q.P1 = topLeft
	q.P2 = topRight
	q.P3 = bottomRight
	q.P4 = bottomLeft
}

type PuzzleSize struct {
	Width       int
	Height      int
	VHintHeight int
	HHintWidth  int
}

type Puzzle struct {
	Size  PuzzleSize
	VHint [][]int
	HHint [][]int
}

func NewEncodedImage(bytes []byte, mimeTypes string) (EncodedImage, error) {
	return EncodedImage{Bytes: bytes, MimeTypes: mimeTypes}, nil
}
