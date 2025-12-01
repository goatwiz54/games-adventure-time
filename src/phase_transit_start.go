// filename: world2/phase_transit_start.go
package main

import (
	"math"
	"math/rand"
)

func (g *Game) PhaseTransitStart(w, h int, rng *rand.Rand, gen *World2Generator) {
	gen.PhaseName = "8. Transit Islands (Route 1)"
	
	// *** 修正: markPath ローカル関数をここに定義 ***
	markPath := func(x1, y1, x2, y2 int, w, h int) {
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
	// **********************************************
	
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

		// 探索範囲を制限（中心から一定範囲のみ）
		searchRadius := 100
		startX := tx - searchRadius
		if startX < 0 { startX = 0 }
		endX := tx + searchRadius
		if endX >= w { endX = w - 1 }
		startY := ty - searchRadius
		if startY < 0 { startY = 0 }
		endY := ty + searchRadius
		if endY >= h { endY = h - 1 }

		// 5マス刻みでスキャン（高速化）
		for x := startX; x <= endX; x += 5 {
			for y := startY; y <= endY; y += 5 {
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
	
	// 海の広さを判定（周囲の大陸密度をチェック）
	checkSeaWidth := func(x, y int) string {
		radius := 20
		landCount := 0
		totalChecked := 0

		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				tx, ty := x+dx, y+dy
				if tx >= 0 && tx < w && ty >= 0 && ty < h {
					totalChecked++
					t := g.World2.Tiles[tx][ty]
					if (t.Type == W2TileSoil || t.Type == W2TileCliff) && t.Source != SrcIsland {
						landCount++
					}
				}
			}
		}

		landRatio := float64(landCount) / float64(totalChecked)
		if landRatio < 0.2 {
			return "wide"  // 広い海
		}
		return "narrow"    // 狭い海
	}

	// 色付き航路描画
	markPathWithColor := func(x1, y1, x2, y2 int, w, h int, sourceType int) {
		dx := x2 - x1
		dy := y2 - y1
		dist := math.Sqrt(float64(dx*dx + dy*dy))
		if dist == 0 { return }
		steps := int(dist) * 2

		for i := 0; i <= steps; i++ {
			tx := int(float64(x1) + float64(dx)*float64(i)/float64(steps))
			ty := int(float64(y1) + float64(dy)*float64(i)/float64(steps))
			if tx >= 0 && tx < w && ty >= 0 && ty < h {
				tile := g.World2.Tiles[tx][ty]
				// 陸地に当たったら描画を停止
				if tile.Type == W2TileSoil || tile.Type == W2TileCliff {
					break
				}
				if tile.Type == W2TileVariableOcean {
					g.World2.Tiles[tx][ty].Source = sourceType
				}
			}
		}
	}

	// A航路: 控えめな曲線（±5マス）
	markCurvedPathA := func(x1, y1, x2, y2 int, w, h int) {
		dx := x2 - x1
		dy := y2 - y1
		dist := math.Sqrt(float64(dx*dx + dy*dy))
		if dist < 10 {
			// 短い距離は直線
			markPathWithColor(x1, y1, x2, y2, w, h, SrcTransitPath)
			return
		}

		// 中間点を計算
		midX := (x1 + x2) / 2
		midY := (y1 + y2) / 2

		// 垂直方向にずらす（±5マス）
		perpDx := -float64(dy) / dist
		perpDy := float64(dx) / dist
		offset := (rng.Float64()*10 - 5) // -5 ~ +5

		ctrlX := float64(midX) + perpDx*offset
		ctrlY := float64(midY) + perpDy*offset

		// 2次ベジェ曲線で描画
		steps := int(dist * 2)
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(steps)
			px := (1-t)*(1-t)*float64(x1) + 2*(1-t)*t*ctrlX + t*t*float64(x2)
			py := (1-t)*(1-t)*float64(y1) + 2*(1-t)*t*ctrlY + t*t*float64(y2)

			ix, iy := int(px), int(py)
			if ix >= 0 && ix < w && iy >= 0 && iy < h {
				if g.World2.Tiles[ix][iy].Type == W2TileVariableOcean {
					g.World2.Tiles[ix][iy].Source = SrcTransitPath
				}
			}
		}
	}

	// B航路: 弧の航路（±10マス）
	markArcPathB := func(x1, y1, x2, y2 int, w, h int) {
		dx := x2 - x1
		dy := y2 - y1
		dist := math.Sqrt(float64(dx*dx + dy*dy))

		// 中間点を計算
		midX := (x1 + x2) / 2
		midY := (y1 + y2) / 2

		// 垂直方向にずらす（±10マス）
		perpDx := -float64(dy) / dist
		perpDy := float64(dx) / dist
		arcHeight := 5.0 + rng.Float64()*10.0 // 5~15マス
		direction := 1.0
		if rng.Float64() < 0.5 { direction = -1.0 }

		ctrlX := float64(midX) + perpDx*arcHeight*direction
		ctrlY := float64(midY) + perpDy*arcHeight*direction

		// 制御点が陸地上にあるかチェック
		ctrlIX, ctrlIY := int(ctrlX), int(ctrlY)
		if ctrlIX >= 0 && ctrlIX < w && ctrlIY >= 0 && ctrlIY < h {
			ctrlTile := g.World2.Tiles[ctrlIX][ctrlIY]
			if ctrlTile.Type == W2TileSoil || ctrlTile.Type == W2TileCliff {
				// 制御点が陸地の場合、直線にフォールバック
				markPathWithColor(x1, y1, x2, y2, w, h, SrcBRoutePath)
				return
			}
		}

		// 2次ベジェ曲線で描画
		steps := int(dist * 2)
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(steps)
			px := (1-t)*(1-t)*float64(x1) + 2*(1-t)*t*ctrlX + t*t*float64(x2)
			py := (1-t)*(1-t)*float64(y1) + 2*(1-t)*t*ctrlY + t*t*float64(y2)

			ix, iy := int(px), int(py)
			if ix >= 0 && ix < w && iy >= 0 && iy < h {
				tile := g.World2.Tiles[ix][iy]
				// 陸地に当たったら描画を停止
				if tile.Type == W2TileSoil || tile.Type == W2TileCliff {
					break
				}
				if tile.Type == W2TileVariableOcean {
					g.World2.Tiles[ix][iy].Source = SrcBRoutePath
				}
			}
		}
	}

	// B航路: ジグザグ航路（±10マス、小さい経由島付き）
	markZigzagPathB := func(x1, y1, x2, y2 int, w, h int) {
		dx := x2 - x1
		dy := y2 - y1
		dist := math.Sqrt(float64(dx*dx + dy*dy))
		segments := int(dist / 12) // 12マスごとにジグザグ
		if segments < 2 { segments = 2 }

		currX, currY := float64(x1), float64(y1)
		for i := 1; i <= segments; i++ {
			t := float64(i) / float64(segments)

			// 基本の直線上の点
			baseX := float64(x1) + float64(dx)*t
			baseY := float64(y1) + float64(dy)*t

			// ランダムに左右にずらす（±10マス）
			perpDx := -float64(dy) / dist
			perpDy := float64(dx) / dist
			offset := (rng.Float64()*20 - 10) // -10 ~ +10

			nextX := baseX + perpDx*offset
			nextY := baseY + perpDy*offset

			// ジグザグ点が陸地上にあるかチェック
			nextIX, nextIY := int(nextX), int(nextY)
			if nextIX >= 0 && nextIX < w && nextIY >= 0 && nextIY < h {
				nextTile := g.World2.Tiles[nextIX][nextIY]
				if nextTile.Type == W2TileSoil || nextTile.Type == W2TileCliff {
					// 陸地の場合、オフセットを減らす（直線に近づける）
					offset = offset * 0.3
					nextX = baseX + perpDx*offset
					nextY = baseY + perpDy*offset
				}
			}

			// 航路を描画（暗緑）
			markPathWithColor(int(currX), int(currY), int(nextX), int(nextY), w, h, SrcBRoutePath)

			// 小さい経由島を配置（1x1、緑）
			ix, iy := int(nextX), int(nextY)
			if ix >= 3 && ix < w-3 && iy >= 3 && iy < h-3 {
				tile := g.World2.Tiles[ix][iy]
				// 陸地でない場合のみ島を配置
				if tile.Type == W2TileVariableOcean {
					g.World2.Tiles[ix][iy].Type = W2TileTransit
					g.World2.Tiles[ix][iy].Source = SrcBRouteIsland
					gen.NewSoils[iy*w+ix] = true
				}
			}

			currX, currY = nextX, nextY
		}

		// 最後のセグメントを終点まで描画
		markPathWithColor(int(currX), int(currY), x2, y2, w, h, SrcBRoutePath)
	}

	// 経由島内部でのランダムウォーク生成 (3x3, 4-7タイル, 周囲3か所に浅瀬)
	genTransitIsland := func(cx, cy int) {
		halfBound := 1 // 3x3 の中心から1マス
		count := 0
		targetCount := 4 + rng.Intn(4) // 4〜7タイル

		// 3x3の島を作成
		wx, wy := cx, cy
		safety := 0
		for count < targetCount && safety < 100 {
			safety++
			if wx >= cx-halfBound && wx <= cx+halfBound && wy >= cy-halfBound && wy <= cy+halfBound {
				tx, ty := wx, wy
				if tx >= 0 && tx < w && ty >= 0 && ty < h && g.World2.Tiles[tx][ty].Type == W2TileVariableOcean {
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

	// 経由島生成
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
				if distToGoal < 5.0 {
					// B航路: 最後の経由地から孤立島まで
					seaWidth := checkSeaWidth(int(destX), int(destY))
					r := rng.Float64()

					if seaWidth == "wide" {
						// 広い海: 弧70% / ジグザグ30%
						if r < 0.7 {
							markArcPathB(int(currX), int(currY), int(destX), int(destY), w, h)
						} else {
							markZigzagPathB(int(currX), int(currY), int(destX), int(destY), w, h)
						}
					} else {
						// 狭い海: ジグザグ40% / 直線60%
						if r < 0.4 {
							markZigzagPathB(int(currX), int(currY), int(destX), int(destY), w, h)
						} else {
							markPathWithColor(int(currX), int(currY), int(destX), int(destY), w, h, SrcBRoutePath)
						}
					}
					break
				}

				// 大陸が近くにあるかチェック（経由島を配置しない）
				tooCloseToLand := false
				checkRadius := 6 // 6マス以内に大陸があれば停止
				for dy := -checkRadius; dy <= checkRadius; dy++ {
					for dx := -checkRadius; dx <= checkRadius; dx++ {
						tx, ty := ix+dx, iy+dy
						if tx >= 0 && tx < w && ty >= 0 && ty < h {
							t := g.World2.Tiles[tx][ty]
							// 孤立島以外の土地（大陸）が近くにある
							if (t.Type == W2TileSoil || t.Type == W2TileCliff) && t.Source != SrcIsland {
								tooCloseToLand = true
								break
							}
						}
					}
					if tooCloseToLand { break }
				}
				if tooCloseToLand {
					// 大陸に近づいたので、B航路で孤立島に直接つなぐ
					seaWidth := checkSeaWidth(int(destX), int(destY))
					r := rng.Float64()
					if seaWidth == "wide" {
						if r < 0.7 {
							markArcPathB(int(currX), int(currY), int(destX), int(destY), w, h)
						} else {
							markZigzagPathB(int(currX), int(currY), int(destX), int(destY), w, h)
						}
					} else {
						if r < 0.4 {
							markZigzagPathB(int(currX), int(currY), int(destX), int(destY), w, h)
						} else {
							markPathWithColor(int(currX), int(currY), int(destX), int(destY), w, h, SrcBRoutePath)
						}
					}
					break
				}

				// A航路: 経由島間の航路（曲線50% / 直線50%）
				if rng.Float64() < 0.5 {
					markCurvedPathA(int(currX), int(currY), ix, iy, w, h)
				} else {
					markPath(int(currX), int(currY), ix, iy, w, h)
				}

				// 経由島の生成
				genTransitIsland(ix, iy)

				currX, currY = nextX, nextY
				if calcDist(int(currX), int(currY), int(destX), int(destY)) < float64(gen.Config.TransitDist) {
					// B航路: 最後の経由地から孤立島まで
					seaWidth := checkSeaWidth(int(destX), int(destY))
					r := rng.Float64()

					if seaWidth == "wide" {
						// 広い海: 弧70% / ジグザグ30%
						if r < 0.7 {
							markArcPathB(int(currX), int(currY), int(destX), int(destY), w, h)
						} else {
							markZigzagPathB(int(currX), int(currY), int(destX), int(destY), w, h)
						}
					} else {
						// 狭い海: ジグザグ40% / 直線60%
						if r < 0.4 {
							markZigzagPathB(int(currX), int(currY), int(destX), int(destY), w, h)
						} else {
							markPathWithColor(int(currX), int(currY), int(destX), int(destY), w, h, SrcBRoutePath)
						}
					}
					break
				}
			}
		}
	}
}