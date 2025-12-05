// filename: phase_init.go
package main

import (
	"math/rand"
)

func (g *Game) PhaseInit(w, h int, rng *rand.Rand, gen *World2Generator) {
	// 初期化処理（将来的に他の初期化処理を追加する可能性あり）
}

func (g *Game) PhaseSea(w, h int, rng *rand.Rand, gen *World2Generator) {
	// 可変海で全体を初期化
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			g.World2.Tiles[x][y] = World2Tile{Type: W2TileVariableOcean}
		}
	}
}

func (g *Game) PhaseFixedSea(w, h int, rng *rand.Rand, gen *World2Generator) {
	// 外周3マスを固定海に設定
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if x < 3 || x >= w-3 || y < 3 || y >= h-3 {
				g.World2.Tiles[x][y] = World2Tile{Type: W2TileFixedOcean}
			}
		}
	}
}