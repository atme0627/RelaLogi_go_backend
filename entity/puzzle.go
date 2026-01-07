package entity

type EncodedImage struct {
	Bytes     []byte
	MimeTypes string
}
type Point struct {
	X int
	Y int
}

type Quad struct {
	P1 Point
	P2 Point
	P3 Point
	P4 Point
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
