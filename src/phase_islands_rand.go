// filename: world2/phase_islands_rand.go
package main

import (
	"math/rand"
)

func (g *Game) PhaseIslandsRand(w, h int, rng *rand.Rand, gen *World2Generator) {
	gen.PhaseName = "7. Islands (Random)"
	
	for k := 0; k < 5; k++ {
		rx, ry := rng.Intn(w), rng.Intn(h)
		if g.World2.Tiles[rx][ry].Type == W2TileVariableOcean {
			g.World2.Tiles[rx][ry].Type = W2TileSoil
			g.World2.Tiles[rx][ry].Source = SrcIsland
			gen.NewSoils[ry*w+rx] = true
		}
	}
}