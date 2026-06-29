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

func NewOpenCVImageProcessor() *OpenCVImageProcessor {
	return &OpenCVImageProcessor{}
}

func (o OpenCVImageProcessor) CropHintsFromImage(img entity.EncodedImage, hintQuad entity.Quad) (entity.EncodedImage, error) {
	mat, err := gocv.IMDecode(img.Bytes, gocv.IMReadColor)
	if err != nil {
		return entity.EncodedImage{}, err
	}
	defer closeMat(&mat)

	hintQuad.SortClockwiseFromTopLeft()

	rectified := o.getRectifiedImageQuad(hintQuad)

	pvFrom := gocv.NewPointVectorFromPoints([]image.Point{image.Point(hintQuad.P1), image.Point(hintQuad.P2), image.Point(hintQuad.P3), image.Point(hintQuad.P4)})
	pvTo := gocv.NewPointVectorFromPoints([]image.Point{image.Point(rectified.P1), image.Point(rectified.P2), image.Point(rectified.P3), image.Point(rectified.P4)})
	defer pvFrom.Close()
	defer pvTo.Close()

	M := gocv.GetPerspectiveTransform(pvFrom, pvTo)

	vCroppedMatDest := gocv.NewMat()
	hCroppedMatDest := gocv.NewMat()
	defer closeMat(&vCroppedMatDest)
	defer closeMat(&hCroppedMatDest)

	err = gocv.WarpPerspective(mat, &vCroppedMatDest, M, image.Point(rectified.P3))
	if err != nil {
		return entity.EncodedImage{}, err
	}

	buf, err := gocv.IMEncode(gocv.PNGFileExt, vCroppedMatDest)
	if err != nil {
		return entity.EncodedImage{}, err
	}
	defer buf.Close()

	croppedImage, err := entity.NewEncodedImage(bytes.Clone(buf.GetBytes()), img.MimeTypes)
	if err != nil {
		return entity.EncodedImage{}, err
	}

	return croppedImage, nil
}

func (o OpenCVImageProcessor) SplitHintToCells(hint entity.EncodedImage, height int, width int) [][]entity.EncodedImage {
	result := make([][]entity.EncodedImage, height)
	decodedImage, _, err := image.Decode(bytes.NewReader(hint.Bytes))
	if err != nil {
		return nil
	}

	xBoundary := func(i int) int {
		b := decodedImage.Bounds()
		return b.Min.X + i*b.Dx()/width
	}
	yBoundary := func(i int) int {
		b := decodedImage.Bounds()
		return b.Min.Y + i*b.Dy()/height
	}

	sub, ok := decodedImage.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok {
		return nil
	}

	// セルを上下左右に overlapRatio ぶん広げて切り出す(隣と重ねる)。
	// 透視補正の歪みで分割境界が実グリッドから少しずれても数字が見切れないようにするため。
	const overlapRatio = 0.10
	b := decodedImage.Bounds()
	marginX := int(math.Round(overlapRatio * float64(b.Dx()) / float64(width)))
	marginY := int(math.Round(overlapRatio * float64(b.Dy()) / float64(height)))

	for i := 0; i < height; i++ {
		result[i] = make([]entity.EncodedImage, width)
		for j := 0; j < width; j++ {
			cell := sub.SubImage(image.Rect(
				max(b.Min.X, xBoundary(j)-marginX), max(b.Min.Y, yBoundary(i)-marginY),
				min(b.Max.X, xBoundary(j+1)+marginX), min(b.Max.Y, yBoundary(i+1)+marginY),
			))
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
	gocv.Threshold(grayMat, &mat, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)

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

		err = gocv.BitwiseNot(mask, &mask)
		if err != nil {
			return nil, err
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
