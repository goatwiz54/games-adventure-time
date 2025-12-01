// filename: world2/phase_transit_route2_calc.go
package main

import (
	"math/rand"
	"math"
	"fmt"
)

// 航路2の生成可否判定を行う
func (g *Game) PhaseTransitRoute2Calc(w, h int, rng *rand.Rand, gen *World2Generator) {
    _ = math.Sqrt(1.0) // math の利用を明示

	gen.PhaseName = "10. Transit Route 2 (Secondary Calc)"

	// 1. 航路1の長さ判定 (25マス以上)
	if g.TotalRoute1Dist < 25.0 {
		gen.PhaseName += " (Skipped: Route1 too short)"
		return
	}
	
	// 2. 島の数を数える
	islandCount := 0
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if g.World2.Tiles[x][y].Source == SrcIsland {
				islandCount++
			}
		}
	}
	
	// 3. 確率判定
	isSecondaryRouteNeeded := false
	r := rng.Intn(100)
	
	switch islandCount {
	case 1:
		if r < 30 { isSecondaryRouteNeeded = true } // 30%
	case 2:
		if r < 70 { isSecondaryRouteNeeded = true } // 70%
	case 3:
		if r < 10 { isSecondaryRouteNeeded = true } // 10%
	case 4:
		if r < 5 { isSecondaryRouteNeeded = true }  // 5%
	}
	
	if isSecondaryRouteNeeded {
		// 描画フェーズに進む
		gen.PhaseName += " -> DRAW"
	} else {
		// スキップして次のフェーズへ
		gen.CurrentStep = Phase_CliffsShallows
	}
	
	fmt.Println("Secondary Route Needed:", isSecondaryRouteNeeded)
}