package infra

import (
	"bytes"
	"image"

	"github.com/atme0627/RelaLogi_go_backend/entity"

	"gocv.io/x/gocv"
)

type OpenCVImageProcessor struct {
}

func (o OpenCVImageProcessor) CropHintsFromImage(img entity.EncodedImage, vHintQuad entity.Quad, hHintQuad entity.Quad) (entity.EncodedImage, entity.EncodedImage, error) {
	closeMat := func(mat *gocv.Mat) {
		err := mat.Close()
		if err != nil {
		}
	}

	mat, err := gocv.IMDecode(img.Bytes, gocv.IMReadColor)
	if err != nil {
		return entity.EncodedImage{}, entity.EncodedImage{}, err
	}
	defer closeMat(&mat)

	vHintQuad.SortClockwiseFromTopLeft()
	hHintQuad.SortClockwiseFromTopLeft()

	vRectified := o.getRectifiedImageQuad(vHintQuad)
	hRectified := o.getRectifiedImageQuad(hHintQuad)

	pvVFrom := gocv.NewPointVectorFromPoints([]image.Point{image.Point(vHintQuad.P1), image.Point(vHintQuad.P2), image.Point(vHintQuad.P3), image.Point(vHintQuad.P4)})
	pvHFrom := gocv.NewPointVectorFromPoints([]image.Point{image.Point(hHintQuad.P1), image.Point(hHintQuad.P2), image.Point(hHintQuad.P3), image.Point(hHintQuad.P4)})
	pvVTo := gocv.NewPointVectorFromPoints([]image.Point{image.Point(vRectified.P1), image.Point(vRectified.P2), image.Point(vRectified.P3), image.Point(vRectified.P4)})
	pvHTo := gocv.NewPointVectorFromPoints([]image.Point{image.Point(hRectified.P1), image.Point(hRectified.P2), image.Point(hRectified.P3), image.Point(hRectified.P4)})
	defer pvVFrom.Close()
	defer pvHFrom.Close()
	defer pvVTo.Close()
	defer pvHTo.Close()

	vM := gocv.GetPerspectiveTransform(pvVFrom, pvVTo)
	hM := gocv.GetPerspectiveTransform(pvHFrom, pvHTo)

	vCroppedMatDest := gocv.NewMat()
	hCroppedMatDest := gocv.NewMat()
	defer closeMat(&vCroppedMatDest)
	defer closeMat(&hCroppedMatDest)

	err = gocv.WarpPerspective(mat, &vCroppedMatDest, vM, image.Point(vRectified.P3))
	if err != nil {
		return entity.EncodedImage{}, entity.EncodedImage{}, err
	}
	err = gocv.WarpPerspective(mat, &hCroppedMatDest, hM, image.Point(hRectified.P3))
	if err != nil {
		return entity.EncodedImage{}, entity.EncodedImage{}, err
	}

	vBuf, err := gocv.IMEncode(gocv.PNGFileExt, vCroppedMatDest)
	if err != nil {
		return entity.EncodedImage{}, entity.EncodedImage{}, err
	}
	defer vBuf.Close()
	hBuf, err := gocv.IMEncode(gocv.PNGFileExt, hCroppedMatDest)
	if err != nil {
		return entity.EncodedImage{}, entity.EncodedImage{}, err
	}
	defer hBuf.Close()

	vCroppedImage, err := entity.NewEncodedImage(bytes.Clone(vBuf.GetBytes()), img.MimeTypes)
	if err != nil {
		return entity.EncodedImage{}, entity.EncodedImage{}, err
	}
	hCroppedImage, err := entity.NewEncodedImage(bytes.Clone(hBuf.GetBytes()), img.MimeTypes)
	if err != nil {
		return entity.EncodedImage{}, entity.EncodedImage{}, err
	}

	return vCroppedImage, hCroppedImage, nil
}

func (o OpenCVImageProcessor) SplitHintToCells(hint entity.EncodedImage, size entity.PuzzleSize) [][]entity.EncodedImage {
	//TODO implement me
	panic("implement me")
}

func (o OpenCVImageProcessor) getRectifiedImageQuad(hintQuad entity.Quad) entity.Quad {
	// clockwiseでソート済みの想定
	topWidth := hintQuad.P1.Distance(hintQuad.P2)
	bottomWidth := hintQuad.P3.Distance(hintQuad.P4)
	leftHeight := hintQuad.P1.Distance(hintQuad.P4)
	rightHeight := hintQuad.P2.Distance(hintQuad.P3)

	height := int(max(leftHeight, rightHeight))
	width := int(max(topWidth, bottomWidth))

	return entity.Quad{entity.Point{0, 0}, entity.Point{width, 0}, entity.Point{width, height}, entity.Point{0, height}}
}
