// filename: world2/phase_centering.go
package main

import (
	"math/rand"
)

func (g *Game) PhaseCentering(w, h int, rng *rand.Rand, gen *World2Generator) {
	if gen.Config.Centering {
		gen.PhaseName = "5. Safe Centering"
		minX, minY, maxX, maxY := w, h, 0, 0
		hasLand := false
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				t := g.World2.Tiles[x][y].Type
				if t == W2TileSoil || t == W2TileTransit {
					if x < minX { minX = x }
					if x > maxX { maxX = x }
					if y < minY { minY = y }
					if y > maxY { maxY = y }
					hasLand = true
				}
			}
		}
		if hasLand {
			contentCx := (minX + maxX) / 2
			contentCy := (minY + maxY) / 2
			mapCx, mapCy := w/2, h/2
			shiftX := mapCx - contentCx
			shiftY := mapCy - contentCy

			if minX+shiftX < 3 { shiftX = 3 - minX }
			if maxX+shiftX > w-4 { shiftX = (w - 4) - maxX }
			if minY+shiftY < 3 { shiftY = 3 - minY }
			if maxY+shiftY > h-4 { shiftY = (h - 4) - maxY }

			newTiles := make([][]World2Tile, w)
			for x := 0; x < w; x++ {
				newTiles[x] = make([]World2Tile, h)
				for y := 0; y < h; y++ {
					newTiles[x][y] = World2Tile{Type: W2TileVariableOcean}
				}
			}
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					if g.World2.Tiles[x][y].Type == W2TileSoil || g.World2.Tiles[x][y].Type == W2TileTransit {
						nx, ny := x+shiftX, y+shiftY
						if nx >= 0 && nx < w && ny >= 0 && ny < h {
							newTiles[nx][ny] = g.World2.Tiles[x][y]
						}
					}
				}
			}
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					if x < 3 || x >= w-3 || y < 3 || y >= h-3 {
						newTiles[x][y].Type = W2TileFixedOcean
						newTiles[x][y].Source = SrcNone
					}
				}
			}
			g.World2.Tiles = newTiles
		}
	} else {
		gen.PhaseName = "5. Safe Centering (Skipped)"
	}
}