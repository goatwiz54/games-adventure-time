// filename: phase_island_shallow_adjust.go
package main

import (
	"math"
	"math/rand"
)

// 円状の島の内側に浅瀬を生成する
func (g *Game) PhaseIslandShallowAdjust(w, h int, rng *rand.Rand, gen *World2Generator) {
	gen.PhaseName = "8.5. Island Shallow Adjust (Skipped)"
	// スキップ
	//return

	// 1. B航路の経由島（緑の小島）をすべて収集
	type Point struct{ x, y int }
	var routeIslands []Point

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			tile := g.World2.Tiles[x][y]
			if tile.Source == SrcBRouteIsland {
				routeIslands = append(routeIslands, Point{x, y})
			}
		}
	}

	// B航路の島が5個未満なら円状とみなさない
	if len(routeIslands) < 5 {
		return
	}

	// 2. 円の中心を計算（重心）
	centerX, centerY := 0.0, 0.0
	for _, p := range routeIslands {
		centerX += float64(p.x)
		centerY += float64(p.y)
	}
	centerX /= float64(len(routeIslands))
	centerY /= float64(len(routeIslands))

	// 3. 円の半径を計算（中心から各島までの平均距離）
	radius := 0.0
	for _, p := range routeIslands {
		dx := float64(p.x) - centerX
		dy := float64(p.y) - centerY
		dist := math.Sqrt(dx*dx + dy*dy)
		radius += dist
	}
	radius /= float64(len(routeIslands))

	// 半径が小さすぎる場合はスキップ
	if radius < 10.0 {
		return
	}

	// 4. フェーズ1: 円の内側で土地に隣接する海を浅瀬化
	dxs := []int{0, 1, 0, -1}
	dys := []int{-1, 0, 1, 0}

	firstShallows := 0
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			// 円の内側かチェック
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			distFromCenter := math.Sqrt(dx*dx + dy*dy)

			if distFromCenter < radius {
				tile := g.World2.Tiles[x][y]

				// 海タイルで、土地に隣接しているか
				if tile.Type == W2TileVariableOcean {
					hasAdjacentLand := false
					for i := 0; i < 4; i++ {
						nx, ny := x+dxs[i], y+dys[i]
						if nx >= 0 && nx < w && ny >= 0 && ny < h {
							nt := g.World2.Tiles[nx][ny]
							// B航路の経由島（SrcBRouteIsland）のみを対象にする
							if (nt.Type == W2TileSoil || nt.Type == W2TileTransit || nt.Type == W2TileCliff) && nt.Source == SrcBRouteIsland {
								hasAdjacentLand = true
								break
							}
						}
					}

					if hasAdjacentLand {
						g.World2.Tiles[x][y].Type = W2TileShallow
						gen.NewSoils[y*w+x] = true
						firstShallows++
					}
				}
			}
		}
	}

	// 浅瀬が1つも作られなかった場合は終了
	if firstShallows == 0 {
		return
	}

	// 5. フェーズ2: 浅瀬の隣接カウント
	adjacencyCount := make(map[int]int)

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			tile := g.World2.Tiles[x][y]

			// 浅瀬タイルの場合
			if tile.Type == W2TileShallow {
				// 上下左右の海タイルにフラグ+1
				for i := 0; i < 4; i++ {
					nx, ny := x+dxs[i], y+dys[i]
					if nx >= 0 && nx < w && ny >= 0 && ny < h {
						nt := g.World2.Tiles[nx][ny]
						if nt.Type == W2TileVariableOcean {
							idx := ny*w + nx
							adjacencyCount[idx]++
						}
					}
				}
			}
		}
	}

	// 6. フェーズ3: フラグ2以上のタイルを70%確率で浅瀬化
	for idx, count := range adjacencyCount {
		if count >= 2 {
			// 70%の確率で浅瀬化
			if rng.Float64() < 0.7 {
				x := idx % w
				y := idx / w

				// 円の内側かチェック
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				distFromCenter := math.Sqrt(dx*dx + dy*dy)

				if distFromCenter < radius {
					g.World2.Tiles[x][y].Type = W2TileShallow
					gen.NewSoils[idx] = true
				}
			}
		}
	}
}
