// filename: phase_harbor_generation.go
package main

import (
	"math"
	"math/rand"
)

// PhaseHarborGeneration: 港生成フェーズ
// 深海領域が配置された後、各大島から大陸への港を生成する
func (g *Game) PhaseHarborGeneration(w, h int, rng *rand.Rand, gen *World2Generator) {
	// 大島データがない場合はスキップ
	if len(gen.Islands) == 0 {
		gen.PhaseName += " (No islands found)"
		return
	}

	// 現在のレイヤーを取得
	layerIdx := g.World2.CurrentLayerIdx
	layer := &g.World2.Layers[layerIdx]

	// すべての大島に対して処理を実行
	for islandIdx := range gen.Islands {
		island := &gen.Islands[islandIdx]

		// Step 1: 大島側の港を決定（深海の反対側）
		islandPortX, islandPortY := findIslandSeaPort(island, g, w, h)
		island.IslandSeaPortX = islandPortX
		island.IslandSeaPortY = islandPortY

		// Step 2: 螺旋探索で港候補を探す
		harborB, harborC, harborA := findHarborCandidates(
			islandPortX, islandPortY,
			g, w, h, layer,
		)

		if harborB != nil {
			island.HarborCandidateBX = harborB.x
			island.HarborCandidateBY = harborB.y

			// 作業タイルに港候補Bを記録（明るい緑）
			layer.WorkTiles[harborB.x][harborB.y] = WorkTile{
				DataType: WorkDataHarborCandidate,
				Color:    WorkColorBrightGreen,
			}
		}

		if harborC != nil {
			island.HarborCandidateCX = harborC.x
			island.HarborCandidateCY = harborC.y

			// 作業タイルに港候補Cを記録（明るい緑）
			layer.WorkTiles[harborC.x][harborC.y] = WorkTile{
				DataType: WorkDataHarborCandidate,
				Color:    WorkColorBrightGreen,
			}
		}

		if harborA != nil {
			island.HarborCandidateAX = harborA.x
			island.HarborCandidateAY = harborA.y

			// 作業タイルに港候補Aを記録（明るい緑）
			layer.WorkTiles[harborA.x][harborA.y] = WorkTile{
				DataType: WorkDataHarborCandidate,
				Color:    WorkColorBrightGreen,
			}
		}

		// 大島側の港も記録（明るい緑）
		layer.WorkTiles[islandPortX][islandPortY] = WorkTile{
			DataType: WorkDataHarborCandidate,
			Color:    WorkColorBrightGreen,
		}
	}

	_ = rng // 未使用変数警告を回避
}

// findIslandSeaPort: 大島側の港を決定（深海の反対側）
func findIslandSeaPort(island *IslandData, g *Game, w, h int) (int, int) {
	// 大島矩形の中心点から深海を探す
	cpX := island.CenterX
	cpY := island.CenterY

	// 深海の方向を探す（最も近い深海タイルを探す）
	var deepSeaX, deepSeaY int
	minDist := math.MaxFloat64
	foundDeepSea := false

	// 大島周辺を探索
	searchRadius := 50
	for dx := -searchRadius; dx <= searchRadius; dx++ {
		for dy := -searchRadius; dy <= searchRadius; dy++ {
			x := cpX + dx
			y := cpY + dy

			if x < 0 || x >= w || y < 0 || y >= h {
				continue
			}

			tile := g.World2.Tiles[x][y]
			if tile.Type == W2TileDeepSea || tile.Type == W2TileVeryDeepSea {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist < minDist {
					minDist = dist
					deepSeaX = x
					deepSeaY = y
					foundDeepSea = true
				}
			}
		}
	}

	if !foundDeepSea {
		// 深海が見つからない場合は、大島矩形の中心から北側を港とする
		return cpX, cpY - (island.MaxY-island.MinY)/2
	}

	// 深海の反対方向を計算
	dx := cpX - deepSeaX
	dy := cpY - deepSeaY
	length := math.Sqrt(float64(dx*dx + dy*dy))

	if length == 0 {
		return cpX, cpY
	}

	// 正規化
	normDx := float64(dx) / length
	normDy := float64(dy) / length

	// 大島矩形の端まで延長
	halfWidth := (island.MaxX - island.MinX) / 2
	halfHeight := (island.MaxY - island.MinY) / 2
	radius := math.Sqrt(float64(halfWidth*halfWidth + halfHeight*halfHeight))

	portX := cpX + int(normDx*radius)
	portY := cpY + int(normDy*radius)

	// 範囲チェック
	if portX < 0 {
		portX = 0
	}
	if portX >= w {
		portX = w - 1
	}
	if portY < 0 {
		portY = 0
	}
	if portY >= h {
		portY = h - 1
	}

	return portX, portY
}

// HarborPoint: 港候補点
type HarborPoint struct {
	x, y int
}

// findHarborCandidates: 螺旋探索で港候補B, C, Aを探す
func findHarborCandidates(startX, startY int, g *Game, w, h int, layer *Layer) (*HarborPoint, *HarborPoint, *HarborPoint) {
	// アルキメデスの螺旋探索
	// r = b * theta
	// 北(-90度)から開始し、時計回りに探索

	maxDist := math.Sqrt(float64(w*w + h*h))
	pitch := 1.0
	b := pitch / (2 * math.Pi)

	currentTheta := 0.0
	startAngleOffset := -math.Pi / 2.0 // 北(-90度)から開始
	maxTheta := maxDist / b

	var harborB *HarborPoint
	var harborC *HarborPoint
	var harborA *HarborPoint

	foundB := false
	foundC := false

	for currentTheta <= maxTheta {
		r := b * currentTheta
		angle := currentTheta + startAngleOffset

		x := startX + int(r*math.Cos(angle))
		y := startY + int(r*math.Sin(angle))

		if x >= 0 && x < w && y >= 0 && y < h {
			// 螺旋探索点を作業タイルに記録（黄色）
			if layer.WorkTiles[x][y].DataType == WorkDataNone {
				layer.WorkTiles[x][y] = WorkTile{
					DataType: WorkDataSpiral,
					Color:    WorkColorYellow,
				}
			}

			tile := g.World2.Tiles[x][y]

			if !foundB {
				// 港候補B: 最初の大陸の土タイル
				if tile.Type == W2TileSoil && (tile.Source == SrcMain || tile.Source == SrcSub) {
					harborB = &HarborPoint{x: x, y: y}
					foundB = true
				}
			} else if !foundC {
				// 港候補C: B以降の最初の海タイル（C字型の湾の入り口）
				if tile.Type == W2TileVariableOcean || tile.Type == W2TileShallow {
					harborC = &HarborPoint{x: x, y: y}
					foundC = true
				}
			} else {
				// 港候補A: C以降の最初の大陸の土タイル（C字型の湾内）
				if tile.Type == W2TileSoil && (tile.Source == SrcMain || tile.Source == SrcSub) {
					harborA = &HarborPoint{x: x, y: y}
					break
				}
			}
		}

		// 次のステップ
		if r < 1.0 {
			currentTheta += 0.5
		} else {
			currentTheta += 0.5 / r
		}
	}

	return harborB, harborC, harborA
}
