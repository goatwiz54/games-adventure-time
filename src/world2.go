// filename: world2.go
package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font/basicfont"
	"github.com/hajimehoshi/ebiten/v2/text"
)

// GenerateMask: タイプごとの形状マスクを生成 (0.0~1.0)
func GenerateMask(w, h, typeID int, rng *rand.Rand) [][]float64 {
	mask := make([][]float64, w)
	for x := range mask {
		mask[x] = make([]float64, h)
	}

	cx, cy := float64(w)/2, float64(h)/2
	maxR := math.Min(cx, cy)

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			val := 0.0

			switch typeID {
			case 1: // クラシック
				val = 0.9
			case 2: // 中央島
				if dist < maxR*0.4 { val = 1.0 }
				if x < int(float64(w)*0.1) || x > int(float64(w)*0.9) { val = 0.8 }
			case 3: // 左大陸
				if x < int(float64(w)*0.6) { val = 1.0 }
			case 4: // 上方大陸
				if y < int(float64(h)*0.5) { val = 1.0 }
			case 5: // 右大陸
				if x > int(float64(w)*0.4) { val = 1.0 }
			case 6: // 回・大陸
				if dist < maxR*0.2 { val = 1.0 }
				if dist > maxR*0.4 && dist < maxR*0.7 { val = 0.8 }
			case 7, 8: // 諸島・連結諸島
				if rng.Float64() > 0.85 { val = 1.0 }
			case 9: // 勾玉
				if (x < int(cx) && y < int(cy) && dist < maxR*0.7) || 
				   (x > int(cx) && y > int(cy) && dist < maxR*0.7) { val = 1.0 }
			}
			mask[x][y] = val
		}
	}
	return mask
}

// --- Generator Logic ---

func (gen *World2Generator) UpdateMaskImage(w, h int) {
	if gen.MaskImage == nil {
		gen.MaskImage = ebiten.NewImage(w, h)
	}
	gen.MaskImage.Clear()
	
	pix := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y*w + x) * 4
			if gen.FinalMask != nil {
				val := gen.FinalMask[x][y]
				if val > 0 {
					gray := uint8(val * 255)
					pix[idx] = gray
					pix[idx+1] = gray
					pix[idx+2] = gray
					pix[idx+3] = 100
				} else {
					pix[idx+3] = 0
				}
			} else {
				pix[idx+3] = 0
			}
		}
	}
	gen.MaskImage.WritePixels(pix)
}

func (g *Game) InitWorld2Generator() {
	// Validation
	if g.W2Width < 40 || g.W2Height < 40 {
		g.WarningMsg = "Size >= 40x40"
		g.WarningTimer = 3.0
		return
	}
	if g.W2Width > 500 || g.W2Height > 500 {
		g.WarningMsg = "Size <= 500x500"
		g.WarningTimer = 3.0
		return
	}

	g.World2 = &WorldMap2{
		Width:   g.W2Width,
		Height:  g.W2Height,
		Tiles:   make([][]World2Tile, g.W2Width),
		OffsetX: float64(g.W2Width*World2TileSize / 2),
		OffsetY: float64(g.W2Height*World2TileSize / 2),
		PinkRects: []Rect{},
	}

	mapPixelW := float64(g.W2Width * World2TileSize)
	mapPixelH := float64(g.W2Height * World2TileSize)
	scaleW := float64(ScreenWidth) / mapPixelW
	scaleH := float64(ScreenHeight) / mapPixelH
	if scaleW < scaleH {
		g.World2.Zoom = scaleW
	} else {
		g.World2.Zoom = scaleH
	}
	g.World2.Zoom *= 0.9

	// シード初期化
	startSeed := rand.Int63()

	gen := &World2Generator{
		CurrentStep: 0,
		IsFinished:  false,
		PhaseName:   "0. Init (Fixed Ocean)",
		History:     []GenSnapshot{},
		Rng:         rand.New(rand.NewSource(startSeed)),
		CurrentSeed: startSeed,
		Config: GenConfig{
			MinPct: g.SoilMin, MaxPct: g.SoilMax, W: g.W2Width, H: g.W2Height,
			TransitDist: g.TransitDist,
			VastOcean: g.VastOceanSize, IslandBound: g.IslandBoundSize,
			// MapTypeMain, MapTypeSub は固定値 1
			MainType: 1, SubType: 1, Ratio: g.MapRatio, 
			Centering: g.EnableCentering,
			CliffInit: g.CliffInitVal, CliffDec: g.CliffDecVal, ShallowDec: g.ShallowDecVal,
			CliffPathLen: g.CliffPathLen,
			ForceSwitch: g.ForceSwitch,
		},
		Multiplier: g.CliffInitVal,
		Excluded:   make(map[int]bool),
		NewSoils:   make(map[int]bool),
	}

	for x := 0; x < g.W2Width; x++ {
		g.World2.Tiles[x] = make([]World2Tile, g.W2Height)
		for y := 0; y < g.W2Height; y++ {
			if x < 3 || x >= g.W2Width-3 || y < 3 || y >= g.W2Height-3 {
				g.World2.Tiles[x][y] = World2Tile{Type: W2TileFixedOcean}
			} else {
				g.World2.Tiles[x][y] = World2Tile{Type: W2TileVariableOcean}
			}
		}
	}

	g.Gen2 = gen
	g.Gen2IsPaused = false
	g.Gen2PausedQuadID = -1 
	g.Gen2PausedIslandCenter = struct{x, y int}{0, 0}
	g.Gen2.UpdateMaskImage(g.W2Width, g.W2Height)
	g.SaveSnapshot()
}

func (g *Game) SaveSnapshot() {
	w, h := g.Gen2.Config.W, g.Gen2.Config.H
	tilesCopy := make([][]World2Tile, w)
	for x := 0; x < w; x++ {
		tilesCopy[x] = make([]World2Tile, h)
		copy(tilesCopy[x], g.World2.Tiles[x])
	}
	
	newSoilsCopy := make(map[int]bool)
	for k, v := range g.Gen2.NewSoils { newSoilsCopy[k] = v }
	
	exCopy := make(map[int]bool)
	for k, v := range g.Gen2.Excluded { exCopy[k] = v }

	pinkCopy := make([]Rect, len(g.World2.PinkRects))
	copy(pinkCopy, g.World2.PinkRects)

	walkersCopy := make([]struct{x, y int}, len(g.Gen2.Walkers))
	copy(walkersCopy, g.Gen2.Walkers)

	g.Gen2.History = append(g.Gen2.History, GenSnapshot{
		Tiles:     tilesCopy,
		PhaseName: g.Gen2.PhaseName,
		StepID:    g.Gen2.CurrentStep,
		NewSoils:  newSoilsCopy,
		Excluded:  exCopy,
		Multiplier: g.Gen2.Multiplier,
		PinkRects: pinkCopy,
		Walkers:   walkersCopy,
		CurrentSoilCount: g.Gen2.CurrentSoilCount,
		CurrentSeed: g.Gen2.CurrentSeed,
		CliffStreak: g.Gen2.CliffStreak,
		ShallowStreak: g.Gen2.ShallowStreak,
	})
}

func (g *Game) UndoStep() {
	if len(g.Gen2.History) > 1 {
		g.Gen2.History = g.Gen2.History[:len(g.Gen2.History)-1]
		last := g.Gen2.History[len(g.Gen2.History)-1]
		
		for x := 0; x < g.Gen2.Config.W; x++ {
			copy(g.World2.Tiles[x], last.Tiles[x])
		}
		
		g.Gen2.PhaseName = last.PhaseName
		g.Gen2.CurrentStep = last.StepID
		g.Gen2.IsFinished = false
		g.Gen2IsPaused = false
		g.Gen2PausedQuadID = -1 
		g.Gen2PausedIslandCenter = struct{x, y int}{0, 0}

		g.Gen2.NewSoils = make(map[int]bool)
		for k, v := range last.NewSoils { g.Gen2.NewSoils[k] = v }
		
		g.Gen2.Excluded = make(map[int]bool)
		for k, v := range last.Excluded { g.Gen2.Excluded[k] = v }
		
		g.World2.PinkRects = make([]Rect, len(last.PinkRects))
		copy(g.World2.PinkRects, last.PinkRects)

		g.Gen2.Walkers = make([]struct{x, y int}, len(last.Walkers))
		copy(g.Gen2.Walkers, last.Walkers)
		
		g.Gen2.CurrentSoilCount = last.CurrentSoilCount
		g.Gen2.Multiplier = last.Multiplier
		g.Gen2.CurrentSeed = last.CurrentSeed
		g.Gen2.Rng.Seed(g.Gen2.CurrentSeed)
		g.Gen2.CliffStreak = last.CliffStreak
		g.Gen2.ShallowStreak = last.ShallowStreak
	}
}

// Phase_IslandsQuad のロジックをポーズ/再実行対応のために分離
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
	
	// 1. ポーズが解除された後の島の生成処理
	if !g.Gen2IsPaused && g.Gen2PausedQuadID != -1 {
		// 前回ポーズした場所で島生成を完了させる
		cx := g.Gen2PausedIslandCenter.x
		cy := g.Gen2PausedIslandCenter.y
		
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

		// 処理完了後、ポーズ情報をクリアし、次のフェーズへ進む
		g.Gen2PausedQuadID = -1
		g.Gen2PausedIslandCenter = struct{x, y int}{0, 0}
		g.Gen2.CurrentStep = Phase_IslandsRand
		gen.PhaseName = "7. Islands (Random)"
		return
	}
	
	// 2. 初期/継続的な広い海探索
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
				g.Gen2IsPaused = true 
				g.Gen2PausedQuadID = qid 
				g.Gen2PausedIslandCenter = struct{x, y int}{cx, cy}
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
		if g.Gen2IsPaused {
			// ポーズがかかったらリターンし、UpdateWorld2からの再実行を待つ
			gen.CurrentStep = Phase_IslandsQuad 
			return 
		}
		
		if ok {
			// ポーズロジックが分離されたため、ここはスキップされ、ポーズ解除後に実行される

			// 見つかった象限をリストから削除して、次の象限のチェックに進む
			quadrants[idx] = quadrants[len(quadrants)-1]
			quadrants = quadrants[:len(quadrants)-1]
			
		} else {
			quadrants[idx] = quadrants[len(quadrants)-1]
			quadrants = quadrants[:len(quadrants)-1]
		}
		g.World2.PinkRects = []Rect{} 
	}
	
	// ループが完了したら次のフェーズへ
	g.Gen2.CurrentStep = Phase_IslandsRand
	gen.PhaseName = "7. Islands (Random)"
}

func (g *Game) NextStep() {
	if g.Gen2.IsFinished { return }
	
	// ポーズがTrueの場合、NextStepの実行はUpdateWorld2()で制御される（ここでreturn）
	if g.Gen2IsPaused { return } 
	
	gen := g.Gen2
	w, h := gen.Config.W, gen.Config.H
	
	gen.Rng.Seed(gen.CurrentSeed)
	rng := gen.Rng

	gen.NewSoils = make(map[int]bool)
	g.World2.PinkRects = []Rect{} // PinkRectsをリセット

	switch gen.CurrentStep {
	case Phase_Init: // 0 -> 1
		gen.CurrentStep = Phase_MaskGen
		gen.PhaseName = "1. Generating Mask (Type 1)"
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
		gen.TargetSoilCount = int(float64(w*h) * float64(targetPct) / 100.0)
		g.LastTargetSoil = targetPct


	case Phase_MaskGen: // 1 -> 2
		gen.CurrentStep = Phase_SoilStart
		gen.PhaseName = "2. Soil: Walkers Start"
		gen.CurrentSoilCount = 0
		
		findSpawn := func() (int, int) {
			cx, cy := w/2, h/2
			return cx + rng.Intn(10)-5, cy + rng.Intn(10)-5
		}

		walkers := 10
		gen.Walkers = make([]struct{ x, y int }, walkers)
		for i := 0; i < walkers; i++ {
			sx, sy := findSpawn()
			gen.Walkers[i].x, gen.Walkers[i].y = sx, sy
		}


	case Phase_SoilStart, 3, 4, 5, 6, 7, 8, 9, 10, Phase_SoilProgressEnd: // 2, 3, ..., 11
		stepIndex := gen.CurrentStep - 1
		milestone := float64(stepIndex) * 0.10
		target := int(float64(gen.TargetSoilCount) * milestone)

		placeSoil := func(x, y int, srcOverride int) bool {
			if x >= 3 && x < w-3 && y >= 3 && y < h-3 {
				if g.World2.Tiles[x][y].Type == W2TileVariableOcean {
					g.World2.Tiles[x][y].Type = W2TileSoil
					if srcOverride != SrcNone {
						g.World2.Tiles[x][y].Source = srcOverride
					} else {
						g.World2.Tiles[x][y].Source = SrcMain 
					}
					gen.NewSoils[y*w+x] = true
					return true
				}
			}
			return false
		}
		
		findSpawn := func() (int, int) {
			cx, cy := w/2, h/2
			return cx + rng.Intn(20)-10, cy + rng.Intn(20)-10
		}

		safety := 0
		if len(gen.Walkers) == 0 {
			gen.Walkers = make([]struct{ x, y int }, 10)
			for i := 0; i < 10; i++ {
				gen.Walkers[i].x, gen.Walkers[i].y = findSpawn()
			}
		}

		for gen.CurrentSoilCount < target && safety < 500000 {
			safety++
			for i := range gen.Walkers {
				if placeSoil(gen.Walkers[i].x, gen.Walkers[i].y, SrcNone) {
					gen.CurrentSoilCount++
				}

				dir := rng.Intn(4)
				bestScore := -1.0
				bestDir := dir
				dxs := []int{0, 1, 0, -1}
				dys := []int{-1, 0, 1, 0}
				for d := 0; d < 4; d++ {
					nx, ny := gen.Walkers[i].x+dxs[d], gen.Walkers[i].y+dys[d]
					score := 0.0
					if nx >= 0 && nx < w && ny >= 0 && ny < h {
						score = gen.FinalMask[nx][ny]
					}
					score += rng.Float64() * 0.5
					if score > bestScore {
						bestScore = score
						bestDir = d
					}
				}

				if bestScore < 0.1 || gen.Walkers[i].x < 3 || gen.Walkers[i].x >= w-3 || gen.Walkers[i].y < 3 || gen.Walkers[i].y >= h-3 {
					nx, ny := findSpawn()
					gen.Walkers[i].x, gen.Walkers[i].y = nx, ny
				} else {
					gen.Walkers[i].x += dxs[bestDir]
					gen.Walkers[i].y += dys[bestDir]
				}
			}
		}
		gen.PhaseName = fmt.Sprintf("3. Soil Progress: %d%%", int(milestone*100))

		// Tectonic Shift at ~30% (Step 4)
		if gen.CurrentStep == 4 { // 4
			shiftX := rng.Intn(w/3*2) - (w / 3)
			shiftY := rng.Intn(h/3*2) - (h / 3)
			tempGrid := make([][]World2Tile, w)
			for x := 0; x < w; x++ {
				tempGrid[x] = make([]World2Tile, h)
				for y := 0; y < h; y++ {
					tempGrid[x][y] = g.World2.Tiles[x][y]
					if tempGrid[x][y].Type == W2TileSoil {
						tempGrid[x][y].Type = W2TileVariableOcean
					}
				}
			}
			colX, colY := false, false
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					if g.World2.Tiles[x][y].Type == W2TileSoil {
						nx := x + shiftX
						if nx < 3 || nx >= w-3 {
							colX = true
						}
					}
				}
			}
			if colX {
				if shiftX > 0 { shiftX -= 5 } else { shiftX += 5 }
			}
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					if g.World2.Tiles[x][y].Type == W2TileSoil {
						ny := y + shiftY
						if ny < 3 || ny >= h-3 {
							colY = true
						}
					}
				}
			}
			if colY {
				if shiftY > 0 { shiftY -= 5 } else { shiftY += 5 }
			}
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					if g.World2.Tiles[x][y].Type == W2TileSoil {
						nx, ny := x + shiftX, y + shiftY
						if nx >= 3 && nx < w-3 && ny >= 3 && ny < h-3 {
							tempGrid[nx][ny] = g.World2.Tiles[x][y]
							gen.NewSoils[ny*w+nx] = true
						}
					}
				}
			}
			g.World2.Tiles = tempGrid
			for i := range gen.Walkers {
				gen.Walkers[i].x += shiftX
				gen.Walkers[i].y += shiftY
				if gen.Walkers[i].x < 3 { gen.Walkers[i].x = 3 }
				if gen.Walkers[i].x >= w-3 { gen.Walkers[i].x = w - 4 }
				if gen.Walkers[i].y < 3 { gen.Walkers[i].y = 3 }
				if gen.Walkers[i].y >= h-3 { gen.Walkers[i].y = h - 4 }
			}
			gen.PhaseName += " + Tectonic"
		}
		gen.CurrentStep++
		if gen.CurrentStep > Phase_SoilProgressEnd {
			gen.CurrentStep = Phase_Bridge
		}

	case Phase_Bridge: // 13 -> 14
		gen.CurrentStep = Phase_Centering
		// Type 1 doesn't use bridges

	case Phase_Centering: // 14 -> 15
		gen.CurrentStep = Phase_IslandsQuad
		if gen.Config.Centering {
			gen.PhaseName = "5. Safe Centering"
			minX, minY, maxX, maxY := w, h, 0, 0
			hasLand := false
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					t := g.World2.Tiles[x][y].Type
					if t == W2TileSoil || t == W2TileTransit {
						if x < minX { minX = x }
						if x > maxX { maxX = x }
						if y < minY { minY = y }
						if y > maxY { maxY = y }
						hasLand = true
					}
				}
			}
			if hasLand {
				contentCx := (minX + maxX) / 2
				contentCy := (minY + maxY) / 2
				mapCx, mapCy := w/2, h/2
				shiftX := mapCx - contentCx
				shiftY := mapCy - contentCy

				if minX+shiftX < 3 { shiftX = 3 - minX }
				if maxX+shiftX > w-4 { shiftX = (w - 4) - maxX }
				if minY+shiftY < 3 { shiftY = 3 - minY }
				if maxY+shiftY > h-4 { shiftY = (h - 4) - maxY }

				newTiles := make([][]World2Tile, w)
				for x := 0; x < w; x++ {
					newTiles[x] = make([]World2Tile, h)
					for y := 0; y < h; y++ {
						newTiles[x][y] = World2Tile{Type: W2TileVariableOcean}
					}
				}
				for x := 0; x < w; x++ {
					for y := 0; y < h; y++ {
						if g.World2.Tiles[x][y].Type == W2TileSoil || g.World2.Tiles[x][y].Type == W2TileTransit {
							nx, ny := x+shiftX, y+shiftY
							if nx >= 0 && nx < w && ny >= 0 && ny < h {
								newTiles[nx][ny] = g.World2.Tiles[x][y]
							}
						}
					}
				}
				for x := 0; x < w; x++ {
					for y := 0; y < h; y++ {
						if x < 3 || x >= w-3 || y < 3 || y >= h-3 {
							newTiles[x][y].Type = W2TileFixedOcean
							newTiles[x][y].Source = SrcNone
						}
					}
				}
				g.World2.Tiles = newTiles
			}
		} else {
			gen.PhaseName = "5. Safe Centering (Skipped)"
		}

	case Phase_IslandsQuad: // 15 -> 16
		// *** 修正: processIslandsQuadStepを呼び出し、ポーズ/再実行ロジックを処理 ***
		g.processIslandsQuadStep(w, h, rng, gen)
		if g.Gen2IsPaused {
			return 
		}
		// processIslandsQuadStep内で次のPhase_IslandsRandへ進む

	case Phase_IslandsRand: // 16 -> 17
		gen.CurrentStep = Phase_Transit
		gen.PhaseName = "7. Islands (Random)"
		for k := 0; k < 5; k++ {
			rx, ry := rng.Intn(w), rng.Intn(h)
			if g.World2.Tiles[rx][ry].Type == W2TileVariableOcean {
				g.World2.Tiles[rx][ry].Type = W2TileSoil
				g.World2.Tiles[rx][ry].Source = SrcIsland
				gen.NewSoils[ry*w+rx] = true
			}
		}

	case Phase_Transit: // 17 -> 18
		gen.CurrentStep = Phase_CliffsShallows
		gen.PhaseName = "8. Transit Islands"
		type Point struct{x,y int}
		islands := []Point{}
		for x := 0; x < w; x += 5 {
			for y := 0; y < h; y += 5 {
				if g.World2.Tiles[x][y].Source == SrcIsland {
					islands = append(islands, Point{x, y})
				}
			}
		}
		calcDist := func(x1, y1, x2, y2 int) float64 {
			return math.Sqrt(math.Pow(float64(x1-x2), 2) + math.Pow(float64(y1-y2), 2))
		}
		findNearestSoil := func(tx, ty int) (int, int) {
			minDist := 999999.0
			nx, ny := tx, ty
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					t := g.World2.Tiles[x][y]
					if (t.Type == W2TileSoil || t.Type == W2TileTransit || t.Type == W2TileCliff) && t.Source != SrcIsland {
						d := calcDist(tx, ty, x, y)
						if d < minDist {
							minDist = d
							nx, ny = x, y
						}
					}
				}
			}
			return nx, ny
		}
		
		// 経由島内部でのランダムウォーク生成 (3x3, 4-7タイル, 周囲3か所に浅瀬)
		genTransitIsland := func(cx, cy int) {
			halfBound := 1 // 3x3 の中心から1マス
			count := 0
			targetCount := 4 + rng.Intn(4) // 4〜7タイル
			
			// 3x3の島を作成
			wx, wy := cx, cy
			for count < targetCount {
				if wx >= cx-halfBound && wx <= cx+halfBound && wy >= cy-halfBound && wy <= cy+halfBound {
					tx, ty := wx, wy
					if g.World2.Tiles[tx][ty].Type == W2TileVariableOcean {
						g.World2.Tiles[tx][ty].Type = W2TileTransit
						g.World2.Tiles[tx][ty].Source = SrcBridge
						gen.NewSoils[ty*w+tx] = true
						count++
					}
				}
				
				dir := rng.Intn(4)
				switch dir {
				case 0: wy--
				case 1: wx++
				case 2: wy++
				case 3: wx--
				}
				// 3x3の範囲内に強制的に留める
				if wx < cx-halfBound { wx = cx - halfBound }
				if wx > cx+halfBound { wx = cx + halfBound }
				if wy < cy-halfBound { wy = cy - halfBound }
				if wy > cy+halfBound { wy = cy + halfBound }
			}
			
			// 浅瀬を3か所生成
			shallowCount := 0
			for i := 0; i < 20 && shallowCount < 3; i++ {
				dx, dy := rng.Intn(5)-2, rng.Intn(5)-2 // 5x5の範囲
				tx, ty := cx+dx, cy+dy
				
				// 経由島の外側 (海) で、かつ固定海でないこと
				if tx >= 3 && tx < w-3 && ty >= 3 && ty < h-3 && g.World2.Tiles[tx][ty].Type == W2TileVariableOcean {
					// 周囲1マスにTransitタイルがあるかチェック
					hasTransitNeighbor := false
					for ndy := -1; ndy <= 1; ndy++ {
						for ndx := -1; ndx <= 1; ndx++ {
							// 境界チェックを追加
							if tx+ndx >= 0 && tx+ndx < w && ty+ndy >= 0 && ty+ndy < h {
								if g.World2.Tiles[tx+ndx][ty+ndy].Type == W2TileTransit {
									hasTransitNeighbor = true
									break
								}
							}
						}
						if hasTransitNeighbor { break }
					}
					
					if hasTransitNeighbor {
						g.World2.Tiles[tx][ty].Type = W2TileShallow
						gen.NewSoils[ty*w+tx] = true
						shallowCount++
					}
				}
			}
		}

		for _, center := range islands {
			sx, sy := findNearestSoil(center.x, center.y)
			if sx == center.x && sy == center.y { continue }
			currX, currY := float64(sx), float64(sy)
			destX, destY := float64(center.x), float64(center.y)
			totalDist := calcDist(int(currX), int(currY), int(destX), int(destY))
			
			if totalDist >= float64(gen.Config.TransitDist) {
				for s := 0; s < 50; s++ {
					vecX := destX - currX
					vecY := destY - currY
					vecLen := math.Sqrt(vecX*vecX + vecY*vecY)
					if vecLen == 0 { break }
					
					step := 5.0 + float64(rng.Intn(4)) 
					
					nextX := currX + (vecX/vecLen)*step
					nextY := currY + (vecY/vecLen)*step
					ix, iy := int(nextX), int(nextY)
					
					distToGoal := calcDist(ix, iy, int(destX), int(destY))
					if distToGoal < 5.0 { break }
					
					nearIslandSoil := false
					for dy := -2; dy <= 2; dy++ {
						for dx := -2; dx <= 2; dx++ {
							tx, ty := ix+dx, iy+dy
							if tx >= 0 && tx < w && ty >= 0 && ty < h {
								if g.World2.Tiles[tx][ty].Type == W2TileSoil {
									if calcDist(tx, ty, sx, sy) > 10 {
										nearIslandSoil = true
									}
								}
							}
						}
					}
					if nearIslandSoil { break }

					// *** 航路のマークアップ ***
					prevX, prevY := int(currX), int(currY)
					markPath := func(x1, y1, x2, y2 int) {
						dx := x2 - x1; dy := y2 - y1
						dist := math.Sqrt(float64(dx*dx + dy*dy))
						if dist == 0 { return }
						steps := int(dist) * 2 // 詳細にマーク
						for i := 0; i <= steps; i++ {
							tx := int(float64(x1) + float64(dx)*float64(i)/float64(steps))
							ty := int(float64(y1) + float64(dy)*float64(i)/float64(steps))
							if tx >= 0 && tx < w && ty >= 0 && ty < h && g.World2.Tiles[tx][ty].Type == W2TileVariableOcean {
								g.World2.Tiles[tx][ty].Source = SrcTransitPath
							}
						}
					}
					markPath(prevX, prevY, ix, iy)
					// ************************

					// *** 経由島の生成 ***
					genTransitIsland(ix, iy)
					
					currX, currY = nextX, nextY
					if calcDist(int(currX), int(currY), int(destX), int(destY)) < float64(gen.Config.TransitDist) {
						break
					}
				}
			}
		}

	case Phase_CliffsShallows: // 18 -> 19
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

	case Phase_LakesFinal: // 19
		gen.CurrentStep++ // 20
		gen.PhaseName = "10. Lakes & Done"
		reached := make([][]bool, w)
		for x := range reached { reached[x] = make([]bool, h) }
		type P struct { x, y int }
		queue := []P{}
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				if g.World2.Tiles[x][y].Type == W2TileFixedOcean {
					reached[x][y] = true
					queue = append(queue, P{x, y})
				}
			}
		}
		dx := []int{0, 1, 0, -1}
		dy := []int{-1, 0, 1, 0}
		for len(queue) > 0 {
			p := queue[0]
			queue = queue[1:]
			for i := 0; i < 4; i++ {
				nx, ny := p.x+dx[i], p.y+dy[i]
				if nx >= 0 && nx < w && ny >= 0 && ny < h {
					t := g.World2.Tiles[nx][ny].Type
					isLand := (t == W2TileSoil || t == W2TileTransit || t == W2TileCliff)
					if !reached[nx][ny] && !isLand {
						reached[nx][ny] = true
						queue = append(queue, P{nx, ny})
					}
				}
			}
		}
		counts := make(map[int]int)
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				t := g.World2.Tiles[x][y].Type
				g.World2.Tiles[x][y].IsLake = false
				if t == W2TileVariableOcean || t == W2TileShallow {
					if !reached[x][y] {
						g.World2.Tiles[x][y].IsLake = true
						counts[-1]++
					} else {
						counts[t]++
					}
				} else {
					counts[t]++
				}
			}
		}
		g.World2.StatsInfo = []string{
			fmt.Sprintf("Phase: %s", gen.PhaseName),
			fmt.Sprintf("Soil:%d Cliff:%d", counts[W2TileSoil], counts[W2TileCliff]),
			fmt.Sprintf("Lake:%d Shlw:%d", counts[-1], counts[W2TileShallow]),
		}
		gen.IsFinished = true
	}
	
	gen.CurrentSeed = gen.Rng.Int63() // Save next seed
	g.SaveSnapshot()
}

// Update
func (g *Game) UpdateWorld2() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.State = StateMenu
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.InitWorld2Generator() // Reset
	}

	// ポーズ中の処理: PgDn (Next) / Enter のみ有効
	if g.Gen2IsPaused {
		if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.Gen2IsPaused = false // ポーズを解除し、次のステップに進む
			g.NextStep()           // NextStepを再度実行
		}
		return nil // 他の操作（移動、ズーム、入力）を無効化
	}
	
	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		g.NextStep()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		g.UndoStep()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && g.InputMode == EditNone {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			for !g.Gen2.IsFinished {
				g.NextStep()
			}
		} else {
			if g.Gen2.IsFinished {
				g.InitWorld2Generator()
			}
			for !g.Gen2.IsFinished {
				g.NextStep()
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		g.World2.ShowGrid = !g.World2.ShowGrid
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		g.InputMode = EditNone
		if mx >= 10 && mx <= 210 {
			if my >= 100 && my <= 130 {
				g.InputMode = EditSoilMin
				g.InputBuffer = fmt.Sprintf("%d", g.SoilMin)
			} else if my >= 140 && my <= 170 {
				g.InputMode = EditSoilMax
				g.InputBuffer = fmt.Sprintf("%d", g.SoilMax)
			} else if my >= 180 && my <= 210 {
				g.InputMode = EditW2Width
				g.InputBuffer = fmt.Sprintf("%d", g.W2Width)
			} else if my >= 220 && my <= 250 {
				g.InputMode = EditW2Height
				g.InputBuffer = fmt.Sprintf("%d", g.W2Height)
			} else if my >= 260 && my <= 290 {
				g.InputMode = EditTransitDist
				g.InputBuffer = fmt.Sprintf("%d", g.TransitDist)
			} else if my >= 300 && my <= 330 {
				g.InputMode = EditMapRatio
				g.InputBuffer = fmt.Sprintf("%d", g.MapRatio)
			} else if my >= 340 && my <= 370 {
				g.EnableCentering = !g.EnableCentering
			} else if my >= 380 && my <= 410 {
				g.InputMode = EditCliffInit
				g.InputBuffer = fmt.Sprintf("%.1f", g.CliffInitVal)
			} else if my >= 420 && my <= 450 {
				g.InputMode = EditCliffDec
				g.InputBuffer = fmt.Sprintf("%.2f", g.CliffDecVal)
			} else if my >= 460 && my <= 490 {
				g.InputMode = EditShallowDec
				g.InputBuffer = fmt.Sprintf("%.2f", g.ShallowDecVal)
			} else if my >= 500 && my <= 530 {
				g.InputMode = EditCliffPath
				g.InputBuffer = fmt.Sprintf("%d", g.CliffPathLen)
			} else if my >= 540 && my <= 570 {
				g.InputMode = EditForceSwitch
				g.InputBuffer = fmt.Sprintf("%d", g.ForceSwitch)
			}
		}
	}

	if g.InputMode != EditNone {
		for k := ebiten.Key0; k <= ebiten.Key9; k++ {
			if inpututil.IsKeyJustPressed(k) {
				g.InputBuffer += string('0' + (k - ebiten.Key0))
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
			g.InputBuffer += "."
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			if len(g.InputBuffer) > 0 {
				g.InputBuffer = g.InputBuffer[:len(g.InputBuffer)-1]
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			valInt, errInt := strconv.Atoi(g.InputBuffer)
			valFloat, errFloat := strconv.ParseFloat(g.InputBuffer, 64)

			if errInt == nil {
				switch g.InputMode {
				case EditSoilMin:
					g.SoilMin = valInt
				case EditSoilMax:
					g.SoilMax = valInt
				case EditW2Width:
					g.W2Width = valInt
				case EditW2Height:
					g.W2Height = valInt
				case EditTransitDist:
					g.TransitDist = valInt
				case EditMapRatio:
					if valInt >= 0 && valInt <= 10 {
						g.MapRatio = valInt
					}
				case EditCliffPath:
					if valInt > 0 {
						g.CliffPathLen = valInt
					}
				case EditForceSwitch:
					g.ForceSwitch = valInt
				}
			}
			if errFloat == nil {
				switch g.InputMode {
				case EditCliffInit:
					g.CliffInitVal = valFloat
				case EditCliffDec:
					g.CliffDecVal = valFloat
				case EditShallowDec:
					g.ShallowDecVal = valFloat
				}
			}
			g.InitWorld2Generator()
			g.InputMode = EditNone
		}
	}

	moveSpd := 10.0 / g.World2.Zoom
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.World2.OffsetX -= moveSpd
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.World2.OffsetX += moveSpd
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.World2.OffsetY -= moveSpd
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.World2.OffsetY += moveSpd
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		cx, cy := ebiten.CursorPosition()
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.IsDragging = true
			g.MouseStartX, g.MouseStartY = cx, cy
		}
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			g.IsDragging = false
		}
		if g.IsDragging {
			dx := float64(g.MouseStartX-cx) / g.World2.Zoom
			dy := float64(g.MouseStartY-cy) / g.World2.Zoom
			g.World2.OffsetX += dx
			g.World2.OffsetY += dy
			g.MouseStartX, g.MouseStartY = cx, cy
		}
	} else {
		g.IsDragging = false
	}

	_, dy := ebiten.Wheel()
	if ebiten.IsKeyPressed(ebiten.KeyControl) && dy != 0 {
		mx, my := ebiten.CursorPosition()
		worldMx := (float64(mx)-ScreenWidth/2)/g.World2.Zoom + g.World2.OffsetX
		worldMy := (float64(my)-ScreenHeight/2)/g.World2.Zoom + g.World2.OffsetY
		if dy > 0 {
			g.World2.Zoom *= 1.1
		} else {
			g.World2.Zoom /= 1.1
		}
		g.World2.OffsetX = worldMx - (float64(mx)-ScreenWidth/2)/g.World2.Zoom
		g.World2.OffsetY = worldMy - (float64(my)-ScreenHeight/2)/g.World2.Zoom
	}
	return nil
}

func (g *Game) DrawWorld2(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	w := g.World2.Width
	h := g.World2.Height

	startX := int(g.World2.OffsetX/float64(World2TileSize)) - int(float64(ScreenWidth)/g.World2.Zoom/float64(World2TileSize))/2 - 2
	startY := int(g.World2.OffsetY/float64(World2TileSize)) - int(float64(ScreenHeight)/g.World2.Zoom/float64(World2TileSize))/2 - 2
	endX := startX + int(float64(ScreenWidth)/g.World2.Zoom/float64(World2TileSize)) + 4
	endY := startY + int(float64(ScreenHeight)/g.World2.Zoom/float64(World2TileSize)) + 4

	for x := startX; x <= endX; x++ {
		for y := startY; y <= endY; y++ {
			if x < 0 || x >= w || y < 0 || y >= h {
				continue
			}
			tile := g.World2.Tiles[x][y]

			sx := (float64(x)*float64(World2TileSize) - g.World2.OffsetX) * g.World2.Zoom + ScreenWidth/2
			sy := (float64(y)*float64(World2TileSize) - g.World2.OffsetY) * g.World2.Zoom + ScreenHeight/2
			size := float64(World2TileSize) * g.World2.Zoom

			var c color.Color
			// --- タイルカラー判定 ---
			switch tile.Type {
			case W2TileSoil:
				if g.Gen2.NewSoils[y*w+x] {
					c = color.RGBA{210, 180, 140, 255}
				} else {
					switch tile.Source {
					case SrcMain:
						c = color.RGBA{180, 100, 80, 255}
					case SrcSub:
						c = color.RGBA{100, 160, 80, 255}
					case SrcMix:
						c = color.RGBA{160, 100, 160, 255}
					case SrcBridge:
						c = color.RGBA{150, 150, 160, 255}
					case SrcIsland:
						c = color.RGBA{230, 190, 100, 255}
					default:
						c = color.RGBA{139, 69, 19, 255}
					}
				}
			case W2TileVariableOcean:
				if tile.IsLake {
					c = color.RGBA{60, 100, 200, 255}
				} else {
					c = color.RGBA{30, 60, 180, 255}
				}
				// 航路の色付け (SrcTransitPath)
				if tile.Source == SrcTransitPath {
					c = color.RGBA{40, 80, 160, 255} // 浅瀬より暗い色
				}
			case W2TileFixedOcean:
				c = color.RGBA{10, 20, 80, 255}
			case W2TileTransit:
				c = color.RGBA{200, 180, 80, 255}
			case W2TileCliff:
				c = color.RGBA{80, 40, 10, 255}
				if g.Gen2.NewSoils[y*w+x] {
					c = color.RGBA{120, 60, 30, 255}
				}
			case W2TileShallow:
				c = color.RGBA{60, 160, 200, 255}
				if g.Gen2.NewSoils[y*w+x] {
					c = color.RGBA{100, 200, 255, 255}
				}
			}
			ebitenutil.DrawRect(screen, sx, sy, size+1, size+1, c)
			// --- タイルカラー判定 終 ---

			if g.World2.ShowGrid || tile.Type == W2TileTransit {
				if g.World2.ShowGrid {
					ebitenutil.DrawRect(screen, sx, sy, size, 1, color.RGBA{255, 255, 255, 50})
					ebitenutil.DrawRect(screen, sx, sy, 1, size, color.RGBA{255, 255, 255, 50})
				}
				if tile.Type == W2TileTransit && size > 10 {
					text.Draw(screen, "経", basicfont.Face7x13, int(sx), int(sy+10), color.Black)
				}
			}
			
			// Gen Mask Imageの描画
			if g.Gen2.CurrentStep <= Phase_SoilStart && g.Gen2.FinalMask != nil {
				val := g.Gen2.FinalMask[x][y]
				if val > 0 {
					gray := uint8(val * 255)
					ebitenutil.DrawRect(screen, sx, sy, size+1, size+1, color.RGBA{gray, gray, gray, 100})
				}
			}
		}
	}
	
	// --- ポーズ中の強調描画 (浅瀬と同じ色) ---
	if g.Gen2IsPaused {
		for _, rect := range g.World2.PinkRects {
			sx := (float64(rect.x)*float64(World2TileSize) - g.World2.OffsetX) * g.World2.Zoom + ScreenWidth/2
			sy := (float64(rect.y)*float64(World2TileSize) - g.World2.OffsetY) * g.World2.Zoom + ScreenHeight/2
			w := float64(rect.w) * g.World2.Zoom
			h := float64(rect.h) * g.World2.Zoom
			
			// 浅瀬の色で半透明の矩形を描画
			ebitenutil.DrawRect(screen, sx, sy, w, h, color.RGBA{60, 160, 200, 100}) 
		}
		
		// ポーズメッセージ
		msg := "Paused: Large Ocean Found (Press [PgDn] or [Enter] to continue)"
		w := len(msg) * 7
		ebitenutil.DrawRect(screen, float64(ScreenWidth/2-w/2-10), float64(ScreenHeight/2-20), float64(w+20), 40, color.RGBA{50, 50, 0, 230})
		text.Draw(screen, msg, basicfont.Face7x13, ScreenWidth/2-w/2, ScreenHeight/2+5, color.White)
	}
	// --- ポーズ中の強調描画 終 ---


	vectorY := 20
	text.Draw(screen, "Phase: "+g.Gen2.PhaseName, basicfont.Face7x13, 220, 20, color.White)
	if g.LastTargetSoil > 0 {
		text.Draw(screen, fmt.Sprintf("Target Soil: %d%%", g.LastTargetSoil), basicfont.Face7x13, 220, 40, color.White)
	}

	for _, s := range g.World2.StatsInfo {
		text.Draw(screen, s, basicfont.Face7x13, 10, vectorY, color.White)
		vectorY += 15
	}

	if g.WarningTimer > 0 {
		msgWidth := len(g.WarningMsg) * 7
		ebitenutil.DrawRect(screen, float64(ScreenWidth/2-msgWidth/2-10), float64(ScreenHeight/2-20), float64(msgWidth+20), 40, color.RGBA{200, 0, 0, 200})
		text.Draw(screen, g.WarningMsg, basicfont.Face7x13, ScreenWidth/2-msgWidth/2, ScreenHeight/2+5, color.White)
	}

	drawInputBox := func(y int, label string, val interface{}, mode int) {
		boxColor := color.RGBA{50, 50, 50, 200}
		if g.InputMode == mode {
			boxColor = color.RGBA{100, 100, 50, 200}
		}
		ebitenutil.DrawRect(screen, 10, float64(y), 200, 30, boxColor)
		
		var txt string
		switch v := val.(type) {
		case int:
			txt = fmt.Sprintf("%s: %d", label, v)
		case float64:
			txt = fmt.Sprintf("%s: %.2f", label, v)
		}

		if g.InputMode == mode {
			txt = fmt.Sprintf("%s: %s_", label, g.InputBuffer)
		}
		text.Draw(screen, txt, basicfont.Face7x13, 20, y+20, color.White)
	}
	
	drawInputBox(100, "Min Soil %", g.SoilMin, EditSoilMin)
	drawInputBox(140, "Max Soil %", g.SoilMax, EditSoilMax)
	drawInputBox(180, "Width", g.W2Width, EditW2Width)
	drawInputBox(220, "Height", g.W2Height, EditW2Height)
	drawInputBox(260, "Transit Dist", g.TransitDist, EditTransitDist)

	drawInputBox(300, "Ratio", g.MapRatio, EditMapRatio)

	cenColor := color.RGBA{50, 0, 0, 200}
	cenText := "OFF"
	if g.EnableCentering {
		cenColor = color.RGBA{0, 100, 0, 200}
		cenText = "ON"
	}
	ebitenutil.DrawRect(screen, 10, 340, 200, 30, cenColor)
	text.Draw(screen, "Centering: "+cenText, basicfont.Face7x13, 20, 360, color.White)

	drawInputBox(380, "Cliff Init", g.CliffInitVal, EditCliffInit)
	drawInputBox(420, "Cliff Dec", g.CliffDecVal, EditCliffDec)
	drawInputBox(460, "Shallow Dec", g.ShallowDecVal, EditShallowDec)
	drawInputBox(500, "Cliff Path", g.CliffPathLen, EditCliffPath)
	drawInputBox(540, "Force Turn", g.ForceSwitch, EditForceSwitch)

	text.Draw(screen, "[PgDn] Next, [PgUp] Back, [Enter] All", basicfont.Face7x13, 10, 670, color.White)
	text.Draw(screen, "[Drag]: Move, [Ctrl+Wheel]: Zoom, [R]: Reset", basicfont.Face7x13, 10, ScreenHeight-20, color.White)
}