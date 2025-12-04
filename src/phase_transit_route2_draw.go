// filename: phase_transit_route2_draw.go
package main

import (
	"math/rand"
)

// 航路2を描画する（旧処理 - スキップ）
func (g *Game) PhaseTransitRoute2Draw(w, h int, rng *rand.Rand, gen *World2Generator) {
	gen.PhaseName = "11. Transit Route 2 (Secondary Draw - Skipped)"
	_ = w
	_ = h
	_ = rng
	// 旧処理はスキップ - 警告を出さないため変数を使用
}
