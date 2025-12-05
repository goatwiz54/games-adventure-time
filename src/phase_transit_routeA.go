// filename: phase_transit_routeA.go
package main

import (
	"math"
	"math/rand"
)

// PhaseTransitRouteA - 航路探索A（深海迂回航路）
// 深海が配置された後、島から大陸の港への迂回航路を生成する
func (g *Game) PhaseTransitRouteA(w, h int, rng *rand.Rand, gen *World2Generator) {
	// 大島データがない場合はスキップ
	if len(gen.Islands) == 0 {
		gen.PhaseName += " (No islands found)"
		return
	}

	// TODO: 深海を迂回する航路探索アルゴリズムを実装
	// 現在の実装は一時的なプレースホルダー

	// PinkRectsをクリア（深海フェーズで使用済み）
	g.World2.PinkRects = []Rect{}

	// すべての大島に対して処理を実行
	for _, island := range gen.Islands {
		// 島の中心点CP (Center Point)は既に計算済み
		cpX := island.CenterX
		cpY := island.CenterY

		// CPを起点に円を描き、大陸の港を探す
		// TODO: 深海を考慮した港探索に変更
		type PortInfo struct {
			x, y int
			dist float64
		}

		var harbor *PortInfo

		// アルキメデスの螺旋探索 (Archimedean spiral)
		// r = b * theta (thetaは累積ラジアン)
		// 北(-90度)から開始し、時計回りに探索

		maxDist := math.Sqrt(float64(w*w + h*h))

		// 螺旋のパラメータ
		// 1回転(2*PI)で半径がどれくらい増えるか (ピッチ)
		// ピッチを1.0に設定（1回転で半径が1マス増える＝隙間なく探索）
		pitch := 1.0
		b := pitch / (2 * math.Pi)

		// thetaを増やしながら探索
		// thetaの増分は、半径に応じて調整する（外側ほど細かくして、ステップ距離を一定に保つ）
		// ds = r * d_theta => d_theta = ds / r
		// ds (ステップ距離) は 0.5 程度にして、タイルの取りこぼしを防ぐ

		currentTheta := 0.0
		startAngleOffset := -math.Pi / 2.0 // 北(-90度)から開始

		// 無限ループ防止のため、最大回転数を設定（念のため）
		maxTheta := maxDist / b

		for currentTheta <= maxTheta {
			r := b * currentTheta

			// 実際の角度
			angle := currentTheta + startAngleOffset

			x := cpX + int(r*math.Cos(angle))
			y := cpY + int(r*math.Sin(angle))

			if x >= 0 && x < w && y >= 0 && y < h {
				tile := g.World2.Tiles[x][y]
				// 大陸の土のみ（SrcMain または SrcSub）を探す
				if tile.Type == W2TileSoil && (tile.Source == SrcMain || tile.Source == SrcSub) {
					dist := math.Sqrt(float64((x-cpX)*(x-cpX) + (y-cpY)*(y-cpY)))
					harbor = &PortInfo{x: x, y: y, dist: dist}
					break
				}
			}

			// 次のステップ
			if r < 1.0 {
				currentTheta += 0.5 // 半径が小さいときは粗くても良い（中心付近）
			} else {
				currentTheta += 0.5 / r // ステップ距離0.5を維持
			}
		}

		if harbor == nil {
			// この島には大陸への港が見つからなかった（スキップ）
			continue
		}

		// CP-港の直線距離をFT (First Distance)とする
		ft := harbor.dist

		// CP-港の辺を基準に、FT距離の2倍の長さの長方形を描く（一時的なプレースホルダー）
		// TODO: 深海迂回航路に変更
		// rect1: CP-港方向
		// rect2: 反対方向

		// CP-港のベクトルを計算
		dx := harbor.x - cpX
		dy := harbor.y - cpY

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
