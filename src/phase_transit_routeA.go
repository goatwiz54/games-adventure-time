// filename: phase_transit_routeA.go
package main

import (
	"math"
	"math/rand"
)

// 航路探索A - 新しい航路生成方式
func (g *Game) PhaseTransitRouteA(w, h int, rng *rand.Rand, gen *World2Generator) {
	gen.PhaseName = "12. Transit Route A (New Method)"

	// 大島データがない場合はスキップ
	if len(gen.Islands) == 0 {
		gen.PhaseName += " (No islands found)"
		return
	}

	// PinkRectsをクリアして新規描画
	g.World2.PinkRects = []Rect{}

	// すべての大島に対して処理を実行
	for _, island := range gen.Islands {
		// 島の中心点CP (Center Point)は既に計算済み
		cpX := island.CenterX
		cpY := island.CenterY

		// CPを起点に円を描き、大陸の最初に一致した「土」を仮の港SP (Sea Port)とする
		// 北東南西の順番で判定
		type PortInfo struct {
			x, y int
			dist float64
		}

		var seaPort *PortInfo

		// 北東南西の方向を定義
		directions := []struct {
			name  string
			check func(angle float64) bool
		}{
			{"North", func(angle float64) bool { return angle >= 225 && angle < 315 }}, // 北 (上方向)
			{"East", func(angle float64) bool { return angle >= 315 || angle < 45 }},   // 東 (右方向)
			{"South", func(angle float64) bool { return angle >= 45 && angle < 135 }},  // 南 (下方向)
			{"West", func(angle float64) bool { return angle >= 135 && angle < 225 }},  // 西 (左方向)
		}

		// 円を拡大しながら大陸の土を探す（北東南西の順番）
		maxRadius := int(math.Sqrt(float64(w*w + h*h)))
		for _, dir := range directions {
			found := false
			for radius := 1; radius <= maxRadius && !found; radius++ {
				// 円周上の点を1度刻みでチェック
				for angle := 0.0; angle < 360.0; angle += 1.0 {
					if !dir.check(angle) {
						continue // 指定方向でない場合はスキップ
					}

					rad := angle * math.Pi / 180.0
					x := cpX + int(float64(radius)*math.Cos(rad))
					y := cpY + int(float64(radius)*math.Sin(rad))

					if x >= 0 && x < w && y >= 0 && y < h {
						tile := g.World2.Tiles[x][y]
						// 大陸の土のみ（SrcMain または SrcSub）を探す
						// ランダム5点島、経由島、大島は除外
						if tile.Type == W2TileSoil && (tile.Source == SrcMain || tile.Source == SrcSub) {
							dist := math.Sqrt(float64((x-cpX)*(x-cpX) + (y-cpY)*(y-cpY)))
							seaPort = &PortInfo{x: x, y: y, dist: dist}
							found = true
							break
						}
					}
				}
			}
			if found {
				break // 北東南西の順で最初に見つかった方向でSPが確定
			}
		}

		if seaPort == nil {
			// この島には大陸への港が見つからなかった（スキップ）
			continue
		}

		// CP-SPの直線距離をFT (First Distance)とする
		ft := seaPort.dist

		// CP-SPの辺を基準に、FT距離の2倍の長さの長方形を描く
		// rect1: CP-SP方向
		// rect2: 反対方向

		// CP-SPのベクトルを計算
		dx := seaPort.x - cpX
		dy := seaPort.y - cpY

		// ベクトルの長さで正規化
		length := math.Sqrt(float64(dx*dx + dy*dy))
		if length == 0 {
			continue
		}

		normDx := float64(dx) / length
		normDy := float64(dy) / length

		// 長方形の長さ（FT * 2、ただし固定海まで）
		rectLength := int(ft * 2)

		// 長方形の幅（仮に10マスとする）
		rectWidth := 10

		// 垂直方向のベクトル（幅方向）
		perpDx := -normDy
		perpDy := normDx

		// rect1: CP-SP方向に長方形を描画
		for i := 0; i < rectLength; i++ {
			centerX := cpX + int(normDx*float64(i))
			centerY := cpY + int(normDy*float64(i))

			// 固定海に達したら停止
			if centerX < 0 || centerX >= w || centerY < 0 || centerY >= h {
				break
			}
			if g.World2.Tiles[centerX][centerY].Type == W2TileFixedOcean {
				break
			}

			// 幅方向に広げる
			for j := -rectWidth / 2; j <= rectWidth/2; j++ {
				wx := centerX + int(perpDx*float64(j))
				wy := centerY + int(perpDy*float64(j))

				if wx >= 0 && wx < w && wy >= 0 && wy < h {
					if g.World2.Tiles[wx][wy].Type != W2TileFixedOcean {
						// 赤の半透明で描画するため、PinkRectsに追加
						g.World2.PinkRects = append(g.World2.PinkRects, Rect{
							x: wx,
							y: wy,
							w: World2TileSize,
							h: World2TileSize,
						})
					}
				}
			}
		}

		// rect2: 反対方向に長方形を描画
		for i := 1; i < rectLength; i++ { // i=1から開始（中心は既に描画済み）
			centerX := cpX - int(normDx*float64(i))
			centerY := cpY - int(normDy*float64(i))

			// 固定海に達したら停止
			if centerX < 0 || centerX >= w || centerY < 0 || centerY >= h {
				break
			}
			if g.World2.Tiles[centerX][centerY].Type == W2TileFixedOcean {
				break
			}

			// 幅方向に広げる
			for j := -rectWidth / 2; j <= rectWidth/2; j++ {
				wx := centerX + int(perpDx*float64(j))
				wy := centerY + int(perpDy*float64(j))

				if wx >= 0 && wx < w && wy >= 0 && wy < h {
					if g.World2.Tiles[wx][wy].Type != W2TileFixedOcean {
						// 赤の半透明で描画するため、PinkRectsに追加
						g.World2.PinkRects = append(g.World2.PinkRects, Rect{
							x: wx,
							y: wy,
							w: World2TileSize,
							h: World2TileSize,
						})
					}
				}
			}
		}
	}

	_ = rng // 未使用変数警告を回避
}
