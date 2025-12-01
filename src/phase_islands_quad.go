// filename: world2/phase_islands_quad.go
package main

import (
	"math"
	"math/rand"
)

// Phase_IslandsQuad のロジックをポーズ/再実行対応のために分離 (ポーズ処理は完全に削除)
func (g *Game) processIslandsQuadStep(w, h int, rng *rand.Rand, gen *World2Generator) {
	vastSize := gen.Config.VastOcean
	boundSize := gen.Config.IslandBound
	
	placeSoil := func(x, y int, src int) {
		if x >= 3 && x < w-3 && y >= 3 && y < h-3 {
			g.World2.Tiles[x][y].Type = W2TileSoil
			g.World2.Tiles[x][y].Source = src
			gen.NewSoils[y*w+x] = true
		}
	}
	
	// 2. 広い海探索ロジック（ポーズ処理は削除）
	quadrants := []int{0, 1, 2, 3}
	created := 0
	
	findVast := func(qid int) (int, int, bool) {
		var sx, sy, ex, ey int
		halfW, halfH := w/2, h/2
		switch qid {
		case 0: sx, sy, ex, ey = 3, 3, halfW, halfH
		case 1: sx, sy, ex, ey = halfW, 3, w-3, halfH
		case 2: sx, sy, ex, ey = 3, halfH, halfW, h-3
		case 3: sx, sy, ex, ey = halfW, halfH, w-3, h-3
		}
		for try := 0; try < 30; try++ {
			cx := sx + rng.Intn(ex-sx)
			cy := sy + rng.Intn(ey-sy)
			halfVast := vastSize / 2
			isVast := true
			
			// 描画用の矩形は設定するが、ポーズはしない
			g.World2.PinkRects = []Rect{} 
			
			vastRect := Rect{x: cx - halfVast, y: cy - halfVast, w: vastSize, h: vastSize}
			g.World2.PinkRects = append(g.World2.PinkRects, vastRect)
			
			for dy := -halfVast; dy <= halfVast; dy++ {
				for dx := -halfVast; dx <= halfVast; dx++ {
					tx, ty := cx+dx, cy+dy
					if tx < 3 || tx >= w-3 || ty < 3 || ty >= h-3 {
						isVast = false
						break
					}
					if g.World2.Tiles[tx][ty].Type != W2TileVariableOcean {
						isVast = false
						break
					}
				}
				if !isVast { break }
			}
			
			if isVast {
				// ポーズ処理はすべて削除し、見つけたら即座に島の生成に利用
				return cx, cy, true 
			} else {
				g.World2.PinkRects = []Rect{}
			}
		}
		return 0, 0, false
	}
	
	// 3. 島の探索と生成ループ
	for created < 4 && len(quadrants) > 0 { 
		idx := rng.Intn(len(quadrants))
		qid := quadrants[idx]
		
		cx, cy, ok := findVast(qid)
		
		if ok {
			// 島を生成するロジック
			area := boundSize * boundSize
			ratio := 0.3 + rng.Float64()*0.4
			tgt := int(math.Round(float64(area) * ratio))
			cnt := 0
			wx, wy := cx, cy
			halfBound := boundSize / 2
			for cnt < tgt {
				placeSoil(wx, wy, SrcIsland)
				cnt++
				dir := rng.Intn(4)
				switch dir {
				case 0: wy--
				case 1: wx++
				case 2: wy++
				case 3: wx--
				}
				if wx < cx-halfBound { wx = cx - halfBound }
				if wx > cx+halfBound { wx = cx + halfBound }
				if wy < cy-halfBound { wy = cy - halfBound }
				if wy > cy+halfBound { wy = cy + halfBound }
			}
			
			// 見つかった象限をリストから削除して、次の象限のチェックに進む
			quadrants[idx] = quadrants[len(quadrants)-1]
			quadrants = quadrants[:len(quadrants)-1]
			created++
			
		} else {
			quadrants[idx] = quadrants[len(quadrants)-1]
			quadrants = quadrants[:len(quadrants)-1]
		}
		g.World2.PinkRects = []Rect{} 
	}

	// ループが完了したら次のフェーズへ
	gen.PhaseName = "6. Islands (Quad)"
}

func (g *Game) PhaseIslandsQuad(w, h int, rng *rand.Rand, gen *World2Generator) {
	g.processIslandsQuadStep(w, h, rng, gen)
}