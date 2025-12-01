// filename: phase_mask_gen.go
package main

import (
	"math/rand"
	"math"
)

func (g *Game) PhaseMaskGen(w, h int, rng *rand.Rand, gen *World2Generator) {
	gen.PhaseName = "2. Soil: Walkers Start"
	
	gen.MaskMain = GenerateMask(w, h, 1, rng)
	gen.FinalMask = make([][]float64, w)
	for x := 0; x < w; x++ {
		gen.FinalMask[x] = make([]float64, h)
		for y := 0; y < h; y++ {
			gen.FinalMask[x][y] = gen.MaskMain[x][y] 
		}
	}
	gen.UpdateMaskImage(w, h)

	minP, maxP := gen.Config.MinPct, gen.Config.MaxPct
	if minP > maxP { minP, maxP = maxP, minP }
	targetPct := minP
	if maxP > minP { targetPct = minP + rng.Intn(maxP-minP+1) }
	gen.TargetSoilCount = int(math.Round(float64(w*h) * float64(targetPct) / 100.0))
	g.LastTargetSoil = targetPct
}