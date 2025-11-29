// filename: resources.go
package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func init() {
	TexWhite = ebiten.NewImage(1, 1)
	TexWhite.Fill(color.White)

	// 矢印画像生成
	size := 128
	TexArrow = ebiten.NewImage(size, size)
	orange := color.RGBA{255, 140, 0, 255}
	isInArrow := func(x, y float64) bool {
		if x >= 0 && x <= 64 && y >= 48 && y <= 80 { return true }
		if x >= 64 && x <= 128 {
			topY := 0.75*x - 32; botY := -0.75*x + 160
			if y >= topY && y <= botY { return true }
		}
		return false
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			hits := 0
			for sy := 0.0; sy < 1.0; sy += 0.5 {
				for sx := 0.0; sx < 1.0; sx += 0.5 {
					if isInArrow(float64(x)+sx, float64(y)+sy) { hits++ }
				}
			}
			if hits > 0 {
				c := orange
				c.A = uint8(float64(hits) / 4.0 * 255)
				TexArrow.Set(x, y, c)
			}
		}
	}

	// テクスチャ生成ヘルパー
	genTex := func(baseColor color.RGBA, noiseAmount int, patternType int) *ebiten.Image {
		img := ebiten.NewImage(32, 32)
		pix := make([]byte, 32*32*4)
		// 固定シードのための簡易乱数
		seed := int64(1)
		rnd := func() int {
			seed = seed * 1664525 + 1013904223
			return int((seed >> 16) & 0x7FFF)
		}

		br, bg, bb := int(baseColor.R), int(baseColor.G), int(baseColor.B)
		for i := 0; i < 32*32; i++ {
			x, y := i%32, i/32
			noise := (rnd() % (noiseAmount*2)) - noiseAmount
			
			if patternType == 1 { if rnd()%10 == 0 { noise -= 20 } } // Dirt
			if patternType == 2 { if y%8 == 0 || (y%16 < 8 && x%16 == 0) || (y%16 >= 8 && (x+8)%16 == 0) { noise -= 30 } } // Stone
			if patternType == 0 { if rnd()%8 == 0 { noise += 15 } } // Grass
			
			if x == 0 || y == 0 { noise += 40 } else if x == 31 || y == 31 { noise -= 40 }

			r, g, b := clamp(br+noise), clamp(bg+noise), clamp(bb+noise)
			pix[4*i], pix[4*i+1], pix[4*i+2], pix[4*i+3] = uint8(r), uint8(g), uint8(b), 255
		}
		img.WritePixels(pix)
		return img
	}

	TexGrass = genTex(color.RGBA{85, 125, 70, 255}, 10, 0)
	TexDirt  = genTex(color.RGBA{100, 80, 60, 255}, 15, 1)
	TexStone = genTex(color.RGBA{100, 100, 110, 255}, 10, 2)

	// World 1
	TexW_MountainIcon = ebiten.NewImage(32, 32)
	for y := 10; y < 30; y++ { for x := 4; x < 28; x++ { if y >= int(math.Abs(float64(x-16))*1.2)+10 { TexW_MountainIcon.Set(x, y, color.RGBA{100, 90, 80, 255}) } } }
	TexW_TreeIcon = ebiten.NewImage(16, 16)
	for y := 2; y < 14; y++ { for x := 4; x < 12; x++ { if y >= int(math.Abs(float64(x-8))*1.5)+2 { TexW_TreeIcon.Set(x, y, color.RGBA{30, 80, 40, 255}) } } }
	TexW_CityIcon = ebiten.NewImage(24, 24); TexW_CityIcon.Fill(color.RGBA{200, 50, 50, 255})

	fillImg := func(c color.Color) *ebiten.Image { img := ebiten.NewImage(32, 32); img.Fill(c); return img }
	TexW_Ocean    = fillImg(color.RGBA{20, 60, 120, 255})
	TexW_Plains   = fillImg(color.RGBA{100, 160, 80, 255})
	TexW_Forest   = fillImg(color.RGBA{40, 100, 50, 255})
	TexW_Desert   = fillImg(color.RGBA{200, 180, 100, 255})
	TexW_Mountain = fillImg(color.RGBA{120, 110, 100, 255})

	// World 2
	TexW2_Ocean = ebiten.NewImage(16, 16); TexW2_Ocean.Fill(color.RGBA{20, 60, 150, 255})
	TexW2_Soil = ebiten.NewImage(16, 16); TexW2_Soil.Fill(color.RGBA{139, 69, 19, 255})
}

func clamp(x int) int { if x < 0 { return 0 }; if x > 255 { return 255 }; return x }