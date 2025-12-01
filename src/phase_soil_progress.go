// filename: phase_soil_progress.go
package main

import (
	"math"
	"math/rand"
	"fmt"
)

func (g *Game) PhaseSoilProgress(w, h int, rng *rand.Rand, gen *World2Generator) {
	// Soil Progress Block (Phase_SoilStart, 3, 4, 5, 6, 7, 8, 9, 10, Phase_SoilProgressEnd)
	
	stepIndex := gen.CurrentStep - 1
	milestone := float64(stepIndex) * 0.10
	target := int(math.Round(float64(gen.TargetSoilCount) * milestone))

	placeSoil := func(x, y int, srcOverride int) bool {
		if x >= 3 && x < w-3 && y >= 3 && y < h-3 {
			if g.World2.Tiles[x][y].Type == W2TileVariableOcean {
				g.World2.Tiles[x][y].Type = W2TileSoil
				if srcOverride != SrcNone {
					g.World2.Tiles[x][y].Source = srcOverride
				} else {
					g.World2.Tiles[x][y].Source = SrcMain 
				}
				gen.NewSoils[y*w+x] = true
				return true
			}
		}
		return false
	}
	
	findSpawn := func() (int, int) {
		cx, cy := w/2, h/2
		if gen.CurrentStep == Phase_SoilStart { // Phase_SoilStart (2) のみ狭い範囲
			return cx + rng.Intn(10)-5, cy + rng.Intn(10)-5
		}
		return cx + rng.Intn(20)-10, cy + rng.Intn(20)-10
	}

	if gen.CurrentStep == Phase_SoilStart {
		walkers := 10
		gen.Walkers = make([]struct{ x, y int }, walkers)
		for i := 0; i < walkers; i++ {
			sx, sy := findSpawn()
			gen.Walkers[i].x, gen.Walkers[i].y = sx, sy
		}
	}

	safety := 0
	if len(gen.Walkers) == 0 { // エラー対策
		gen.Walkers = make([]struct{ x, y int }, 10)
		for i := 0; i < 10; i++ {
			gen.Walkers[i].x, gen.Walkers[i].y = findSpawn()
		}
	}

	for gen.CurrentSoilCount < target && safety < 500000 {
		safety++
		for i := range gen.Walkers {
			if placeSoil(gen.Walkers[i].x, gen.Walkers[i].y, SrcNone) {
				gen.CurrentSoilCount++
			}

			dir := rng.Intn(4)
			bestScore := -1.0
			bestDir := dir
			dxs := []int{0, 1, 0, -1}
			dys := []int{-1, 0, 1, 0}
			for d := 0; d < 4; d++ {
				nx, ny := gen.Walkers[i].x+dxs[d], gen.Walkers[i].y+dys[d]
				score := 0.0
				if nx >= 0 && nx < w && ny >= 0 && ny < h {
					score = gen.FinalMask[nx][ny]
				}
				score += rng.Float64() * 0.5
				if score > bestScore {
					bestScore = score
					bestDir = d
				}
			}

			if bestScore < 0.1 || gen.Walkers[i].x < 3 || gen.Walkers[i].x >= w-3 || gen.Walkers[i].y < 3 || gen.Walkers[i].y >= h-3 {
				nx, ny := findSpawn()
				gen.Walkers[i].x, gen.Walkers[i].y = nx, ny
			} else {
				gen.Walkers[i].x += dxs[bestDir]
				gen.Walkers[i].y += dys[bestDir]
			}
		}
	}
	gen.PhaseName = fmt.Sprintf("3. Soil Progress: %d%%", int(milestone*100))

	// Tectonic Shift at ~30% (Step 4)
	if gen.CurrentStep == 4 { 
		shiftX := rng.Intn(w/3*2) - (w / 3)
		shiftY := rng.Intn(h/3*2) - (h / 3)
		tempGrid := make([][]World2Tile, w)
		for x := 0; x < w; x++ {
			tempGrid[x] = make([]World2Tile, h)
			for y := 0; y < h; y++ {
				tempGrid[x][y] = g.World2.Tiles[x][y]
				if tempGrid[x][y].Type == W2TileSoil {
					tempGrid[x][y].Type = W2TileVariableOcean
				}
			}
		}
		colX, colY := false, false
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				if g.World2.Tiles[x][y].Type == W2TileSoil {
					nx := x + shiftX
					if nx < 3 || nx >= w-3 {
						colX = true
					}
				}
			}
		}
		if colX {
			if shiftX > 0 { shiftX -= 5 } else { shiftX += 5 }
		}
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				if g.World2.Tiles[x][y].Type == W2TileSoil {
					ny := y + shiftY
					if ny < 3 || ny >= h-3 {
						colY = true
					}
				}
			}
		}
		if colY {
			if shiftY > 0 { shiftY -= 5 } else { shiftY += 5 }
		}
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				if g.World2.Tiles[x][y].Type == W2TileSoil {
					nx, ny := x + shiftX, y + shiftY
					if nx >= 3 && nx < w-3 && ny >= 3 && ny < h-3 {
						tempGrid[nx][ny] = g.World2.Tiles[x][y]
						gen.NewSoils[ny*w+nx] = true
					}
				}
			}
		}
		g.World2.Tiles = tempGrid
		for i := range gen.Walkers {
			gen.Walkers[i].x += shiftX
			gen.Walkers[i].y += shiftY
			if gen.Walkers[i].x < 3 { gen.Walkers[i].x = 3 }
			if gen.Walkers[i].x >= w-3 { gen.Walkers[i].x = w - 4 }
			if gen.Walkers[i].y < 3 { gen.Walkers[i].y = 3 }
			if gen.Walkers[i].y >= h-3 { gen.Walkers[i].y = h - 4 }
		}
		gen.PhaseName += " + Tectonic"
	}
}