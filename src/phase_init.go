// filename: phase_init.go
package main

import (
	"math/rand"
)

func (g *Game) PhaseInit(w, h int, rng *rand.Rand, gen *World2Generator) {
	gen.PhaseName = "1. Generating Mask (Type 1)"

	// Phase_Init のロジック本体は InitWorld2Generator() にて実行済み
}