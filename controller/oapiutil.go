package controller

import (
	"encoding/base64"
	"strconv"

	"github.com/atme0627/RelaLogi_go_backend/apperror"
	"github.com/atme0627/RelaLogi_go_backend/entity"
	"github.com/atme0627/RelaLogi_go_backend/transport/rest/oapi"
)

func toEntityQuad(q oapi.Quad) (entity.Quad, error) {
	if len(q) != 4 {
		return entity.Quad{}, apperror.BadRequest("INVALID_HINT_REGION", "ヒント領域は4頂点で指定してください。", nil)
	}

	result := entity.Quad{
		P1: toEntityPoint(q[0]),
		P2: toEntityPoint(q[1]),
		P3: toEntityPoint(q[2]),
		P4: toEntityPoint(q[3]),
	}
	return result, nil
}

func toEntityPoint(p oapi.Point) entity.Point {
	return entity.Point{X: int(p.X), Y: int(p.Y)}
}

func toOapiHintGrid(p entity.Puzzle) (oapi.Grid, oapi.Grid) {
	vSize := oapi.GridSize{
		Cols: p.Size.Width,
		Rows: p.Size.VHintHeight,
	}
	hSize := oapi.GridSize{
		Cols: p.Size.HHintWidth,
		Rows: p.Size.Height,
	}

	return toOapiGrid(p.VHint, vSize), toOapiGrid(p.HHint, hSize)
}

func toOapiGrid(values [][]int, size oapi.GridSize) oapi.Grid {
	result := make([][]string, len(values))
	for i, row := range values {
		result[i] = make([]string, len(row))
		for j, v := range row {
			if v == -1 {
				result[i][j] = ""
			} else {
				result[i][j] = strconv.Itoa(v)
			}
		}
	}
	return oapi.Grid{Size: size, Values: result}
}

func toDataURI(img entity.EncodedImage) string {
	return "data:" + img.MimeTypes + ";base64," + base64.StdEncoding.EncodeToString(img.Bytes)
}
