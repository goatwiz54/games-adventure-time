// filename: world2/world2_utils.go
package main

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// GenerateMask: タイプごとの形状マスクを生成 (0.0~1.0)
func GenerateMask(w, h, typeID int, rng *rand.Rand) [][]float64 {
	mask := make([][]float64, w)
	for x := range mask {
		mask[x] = make([]float64, h)
	}

	cx, cy := float64(w)/2, float64(h)/2
	maxR := math.Min(cx, cy)

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			val := 0.0

			switch typeID {
			case 1: // クラシック
				val = 1.0
			case 2: // 中央島
				if dist < maxR*0.4 {
					val = 1.0
				}
				if x < int(float64(w)*0.1) || x > int(float64(w)*0.9) {
					val = 0.8
				}
			case 3: // 左大陸
				if x < int(float64(w)*0.6) {
					val = 1.0
				}
			case 4: // 上方大陸
				if y < int(float64(h)*0.5) {
					val = 1.0
				}
			case 5: // 右大陸
				if x > int(float64(w)*0.4) {
					val = 1.0
				}
			case 6: // 回・大陸
				if dist < maxR*0.2 {
					val = 1.0
				}
				if dist > maxR*0.4 && dist < maxR*0.7 {
					val = 0.8
				}
			case 7, 8: // 諸島・連結諸島
				if rng.Float64() > 0.85 {
					val = 1.0
				}
			case 9: // 勾玉
				if (x < int(cx) && y < int(cy) && dist < maxR*0.7) ||
					(x > int(cx) && y > int(cy) && dist < maxR*0.7) {
					val = 1.0
				}
			}
			mask[x][y] = val
		}
	}
	return mask
}

// UpdateMaskImage は FinalMask を Ebiten Image に変換する
func (gen *World2Generator) UpdateMaskImage(w, h int) {
	if gen.MaskImage == nil {
		gen.MaskImage = ebiten.NewImage(w, h)
	}
	gen.MaskImage.Clear()

	pix := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			if gen.FinalMask != nil {
				val := gen.FinalMask[x][y]
				if val > 0 {
					gray := uint8(val * 255)
					pix[idx] = gray
					pix[idx+1] = gray
					pix[idx+2] = gray
					pix[idx+3] = 100
				} else {
					pix[idx+3] = 0
				}
			} else {
				pix[idx+3] = 0
			}
		}
	}
	gen.MaskImage.WritePixels(pix)
}
