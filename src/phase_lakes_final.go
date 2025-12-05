// filename: world2/phase_lakes_final.go
package main

import (
	"fmt"
	"math/rand"
)

func (g *Game) PhaseLakesFinal(w, h int, rng *rand.Rand, gen *World2Generator) {
	reached := make([][]bool, w)
	for x := range reached {
		reached[x] = make([]bool, h)
	}
	type P struct{ x, y int }
	queue := []P{}
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if g.World2.Tiles[x][y].Type == W2TileFixedOcean {
				reached[x][y] = true
				queue = append(queue, P{x, y})
			}
		}
	}
	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for i := 0; i < 4; i++ {
			nx, ny := p.x+dx[i], p.y+dy[i]
			if nx >= 0 && nx < w && ny >= 0 && ny < h {
				t := g.World2.Tiles[nx][ny].Type
				isLand := (t == W2TileSoil || t == W2TileTransit || t == W2TileCliff)
				if !reached[nx][ny] && !isLand {
					reached[nx][ny] = true
					queue = append(queue, P{nx, ny})
				}
			}
		}
	}
	counts := make(map[int]int)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			t := g.World2.Tiles[x][y].Type
			g.World2.Tiles[x][y].IsLake = false
			if t == W2TileVariableOcean || t == W2TileShallow {
				if !reached[x][y] {
					g.World2.Tiles[x][y].IsLake = true
					counts[-1]++
				} else {
					counts[t]++
				}
			} else {
				counts[t]++
			}
		}
	}
	g.World2.StatsInfo = []string{
		fmt.Sprintf("Phase: %s", gen.PhaseName),
		fmt.Sprintf("Soil:%d Cliff:%d", counts[W2TileSoil], counts[W2TileCliff]),
		fmt.Sprintf("Lake:%d Shlw:%d", counts[-1], counts[W2TileShallow]),
	}
	gen.IsFinished = true
}
