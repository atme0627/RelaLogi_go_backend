package infra

// 使い捨てデバッグ用テスト（本体コードではない）。
// sample1 をパイプラインに通し、各段を「元のグリッド形に並べた1枚絵」として
// testdata/output/<hint>/ に出力し、ground_truth.json と突合する。
//
//	実行: go test ./infra/puzzle/ -run Test_DumpE2E -v -vet=off
//
// 出力物（ヒントごと）:
//
//	crop.png          … Crop 段の結果
//	montage_cells.png … 分割セルを原寸でグリッド配置（白線区切り）
//	montage_bin.png   … 各セルの二値化直後
//	montage_comps.png … 上位2個に絞る "前" の全連結成分を色分け矩形で表示
//	                     赤=面積上位2個(本体が残す) / 緑=面積は通るが上位2個から漏れる(枠線に負けた数字)
//	                     青=面積フィルタ未満 / 黄=中央判定領域
//	result.txt        … 期待値グリッド vs 認識結果 vs DIFF
//
// 前処理は本体 PreprocessAndSplitCellToDigits を改変せず、ここにインライン再現している。
import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/atme0627/RelaLogi_go_backend/entity"
	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

// ocrRaw は Tesseract の生出力(trim済み)を返す。
func ocrRaw(c *gosseract.Client, b []byte) string {
	if err := c.SetImageFromBytes(b); err != nil {
		return ""
	}
	s, err := c.Text()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

// mapToDigit は本番 RecognizeNumberFromCell と同じマッピングで生出力を数字化する。
func mapToDigit(raw string) int {
	if len(raw) != 1 {
		return -1
	}
	ch := raw[0]
	if ch >= '0' && ch <= '9' {
		return int(ch - '0')
	}
	switch ch {
	case 'o', 'O':
		return 0
	case '|', 'l', 'I':
		return 1
	case 'z', 'Z':
		return 2
	case 's', 'S':
		return 5
	case 'b':
		return 6
	case 'B':
		return 8
	case 'g':
		return 9
	}
	return -1
}

const (
	dumpTrimPixel  = 0    // iso
	dumpOverlapPct = 0.10 // 実験: overlap分割の余白をセルサイズの何割にするか(片側)
	dumpSampleDir  = "../../testdata/puzzle"
	dumpOutputRoot = "testdata/output"
	dumpAreaRatio  = 0.03 // 本体と同じ: セル面積のこの割合未満の成分は無視
	dumpCentralThr = 0.5  // 本体と同じ: 中央何割を「中央」とみなすか
	dumpSep        = 2    // モンタージュの白線(ガター)幅
	dumpDigitPad   = 5    // 実験: OCRに渡す桁画像を bbox + この余白px で切り抜く
)

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// isFrameComp は成分が「上下両方 or 左右両方に接する＝枠線」かを判定する。
// 数字は小さくセルから両方向にははみ出さない前提。端の判定はセル寸法の10%を許容。
func isFrameComp(r image.Rectangle, cols, rows int) bool {
	tx := float64(cols) * 0.1
	ty := float64(rows) * 0.1
	top := float64(r.Min.Y) <= ty
	bottom := float64(r.Max.Y) >= float64(rows)-ty
	left := float64(r.Min.X) <= tx
	right := float64(r.Max.X) >= float64(cols)-tx
	return (top && bottom) || (left && right)
}

type gtHint struct {
	ExtractedImage string                   `json:"extractedImage"`
	Size           struct{ Rows, Cols int } `json:"size"`
	RegionVertices [][]int                  `json:"regionVertices"`
	Expected       [][]string               `json:"expected"`
}

type gtFile struct {
	PuzzleImage string            `json:"puzzleImage"`
	Hints       map[string]gtHint `json:"hints"`
}

func Test_DumpE2E(t *testing.T) {
	gtBytes, err := os.ReadFile(filepath.Join(dumpSampleDir, "ground_truth.json"))
	if err != nil {
		t.Fatal(err)
	}
	var gt gtFile
	if err := json.Unmarshal(gtBytes, &gt); err != nil {
		t.Fatal(err)
	}

	srcBytes, err := os.ReadFile(filepath.Join(dumpSampleDir, gt.PuzzleImage))
	if err != nil {
		t.Fatal(err)
	}
	srcImage, err := entity.NewEncodedImage(srcBytes, "image/png")
	if err != nil {
		t.Fatal(err)
	}

	proc := OpenCVImageProcessor{}
	// 本番と同じ設定(ホワイトリスト無し)で、生出力を観察する
	gocr := gosseract.NewClient()
	_ = gocr.SetLanguage("eng")
	_ = gocr.SetPageSegMode(gosseract.PSM_SINGLE_CHAR)
	defer gocr.Close()

	for name, h := range gt.Hints {
		t.Run(name, func(t *testing.T) {
			outDir := filepath.Join(dumpOutputRoot, name)
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				t.Fatal(err)
			}

			rv := h.RegionVertices
			quad := entity.Quad{
				P1: entity.Point{X: rv[0][0], Y: rv[0][1]},
				P2: entity.Point{X: rv[1][0], Y: rv[1][1]},
				P3: entity.Point{X: rv[2][0], Y: rv[2][1]},
				P4: entity.Point{X: rv[3][0], Y: rv[3][1]},
			}

			cropped, err := proc.CropHintsFromImage(srcImage, quad)
			if err != nil {
				t.Fatalf("Crop failed: %v", err)
			}
			writeDump(t, filepath.Join(outDir, "crop.png"), cropped.Bytes)

			rows, cols := h.Size.Rows, h.Size.Cols
			// 本番関数で検証(overlapは本番SplitHintToCellsに内蔵済み)
			cells := proc.SplitHintToCells(cropped, rows, cols)
			if len(cells) != rows {
				t.Fatalf("Split: 期待 %d 行, 実際 %d 行", rows, len(cells))
			}

			// 3 段ぶんのタイルを溜める
			cellTiles := make([][]gocv.Mat, rows)
			binTiles := make([][]gocv.Mat, rows)
			compTiles := make([][]gocv.Mat, rows)
			otsuTiles := make([][]gocv.Mat, rows)
			for r := range cellTiles {
				cellTiles[r] = make([]gocv.Mat, cols)
				binTiles[r] = make([]gocv.Mat, cols)
				compTiles[r] = make([]gocv.Mat, cols)
				otsuTiles[r] = make([]gocv.Mat, cols)
			}
			// Otsu+面積フィルタのみ(mean判定なし)で評価したときの集計
			var otsuBlankClean, otsuBlankNoise, otsuDigitOK, otsuDigitLost int

			got := make([][]string, rows)
			var total, match, blankCorrect, digitCorrect, falseFail, overDetect, valueMismatch int
			var ocrRawLog []string

			for r := 0; r < rows; r++ {
				got[r] = make([]string, cols)
				for c := 0; c < cols; c++ {
					cell := cells[r][c]

					// タイル生成（cells / bin / comps）
					cellTiles[r][c], binTiles[r][c], compTiles[r][c] = makeTiles(cell, dumpTrimPixel)

					// Otsu二値化(セル単位) + 面積フィルタのみ で何個の数字候補が残るか
					var otsuPass int
					otsuTiles[r][c], otsuPass = makeOtsuTile(cell, dumpTrimPixel)
					expCell := ""
					if r < len(h.Expected) && c < len(h.Expected[r]) {
						expCell = h.Expected[r][c]
					}
					switch {
					case expCell == "" && otsuPass == 0:
						otsuBlankClean++ // 空白セルが綺麗(過検出なし)
					case expCell == "" && otsuPass > 0:
						otsuBlankNoise++ // 空白セルにゴミが残る(過検出)
					case expCell != "" && otsuPass >= 1:
						otsuDigitOK++ // 数字セルで候補が残る(mean判定なしでも生存)
					default:
						otsuDigitLost++ // 数字セルなのに候補ゼロ
					}

					// 認識結果は「順序入替版」で（result.txt 用。本体は未改変）
					recognized := 0
					digits, err := proc.PreprocessAndSplitCellToDigits(cell)
					if err != nil {
						t.Fatalf("Preprocess failed at r%d c%d: %v", r, c, err)
					}
					if len(digits) == 0 {
						recognized = -1
					}
					var raws []string
					for _, dig := range digits {
						raw := ocrRaw(gocr, dig.Bytes)
						raws = append(raws, raw)
						res := mapToDigit(raw)
						if res == -1 {
							recognized = -1
							break
						}
						recognized = recognized*10 + res
					}
					gotStr := ""
					if recognized != -1 {
						gotStr = strconv.Itoa(recognized)
					}
					got[r][c] = gotStr

					// デバッグ: 数字セル or 候補が出たセルの Tesseract 生出力を記録
					if expCell != "" || len(raws) > 0 {
						ocrRawLog = append(ocrRawLog, fmt.Sprintf("r%02d c%02d exp=%-3s got=%-3q raw=%q", r, c, expCell, gotStr, raws))
					}

					exp := ""
					if r < len(h.Expected) && c < len(h.Expected[r]) {
						exp = h.Expected[r][c]
					}
					total++
					switch {
					case exp == gotStr:
						match++
						if exp == "" {
							blankCorrect++
						} else {
							digitCorrect++
						}
					case exp == "" && gotStr != "":
						overDetect++
					case exp != "" && gotStr == "":
						falseFail++
					default:
						valueMismatch++
					}
				}
			}

			writeMontage(t, filepath.Join(outDir, "montage_cells.png"), cellTiles)
			writeMontage(t, filepath.Join(outDir, "montage_bin.png"), binTiles)
			writeMontage(t, filepath.Join(outDir, "montage_comps.png"), compTiles)
			writeMontage(t, filepath.Join(outDir, "montage_bin_otsu.png"), otsuTiles)
			closeTiles(cellTiles, binTiles, compTiles, otsuTiles)

			writeResultText(t, filepath.Join(outDir, "result.txt"), h.Expected, got)

			// 失敗セル(GTに数字があるのに認識できなかった)が、どのフィルタで死んだかを文字で出す
			var why []byte
			why = append(why, "失敗セル(数字→空 / 値違い)が前処理のどの段で死んだか\n\n"...)
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					exp := ""
					if r < len(h.Expected) && c < len(h.Expected[r]) {
						exp = h.Expected[r][c]
					}
					if exp == "" || exp == got[r][c] {
						continue
					}
					reason := analyzeCell(cells[r][c], dumpTrimPixel)
					why = append(why, fmt.Sprintf("r%02d c%02d exp=%s got=%q\n%s\n", r, c, exp, got[r][c], reason)...)
				}
			}
			_ = os.WriteFile(filepath.Join(outDir, "why.txt"), why, 0o644)
			_ = os.WriteFile(filepath.Join(outDir, "ocr_raw.txt"), []byte(strings.Join(ocrRawLog, "\n")), 0o644)

			t.Logf("[%s] %d/%d 一致 (%.1f%%)", name, match, total, 100*float64(match)/float64(total))
			t.Logf("  空白正解: %d / 数字正解: %d", blankCorrect, digitCorrect)
			t.Logf("  見落とし(数字→空): %d / 過検出(空→数字): %d / 値違い: %d", falseFail, overDetect, valueMismatch)
			t.Logf("  [Otsu+面積のみ/mean判定なし] 数字セル: 候補残る%d / 候補ゼロ%d ｜ 空白セル: 綺麗%d / ゴミ残り%d",
				otsuDigitOK, otsuDigitLost, otsuBlankClean, otsuBlankNoise)
		})
	}
}

// extractDigitsReordered は本体 PreprocessAndSplitCellToDigits を「順序入替版」で再現する。
// 本体: 面積フィルタ → 上位2個に絞る → 中央判定で枠除去
// 本関数: 面積フィルタ → 中央判定で枠除去(先) → 上位2個に絞る
func extractDigitsReordered(cell entity.EncodedImage, trim int) ([]entity.EncodedImage, error) {
	mat, err := gocv.IMDecode(cell.Bytes, gocv.IMReadColor)
	if err != nil {
		return nil, err
	}
	defer closeMat(&mat)
	region := mat.Region(image.Rect(trim, trim, mat.Cols()-trim, mat.Rows()-trim))
	defer closeMat(&region)
	gray := gocv.NewMat()
	defer closeMat(&gray)
	_ = gocv.CvtColor(region, &gray, gocv.ColorBGRToGray)
	bin := gocv.NewMat()
	defer closeMat(&bin)
	gocv.Threshold(gray, &bin, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)

	labels := gocv.NewMat()
	defer closeMat(&labels)
	stats := gocv.NewMat()
	defer closeMat(&stats)
	centroids := gocv.NewMat()
	defer closeMat(&centroids)
	count := gocv.ConnectedComponentsWithStats(bin, &labels, &stats, &centroids)
	cellArea := float64(bin.Rows() * bin.Cols())

	cw, ch := float64(bin.Cols()), float64(bin.Rows())
	central := image.Rect(
		int(math.Ceil(cw*(dumpCentralThr/2))), int(math.Ceil(ch*(dumpCentralThr/2))),
		int(math.Ceil(cw*(1-dumpCentralThr/2))), int(math.Ceil(ch*(1-dumpCentralThr/2))),
	)

	type comp struct {
		label, area int
		rect        image.Rectangle
	}
	var kept []comp
	for i := 1; i < count; i++ {
		area := int(stats.GetIntAt(i, int(gocv.CC_STAT_AREA)))
		if float64(area) < cellArea*dumpAreaRatio { // 面積フィルタ
			continue
		}
		x := int(stats.GetIntAt(i, int(gocv.CC_STAT_LEFT)))
		y := int(stats.GetIntAt(i, int(gocv.CC_STAT_TOP)))
		w := int(stats.GetIntAt(i, int(gocv.CC_STAT_WIDTH)))
		hh := int(stats.GetIntAt(i, int(gocv.CC_STAT_HEIGHT)))
		rect := image.Rect(x, y, x+w, y+hh)
		if isFrameComp(rect, bin.Cols(), bin.Rows()) { // 枠線除去(上下 or 左右に接する)
			continue
		}
		m := gocv.NewMat()
		_ = gocv.InRangeWithScalar(labels, gocv.NewScalar(float64(i), 0, 0, 0), gocv.NewScalar(float64(i), 0, 0, 0), &m)
		cen := gocv.CountNonZero(m.Region(central))
		closeMat(&m)
		if cen == 0 { // 中央判定(残った枠除去)
			continue
		}
		kept = append(kept, comp{label: i, area: area, rect: rect})
	}
	// その後で上位2個に絞る
	sort.Slice(kept, func(a, b int) bool { return kept[a].area > kept[b].area })
	if len(kept) > 2 {
		kept = kept[:2]
	}
	sort.Slice(kept, func(a, b int) bool { return kept[a].rect.Min.X < kept[b].rect.Min.X })

	var result []entity.EncodedImage
	for _, c := range kept {
		m := gocv.NewMat()
		_ = gocv.InRangeWithScalar(labels, gocv.NewScalar(float64(c.label), 0, 0, 0), gocv.NewScalar(float64(c.label), 0, 0, 0), &m)
		// 実験: セル全面ではなく bbox + 周囲padで切り抜いてOCRに渡す
		x0 := clampInt(c.rect.Min.X-dumpDigitPad, 0, bin.Cols())
		y0 := clampInt(c.rect.Min.Y-dumpDigitPad, 0, bin.Rows())
		x1 := clampInt(c.rect.Max.X+dumpDigitPad, 0, bin.Cols())
		y1 := clampInt(c.rect.Max.Y+dumpDigitPad, 0, bin.Rows())
		roi := m.Region(image.Rect(x0, y0, x1, y1))
		crop := roi.Clone()
		roi.Close()
		closeMat(&m)
		_ = gocv.BitwiseNot(crop, &crop)
		buf, err := gocv.IMEncode(gocv.PNGFileExt, crop)
		if err != nil {
			closeMat(&crop)
			return nil, err
		}
		enc, _ := entity.NewEncodedImage(bytes.Clone(buf.GetBytes()), cell.MimeTypes)
		buf.Close()
		closeMat(&crop)
		result = append(result, enc)
	}
	return result, nil
}

// splitWithOverlap は本体 SplitHintToCells のぴったり分割に対し、各セルを
// セルサイズの pct ぶん上下左右へ広げて切り出す(隣と 2*pct 重なる)。画像端はクランプ。
func splitWithOverlap(t *testing.T, hint entity.EncodedImage, rows, cols int, pct float64) [][]entity.EncodedImage {
	t.Helper()
	img, _, err := image.Decode(bytes.NewReader(hint.Bytes))
	if err != nil {
		t.Fatal(err)
	}
	sub, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok {
		t.Fatal("画像が SubImage 未対応")
	}
	b := img.Bounds()
	xB := func(i int) int { return b.Min.X + i*b.Dx()/cols }
	yB := func(i int) int { return b.Min.Y + i*b.Dy()/rows }
	mx := int(math.Round(pct * float64(b.Dx()) / float64(cols)))
	my := int(math.Round(pct * float64(b.Dy()) / float64(rows)))
	clamp := func(v, lo, hi int) int {
		if v < lo {
			return lo
		}
		if v > hi {
			return hi
		}
		return v
	}

	result := make([][]entity.EncodedImage, rows)
	for i := 0; i < rows; i++ {
		result[i] = make([]entity.EncodedImage, cols)
		for j := 0; j < cols; j++ {
			x0 := clamp(xB(j)-mx, b.Min.X, b.Max.X)
			y0 := clamp(yB(i)-my, b.Min.Y, b.Max.Y)
			x1 := clamp(xB(j+1)+mx, b.Min.X, b.Max.X)
			y1 := clamp(yB(i+1)+my, b.Min.Y, b.Max.Y)
			cell := sub.SubImage(image.Rect(x0, y0, x1, y1))
			buf := &bytes.Buffer{}
			if err := png.Encode(buf, cell); err != nil {
				t.Fatal(err)
			}
			enc, err := entity.NewEncodedImage(buf.Bytes(), hint.MimeTypes)
			if err != nil {
				t.Fatal(err)
			}
			result[i][j] = enc
		}
	}
	return result
}

// makeTiles はセル1枚から cells/bin/comps の3タイル(いずれもBGR 8UC3)を作って返す。
// 本体 PreprocessAndSplitCellToDigits の前処理（トリム→グレー→AdaptiveThreshold→連結成分）を再現。
func makeTiles(cell entity.EncodedImage, trim int) (cellTile, binTile, compTile gocv.Mat) {
	src, err := gocv.IMDecode(cell.Bytes, gocv.IMReadColor)
	if err != nil {
		return gocv.NewMat(), gocv.NewMat(), gocv.NewMat()
	}
	defer closeMat(&src)

	// cells タイル = 原寸のセル（色）
	cellTile = src.Clone()

	// トリム
	region := src.Region(image.Rect(trim, trim, src.Cols()-trim, src.Rows()-trim))
	defer closeMat(&region)

	gray := gocv.NewMat()
	defer closeMat(&gray)
	_ = gocv.CvtColor(region, &gray, gocv.ColorBGRToGray)

	bin := gocv.NewMat()
	defer closeMat(&bin)
	// 本番と同じ: Otsu(セル単位グローバル) + BinaryInv(暗い数字を前景に)
	gocv.Threshold(gray, &bin, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)

	// bin タイル = 二値化結果をBGR化
	binTile = gocv.NewMat()
	_ = gocv.CvtColor(bin, &binTile, gocv.ColorGrayToBGR)

	// comps タイル = bin をベースに全連結成分を矩形描画
	compTile = gocv.NewMat()
	_ = gocv.CvtColor(bin, &compTile, gocv.ColorGrayToBGR)

	labels := gocv.NewMat()
	defer closeMat(&labels)
	stats := gocv.NewMat()
	defer closeMat(&stats)
	centroids := gocv.NewMat()
	defer closeMat(&centroids)
	count := gocv.ConnectedComponentsWithStats(bin, &labels, &stats, &centroids)

	cellArea := float64(bin.Rows() * bin.Cols())

	// 中央判定領域
	cw, ch := float64(bin.Cols()), float64(bin.Rows())
	central := image.Rect(
		int(math.Ceil(cw*(dumpCentralThr/2))), int(math.Ceil(ch*(dumpCentralThr/2))),
		int(math.Ceil(cw*(1-dumpCentralThr/2))), int(math.Ceil(ch*(1-dumpCentralThr/2))),
	)

	type comp struct {
		label    int
		area     int
		passArea bool
		frame    bool
		central  bool
	}
	var comps []comp
	for i := 1; i < count; i++ {
		area := int(stats.GetIntAt(i, int(gocv.CC_STAT_AREA)))
		passArea := float64(area) >= cellArea*dumpAreaRatio
		x := int(stats.GetIntAt(i, int(gocv.CC_STAT_LEFT)))
		y := int(stats.GetIntAt(i, int(gocv.CC_STAT_TOP)))
		w := int(stats.GetIntAt(i, int(gocv.CC_STAT_WIDTH)))
		hh := int(stats.GetIntAt(i, int(gocv.CC_STAT_HEIGHT)))
		frame := passArea && isFrameComp(image.Rect(x, y, x+w, y+hh), bin.Cols(), bin.Rows())
		cen := false
		if passArea && !frame {
			m := gocv.NewMat()
			_ = gocv.InRangeWithScalar(labels, gocv.NewScalar(float64(i), 0, 0, 0), gocv.NewScalar(float64(i), 0, 0, 0), &m)
			cen = gocv.CountNonZero(m.Region(central)) > 0
			closeMat(&m)
		}
		comps = append(comps, comp{label: i, area: area, passArea: passArea, frame: frame, central: cen})
	}

	// 新順序: 面積通過 かつ 非枠 かつ 中央接触 の中で面積上位2個が「採用」
	type ia struct{ idx, area int }
	var cands []ia
	for k, cp := range comps {
		if cp.passArea && !cp.frame && cp.central {
			cands = append(cands, ia{k, cp.area})
		}
	}
	sort.Slice(cands, func(a, b int) bool { return cands[a].area > cands[b].area })
	chosen := map[int]bool{}
	for i := 0; i < len(cands) && i < 2; i++ {
		chosen[cands[i].idx] = true
	}

	red := color.RGBA{R: 255, A: 255}
	green := color.RGBA{G: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}
	cyan := color.RGBA{G: 255, B: 255, A: 255}
	magenta := color.RGBA{R: 255, B: 255, A: 255}
	yellow := color.RGBA{R: 255, G: 255, A: 255}

	gocv.Rectangle(&compTile, central, yellow, 1)

	// 各成分を「輪郭にフィット」させて色分け描画
	for k, cp := range comps {
		var col color.RGBA
		switch {
		case !cp.passArea:
			col = blue // 面積フィルタ未満
		case cp.frame:
			col = magenta // 枠線(上下 or 左右に接する)
		case !cp.central:
			col = cyan // 中央非接触
		case chosen[k]:
			col = red // 採用(非枠・中央接触・上位2個)
		default:
			col = green // 中央接触だが上位2個外=漏れ
		}
		m := gocv.NewMat()
		_ = gocv.InRangeWithScalar(labels, gocv.NewScalar(float64(cp.label), 0, 0, 0), gocv.NewScalar(float64(cp.label), 0, 0, 0), &m)
		contours := gocv.FindContours(m, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		gocv.DrawContours(&compTile, contours, -1, col, 1)
		contours.Close()
		closeMat(&m)
	}
	return cellTile, binTile, compTile
}

// makeOtsuTile はセルをOtsu(セル単位グローバル二値化)で前景抽出し、
// 面積フィルタ(5%)だけを通した数字候補数と、可視化タイル(BGR)を返す。mean判定は使わない。
func makeOtsuTile(cell entity.EncodedImage, trim int) (tile gocv.Mat, areaPass int) {
	src, err := gocv.IMDecode(cell.Bytes, gocv.IMReadColor)
	if err != nil {
		return gocv.NewMat(), 0
	}
	defer closeMat(&src)
	region := src.Region(image.Rect(trim, trim, src.Cols()-trim, src.Rows()-trim))
	defer closeMat(&region)

	gray := gocv.NewMat()
	defer closeMat(&gray)
	_ = gocv.CvtColor(region, &gray, gocv.ColorBGRToGray)

	bin := gocv.NewMat()
	defer closeMat(&bin)
	// 暗い数字を前景(白)にしたいので BinaryInv + Otsu
	gocv.Threshold(gray, &bin, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)

	tile = gocv.NewMat()
	_ = gocv.CvtColor(bin, &tile, gocv.ColorGrayToBGR)

	labels := gocv.NewMat()
	defer closeMat(&labels)
	stats := gocv.NewMat()
	defer closeMat(&stats)
	centroids := gocv.NewMat()
	defer closeMat(&centroids)
	count := gocv.ConnectedComponentsWithStats(bin, &labels, &stats, &centroids)
	cellArea := float64(bin.Rows() * bin.Cols())

	// 本体と同じ中央判定領域（枠線除去用）
	cw, ch := float64(bin.Cols()), float64(bin.Rows())
	central := image.Rect(
		int(math.Ceil(cw*(dumpCentralThr/2))), int(math.Ceil(ch*(dumpCentralThr/2))),
		int(math.Ceil(cw*(1-dumpCentralThr/2))), int(math.Ceil(ch*(1-dumpCentralThr/2))),
	)
	for i := 1; i < count; i++ {
		area := float64(stats.GetIntAt(i, int(gocv.CC_STAT_AREA)))
		// Otsuは背景(紙)も巨大な1成分になり得るので、セルの大半を占める成分は背景とみなし除外
		if area < cellArea*dumpAreaRatio || area > cellArea*0.9 {
			continue
		}
		// 中央判定: 枠線(セル周縁のみ)は中央に触れないので除外
		mask := gocv.NewMat()
		_ = gocv.InRangeWithScalar(labels, gocv.NewScalar(float64(i), 0, 0, 0), gocv.NewScalar(float64(i), 0, 0, 0), &mask)
		centerCount := gocv.CountNonZero(mask.Region(central))
		closeMat(&mask)
		if centerCount > 0 {
			areaPass++
		}
	}
	return tile, areaPass
}

// analyzeCell は本体 PreprocessAndSplitCellToDigits の判定を再現し、
// 各成分が「面積/上位2個/平均輝度/中央判定」のどこで残り/落ちたかを文字で返す。
func analyzeCell(cell entity.EncodedImage, trim int) string {
	src, err := gocv.IMDecode(cell.Bytes, gocv.IMReadColor)
	if err != nil {
		return "  (decode失敗)"
	}
	defer closeMat(&src)
	region := src.Region(image.Rect(trim, trim, src.Cols()-trim, src.Rows()-trim))
	defer closeMat(&region)

	gray := gocv.NewMat()
	defer closeMat(&gray)
	_ = gocv.CvtColor(region, &gray, gocv.ColorBGRToGray)
	bin := gocv.NewMat()
	defer closeMat(&bin)
	// 本番と同じ: Otsu + BinaryInv
	gocv.Threshold(gray, &bin, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)

	labels := gocv.NewMat()
	defer closeMat(&labels)
	stats := gocv.NewMat()
	defer closeMat(&stats)
	centroids := gocv.NewMat()
	defer closeMat(&centroids)
	count := gocv.ConnectedComponentsWithStats(bin, &labels, &stats, &centroids)
	cellArea := float64(bin.Rows() * bin.Cols())

	type comp struct {
		label, area int
		passArea    bool
	}
	var comps []comp
	for i := 1; i < count; i++ {
		area := int(stats.GetIntAt(i, int(gocv.CC_STAT_AREA)))
		comps = append(comps, comp{label: i, area: area, passArea: float64(area) >= cellArea*dumpAreaRatio})
	}
	// 面積フィルタ通過のうち上位2個（本体が残す候補）
	pass := make([]comp, 0, len(comps))
	for _, cp := range comps {
		if cp.passArea {
			pass = append(pass, cp)
		}
	}
	sort.Slice(pass, func(a, b int) bool { return pass[a].area > pass[b].area })
	top2 := pass
	if len(top2) > 2 {
		top2 = top2[:2]
	}

	cw, ch := float64(bin.Cols()), float64(bin.Rows())
	central := image.Rect(
		int(math.Ceil(cw*(dumpCentralThr/2))), int(math.Ceil(ch*(dumpCentralThr/2))),
		int(math.Ceil(cw*(1-dumpCentralThr/2))), int(math.Ceil(ch*(1-dumpCentralThr/2))),
	)

	lines := fmt.Sprintf("  成分総数=%d 面積通過=%d (cellArea=%.0f, 閾値=%.0f)\n",
		len(comps), len(pass), cellArea, cellArea*dumpAreaRatio)
	survivors := 0
	for _, cp := range top2 {
		mask := gocv.NewMat()
		_ = gocv.InRangeWithScalar(labels, gocv.NewScalar(float64(cp.label), 0, 0, 0), gocv.NewScalar(float64(cp.label), 0, 0, 0), &mask)
		meanMat := gocv.NewMat()
		stdMat := gocv.NewMat()
		_ = gocv.MeanStdDevWithMask(gray, &meanMat, &stdMat, mask)
		mean := meanMat.GetDoubleAt(0, 0) // 参考表示のみ(本番はmean判定を廃止)
		centerCount := gocv.CountNonZero(mask.Region(central))
		verdict := "→ 残る(数字として採用)"
		if centerCount == 0 {
			verdict = "→ 落ちる(中央に触れない=枠線扱い)"
		} else {
			survivors++
		}
		lines += fmt.Sprintf("    [top2] label=%d area=%d(%.1f%%) mean=%.0f center=%d %s\n",
			cp.label, cp.area, 100*float64(cp.area)/cellArea, mean, centerCount, verdict)
		closeMat(&mask)
		closeMat(&meanMat)
		closeMat(&stdMat)
	}
	// 面積は通ったのに上位2個から漏れた成分（=押し出された可能性）
	if len(pass) > 2 {
		for _, cp := range pass[2:] {
			lines += fmt.Sprintf("    [漏れ] label=%d area=%d(%.1f%%) ←面積は通ったが上位2個外で捨てられた\n",
				cp.label, cp.area, 100*float64(cp.area)/cellArea)
		}
	}
	lines += fmt.Sprintf("  生き残り数字=%d\n", survivors)
	return lines
}

// writeMontage はタイル群を元のグリッド形に白線区切りで並べた1枚絵を書き出す。
func writeMontage(t *testing.T, path string, tiles [][]gocv.Mat) {
	t.Helper()
	rows := len(tiles)
	if rows == 0 {
		return
	}
	cols := len(tiles[0])

	tileW, tileH := 0, 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if tiles[r][c].Empty() {
				continue
			}
			if tiles[r][c].Cols() > tileW {
				tileW = tiles[r][c].Cols()
			}
			if tiles[r][c].Rows() > tileH {
				tileH = tiles[r][c].Rows()
			}
		}
	}
	if tileW == 0 || tileH == 0 {
		return
	}

	W := cols*tileW + (cols+1)*dumpSep
	H := rows*tileH + (rows+1)*dumpSep
	white := gocv.NewScalar(255, 255, 255, 0)
	canvas := gocv.NewMatWithSizeFromScalar(white, H, W, gocv.MatTypeCV8UC3)
	defer closeMat(&canvas)

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			tile := tiles[r][c]
			if tile.Empty() {
				continue
			}
			x0 := dumpSep + c*(tileW+dumpSep)
			y0 := dumpSep + r*(tileH+dumpSep)
			roi := canvas.Region(image.Rect(x0, y0, x0+tile.Cols(), y0+tile.Rows()))
			tile.CopyTo(&roi)
			roi.Close()
		}
	}

	buf, err := gocv.IMEncode(gocv.PNGFileExt, canvas)
	if err != nil {
		t.Errorf("montage encode 失敗 %s: %v", path, err)
		return
	}
	defer buf.Close()
	writeDump(t, path, buf.GetBytes())
}

func closeTiles(groups ...[][]gocv.Mat) {
	for _, g := range groups {
		for r := range g {
			for c := range g[r] {
				m := g[r][c]
				closeMat(&m)
			}
		}
	}
}

func writeDump(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Errorf("dump 失敗 %s: %v", path, err)
	}
}

func writeResultText(t *testing.T, path string, expected, got [][]string) {
	t.Helper()
	cell := func(s string) string {
		if s == "" {
			return "·"
		}
		return s
	}
	grid := func(title string, g [][]string) []byte {
		out := []byte(title)
		for _, row := range g {
			for c, s := range row {
				if c > 0 {
					out = append(out, '\t')
				}
				out = append(out, cell(s)...)
			}
			out = append(out, '\n')
		}
		return out
	}
	var b []byte
	b = append(b, "montage_comps.png 凡例: 赤=面積上位2個(本体が残す) 緑=面積通過だが上位2個から漏れた(枠線に負けた数字) 青=面積フィルタ未満 黄=中央判定領域\n\n"...)
	b = append(b, grid("=== EXPECTED ===\n", expected)...)
	b = append(b, grid("\n=== GOT ===\n", got)...)
	b = append(b, "\n=== DIFF (r,c exp -> got) ===\n"...)
	for r := range got {
		for c := range got[r] {
			exp := ""
			if r < len(expected) && c < len(expected[r]) {
				exp = expected[r][c]
			}
			if exp != got[r][c] {
				b = append(b, fmt.Sprintf("r%02d c%02d: %s -> %s\n", r, c, cell(exp), cell(got[r][c]))...)
			}
		}
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Errorf("result.txt 失敗: %v", err)
	}
}
