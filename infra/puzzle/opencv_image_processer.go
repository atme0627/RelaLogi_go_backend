package infra

import (
	"bytes"
	"image"
	"image/png"
	_ "image/png"
	"math"
	"sort"

	"github.com/atme0627/RelaLogi_go_backend/entity"

	"gocv.io/x/gocv"
)

type OpenCVImageProcessor struct {
}

func (o OpenCVImageProcessor) CropHintsFromImage(img entity.EncodedImage, vHintQuad entity.Quad, hHintQuad entity.Quad) (entity.EncodedImage, entity.EncodedImage, error) {
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
	result := make([][]entity.EncodedImage, size.Height)
	decodedImage, _, err := image.Decode(bytes.NewReader(hint.Bytes))
	if err != nil {
		return nil
	}

	xBoundary := func(i int) int {
		b := decodedImage.Bounds()
		return b.Min.X + i*b.Dx()/size.Width
	}
	yBoundary := func(i int) int {
		b := decodedImage.Bounds()
		return b.Min.Y + i*b.Dy()/size.Height
	}

	sub, ok := decodedImage.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok {
		return nil
	}

	for i := 0; i < size.Height; i++ {
		result[i] = make([]entity.EncodedImage, size.Width)
		for j := 0; j < size.Width; j++ {
			cell := sub.SubImage(image.Rect(xBoundary(j), yBoundary(i), xBoundary(j+1), yBoundary(i+1)))
			buf := &bytes.Buffer{}
			err := png.Encode(buf, cell)
			if err != nil {
				return nil
			}
			encodedCell, err := entity.NewEncodedImage(buf.Bytes(), hint.MimeTypes)
			if err != nil {
				return nil
			}
			result[i][j] = encodedCell
		}
	}
	return result
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

func (o OpenCVImageProcessor) PreprocessAndSplitCellToDigits(cell entity.EncodedImage, trimPixel int) ([]entity.EncodedImage, error) {
	const centralThreshold = 0.5 // セルの中央何%に含まれる連結成分を使用するかを決める。

	mat, err := gocv.IMDecode(cell.Bytes, gocv.IMReadColor)
	if err != nil {
		return nil, err
	}
	defer closeMat(&mat)

	//周辺のトリミング
	mat = mat.Region(image.Rect(trimPixel, trimPixel, mat.Cols()-trimPixel, mat.Rows()-trimPixel))

	//グレースケール化
	grayMat := gocv.NewMat()
	defer closeMat(&grayMat)
	err = gocv.CvtColor(mat, &grayMat, gocv.ColorBGRToGray)
	if err != nil {
		return nil, err
	}

	//二値化
	err = gocv.AdaptiveThreshold(grayMat, &mat, 255, gocv.AdaptiveThresholdGaussian, gocv.ThresholdBinaryInv, 31, 2)
	if err != nil {
		return nil, err
	}

	//連結成分の抽出
	labelsMat := gocv.NewMat()
	defer closeMat(&labelsMat)
	statsMat := gocv.NewMat()
	defer closeMat(&statsMat)
	centroidsMat := gocv.NewMat()
	defer closeMat(&centroidsMat)

	count := gocv.ConnectedComponentsWithStats(mat, &labelsMat, &statsMat, &centroidsMat)
	type component struct {
		rect  image.Rectangle
		area  int
		label int
	}

	var components []component

	for i := 1; i < count; i++ {
		area := int(statsMat.GetIntAt(i, int(gocv.CC_STAT_AREA)))
		if area < int(float64(mat.Rows()*mat.Cols())*0.05) {
			continue
		}
		x := int(statsMat.GetIntAt(i, int(gocv.CC_STAT_LEFT)))
		y := int(statsMat.GetIntAt(i, int(gocv.CC_STAT_TOP)))
		w := int(statsMat.GetIntAt(i, int(gocv.CC_STAT_WIDTH)))
		h := int(statsMat.GetIntAt(i, int(gocv.CC_STAT_HEIGHT)))

		components = append(components, component{rect: image.Rect(x, y, x+w, y+h), area: area, label: i})
	}
	sort.Slice(components, func(i, j int) bool { return components[i].area > components[j].area })
	digitsCount := int(math.Min(float64(len(components)), 2))
	components = components[:digitsCount]
	sort.Slice(components, func(i, j int) bool { return components[i].rect.Min.X < components[j].rect.Min.X })

	mask := gocv.NewMat()
	defer closeMat(&mask)

	var result []entity.EncodedImage
	for _, c := range components {
		err := gocv.InRangeWithScalar(labelsMat, gocv.Scalar{float64(c.label), 0, 0, 0}, gocv.Scalar{float64(c.label), 0, 0, 0}, &mask)
		if err != nil {
			return nil, err
		}

		//連結成分がグレースケールで実際に黒い場所か判断。空白セル時に、ほぼ白いのにノイズを拾う場合を除去
		meanMat := gocv.NewMat()
		defer closeMat(&meanMat)
		stdDevMat := gocv.NewMat()
		defer closeMat(&stdDevMat)
		err = gocv.MeanStdDevWithMask(grayMat, &meanMat, &stdDevMat, mask)
		if err != nil {
			return nil, err
		}
		mean := meanMat.GetDoubleAt(0, 0)
		if mean > 128 {
			continue
		}

		//連結成分がセルの中央に触れているかを判断。生き残った枠線を除去
		cols := float64(labelsMat.Cols())
		rows := float64(labelsMat.Rows())
		centralRect := image.Rect(int(math.Ceil(cols*(centralThreshold/2))), int(math.Ceil(rows*(centralThreshold/2))), int(math.Ceil(cols*(1-centralThreshold/2))), int(math.Ceil(rows*(1-centralThreshold/2))))
		centerCount := gocv.CountNonZero(mask.Region(centralRect))
		if centerCount == 0 {
			continue
		}

		buf, err := gocv.IMEncode(gocv.PNGFileExt, mask)
		if err != nil {
			return nil, err
		}
		defer buf.Close()
		encodedDigit, err := entity.NewEncodedImage(bytes.Clone(buf.GetBytes()), cell.MimeTypes)
		if err != nil {
			return nil, err
		}
		result = append(result, encodedDigit)

	}

	return result, nil
}

func closeMat(mat *gocv.Mat) {
	err := mat.Close()
	if err != nil {
	}
}
