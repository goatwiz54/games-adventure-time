// filename: phase_tectonic.go
package main

import (
	"fmt"
	"math/rand"
)

// PhaseTectonicShift: 地殻変動（土壌全体をランダムにシフト）
func (g *Game) PhaseTectonicShift(w, h int, rng *rand.Rand, gen *World2Generator) {
	// ランダムなシフト量を決定（±w/3、±h/3）
	shiftX := rng.Intn(w/3*2) - (w / 3)
	shiftY := rng.Intn(h/3*2) - (h / 3)

	// 一時グリッドを作成（全体を可変海で初期化）
	tempGrid := make([][]World2Tile, w)
	for x := 0; x < w; x++ {
		tempGrid[x] = make([]World2Tile, h)
		for y := 0; y < h; y++ {
			tempGrid[x][y] = g.World2.Tiles[x][y]
			// 土壌を一旦海に戻す
			if tempGrid[x][y].Type == W2TileSoil {
				tempGrid[x][y].Type = W2TileVariableOcean
			}
		}
	}

	// X方向の境界衝突チェック
	colX := false
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if g.World2.Tiles[x][y].Type == W2TileSoil {
				nx := x + shiftX
				// 固定海の外周3マスに達するかチェック
				if nx < 3 || nx >= w-3 {
					colX = true
					break
				}
			}
		}
		if colX {
			break
		}
	}

	// X方向の衝突があれば補正
	if colX {
		if shiftX > 0 {
			shiftX -= 5
		} else {
			shiftX += 5
		}
	}

	// Y方向の境界衝突チェック
	colY := false
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if g.World2.Tiles[x][y].Type == W2TileSoil {
				ny := y + shiftY
				// 固定海の外周3マスに達するかチェック
				if ny < 3 || ny >= h-3 {
					colY = true
					break
				}
			}
		}
		if colY {
			break
		}
	}

	// Y方向の衝突があれば補正
	if colY {
		if shiftY > 0 {
			shiftY -= 5
		} else {
			shiftY += 5
		}
	}

	// 土壌をシフトして配置
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if g.World2.Tiles[x][y].Type == W2TileSoil {
				nx, ny := x+shiftX, y+shiftY
				// 固定海の範囲内に収まる場合のみ配置
				if nx >= 3 && nx < w-3 && ny >= 3 && ny < h-3 {
					tempGrid[nx][ny] = g.World2.Tiles[x][y]
					gen.NewSoils[ny*w+nx] = true
				}
			}
		}
	}

	// グリッドを更新
	g.World2.Tiles = tempGrid

	// Walkerもシフト
	for i := range gen.Walkers {
		gen.Walkers[i].x += shiftX
		gen.Walkers[i].y += shiftY

		// Walkerが境界外に出た場合は補正
		if gen.Walkers[i].x < 3 {
			gen.Walkers[i].x = 3
		}
		if gen.Walkers[i].x >= w-3 {
			gen.Walkers[i].x = w - 4
		}
		if gen.Walkers[i].y < 3 {
			gen.Walkers[i].y = 3
		}
		if gen.Walkers[i].y >= h-3 {
			gen.Walkers[i].y = h - 4
		}
	}

	// フェーズ名を更新
	gen.PhaseName = fmt.Sprintf("Tectonic Shift (offset: %d, %d)", shiftX, shiftY)
}
