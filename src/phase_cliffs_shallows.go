// filename: world2/phase_cliffs_shallows.go
package main

import (
	"math"
	"math/rand"
)

func (g *Game) PhaseCliffsShallows(w, h int, rng *rand.Rand, gen *World2Generator) {
    _ = math.Abs(0) // math の利用を明示

	gen.CurrentStep = Phase_LakesFinal
	gen.PhaseName = "9. Cliffs & Shallows"
	type P struct { x, y int }
	isCoastal := func(x, y int) bool {
		if g.World2.Tiles[x][y].Type != W2TileSoil && g.World2.Tiles[x][y].Type != W2TileTransit { return false } 
		dxs := []int{0, 1, 0, -1}
		dys := []int{-1, 0, 1, 0}
		for i := 0; i < 4; i++ {
			nx, ny := x+dxs[i], y+dys[i]
			if nx >= 0 && nx < w && ny >= 0 && ny < h && (g.World2.Tiles[nx][ny].Type == W2TileVariableOcean || g.World2.Tiles[nx][ny].Type == W2TileShallow) {
				return true
			}
		}
		return false
	}
	findPath := func(sx, sy, ex, ey int) []P {
		type Node struct { x, y int; path []P }
		queue := []Node{{x: sx, y: sy, path: []P{{sx, sy}}}}
		visited := make(map[int]bool)
		visited[sy*w+sx] = true
		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			if curr.x == ex && curr.y == ey { return curr.path }
			dxs := []int{0, 1, 0, -1}
			dys := []int{-1, 0, 1, 0}
			for i := 0; i < 4; i++ {
				nx, ny := curr.x+dxs[i], curr.y+dys[i]
				if nx >= 0 && nx < w && ny >= 0 && ny < h {
					idx := ny*w + nx
					t := g.World2.Tiles[nx][ny].Type
					if !visited[idx] && (t == W2TileSoil || t == W2TileTransit || t == W2TileCliff) {
						visited[idx] = true
						newPath := make([]P, len(curr.path))
						copy(newPath, curr.path)
						newPath = append(newPath, P{nx, ny})
						queue = append(queue, Node{x: nx, y: ny, path: newPath})
					}
				}
			}
		}
		return nil
	}
	applyShallow := func(cx, cy int) {
		es := []P{}
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 { continue }
				nx, ny := cx+dx, cy+dy
				if nx >= 0 && nx < w && ny >= 0 && ny < h && g.World2.Tiles[nx][ny].Type == W2TileVariableOcean {
					es = append(es, P{nx, ny})
				}
			}
		}
		os := []P{}
		for _, e := range es {
			nCnt := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 { continue }
					nx, ny := e.x+dx, e.y+dy
					if nx >= 0 && nx < w && ny >= 0 && ny < h && g.World2.Tiles[nx][ny].Type == W2TileVariableOcean {
						nCnt++
					}
				}
			}
			if nCnt >= 3 {
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dx == 0 && dy == 0 { continue }
						nx, ny := e.x+dx, e.y+dy
						if nx >= 0 && nx < w && ny >= 0 && ny < h && g.World2.Tiles[nx][ny].Type == W2TileVariableOcean {
							os = append(os, P{nx, ny})
						}
					}
				}
			}
		}
		for _, p := range es { g.World2.Tiles[p.x][p.y].Type = W2TileShallow; gen.NewSoils[p.y*w+p.x] = true }
		for _, p := range os { g.World2.Tiles[p.x][p.y].Type = W2TileShallow; gen.NewSoils[p.y*w+p.x] = true }
	}

	// Safety Loop for Cliff Gen
	safetyLoop := 0
	for gen.Multiplier >= 0 && safetyLoop < 10000 {
		safetyLoop++
		candidatesA := []P{}
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				idx := y*w + x
				if !gen.Excluded[idx] && isCoastal(x, y) {
					candidatesA = append(candidatesA, P{x, y})
				}
			}
		}
		if len(candidatesA) == 0 { break }
		idxA := rng.Intn(len(candidatesA))
		pA := candidatesA[idxA]
		rDist := 1 + rng.Intn(5)
		candidatesB := []P{}
		for x := pA.x - rDist; x <= pA.x+rDist; x++ {
			for y := pA.y - rDist; y <= pA.y+rDist; y++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					idx := y*w + x
					if !gen.Excluded[idx] && isCoastal(x, y) && (x != pA.x || y != pA.y) {
						candidatesB = append(candidatesB, P{x, y})
					}
				}
			}
		}
		if len(candidatesB) == 0 {
			gen.Multiplier -= 0.1
			continue
		}
		pB := candidatesB[rng.Intn(len(candidatesB))]
		pathC := findPath(pA.x, pA.y, pB.x, pB.y)
		
		maxPathLen := gen.Config.CliffPathLen
		if maxPathLen <= 0 { maxPathLen = 5 }

		if len(pathC) == 0 || len(pathC) > maxPathLen {
			gen.Excluded[pA.y*w+pA.x] = true
			gen.Excluded[pB.y*w+pB.x] = true
			continue
		}
		
		forceCliff := false
		forceShallow := false
		if gen.Config.ForceSwitch > 0 {
			if gen.CliffStreak >= gen.Config.ForceSwitch { forceShallow = true } else if gen.ShallowStreak >= gen.Config.ForceSwitch { forceCliff = true }
		}

		isCliff := false
		if forceCliff {
			isCliff = true
		} else if forceShallow {
			isCliff = false
		} else {
			prob := float64(len(pathC)) * gen.Multiplier
			if rng.Float64()*100.0 < prob { isCliff = true }
		}

		if isCliff {
			for _, p := range pathC {
				g.World2.Tiles[p.x][p.y].Type = W2TileCliff
				gen.Excluded[p.y*w+p.x] = true
				gen.NewSoils[p.y*w+p.x] = true
			}
			gen.Multiplier -= gen.Config.CliffDec
			gen.CliffStreak++
			gen.ShallowStreak = 0
		} else {
			for _, p := range pathC {
				applyShallow(p.x, p.y)
				gen.Excluded[p.y*w+p.x] = true
			}
			gen.Multiplier -= gen.Config.ShallowDec
			gen.ShallowStreak++
			gen.CliffStreak = 0
		}
	}
}