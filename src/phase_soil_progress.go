// filename: phase_soil_progress.go
package main

import (
	"math/rand"
)

func (g *Game) PhaseSoilProgress(w, h int, rng *rand.Rand, gen *World2Generator) {
	// 土壌生成のプログレス処理
	// この関数は現在使用されていない（PhaseSoilCompleteで100%生成する）
	// 将来的に段階的な生成が必要になった場合のために残している
}
