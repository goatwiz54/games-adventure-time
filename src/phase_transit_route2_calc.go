// filename: world2/phase_transit_route2_calc.go
package main

import (
	"math/rand"
)

// 航路2の生成可否判定を行う（旧処理 - スキップ）
func (g *Game) PhaseTransitRoute2Calc(w, h int, rng *rand.Rand, gen *World2Generator) {
	gen.PhaseName = "10. Transit Route 2 (Secondary Calc - Skipped)"
	_ = w
	_ = h
	_ = rng
	// 旧処理はスキップ - 警告を出さないため変数を使用
}
