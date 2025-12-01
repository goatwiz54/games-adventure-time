// filename: phase_bridge.go
package main

import (
	"math/rand"
)

func (g *Game) PhaseBridge(w, h int, rng *rand.Rand, gen *World2Generator) {
	// Type 1 doesn't use bridges
	gen.PhaseName = "4. Bridge (Skipped)"
}