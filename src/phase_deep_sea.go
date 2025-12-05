// filename: phase_deep_sea.go
package main

import (
	"math"
	"math/rand"
)

// PhaseDeepSea: 深海処理
// 各大島に対して、深海ベルトを計算し、その中心に楕円形の深海エリアを生成
func (g *Game) PhaseDeepSea(w, h int, rng *rand.Rand, gen *World2Generator) {
	// 深海領域の生成（楕円形）大島データがない場合はスキップ
	if len(gen.Islands) == 0 {
		gen.PhaseName += " (No islands found)"
		return
	}

	// 各大島に対して深海エリアを生成
	for _, island := range gen.Islands {
		cpX := island.CenterX
		cpY := island.CenterY

		// CP-DSP方向を計算（大陸との深海候補点を探す）
		var deepSeaPointX, deepSeaPointY int
		var foundDSP bool

		// 北東南西の方向を定義
		directions := []struct {
			name  string
			check func(angle float64) bool
		}{
			{"North", func(angle float64) bool { return angle >= 225 && angle < 315 }},
			{"East", func(angle float64) bool { return angle >= 315 || angle < 45 }},
			{"South", func(angle float64) bool { return angle >= 45 && angle < 135 }},
			{"West", func(angle float64) bool { return angle >= 135 && angle < 225 }},
		}

		// 円を拡大しながら大陸の土を探す
		maxRadius := int(math.Sqrt(float64(w*w + h*h)))
		for _, dir := range directions {
			found := false
			for radius := 1; radius <= maxRadius && !found; radius++ {
				for angle := 0.0; angle < 360.0; angle += 1.0 {
					if !dir.check(angle) {
						continue
					}

					rad := angle * math.Pi / 180.0
					x := cpX + int(float64(radius)*math.Cos(rad))
					y := cpY + int(float64(radius)*math.Sin(rad))

					if x >= 0 && x < w && y >= 0 && y < h {
						tile := g.World2.Tiles[x][y]
						if tile.Type == W2TileSoil && (tile.Source == SrcMain || tile.Source == SrcSub) {
							deepSeaPointX, deepSeaPointY = x, y
							foundDSP = true
							found = true
							break
						}
					}
				}
			}
			if found {
				break
			}
		}

		if !foundDSP {
			continue
		}

		// CP-DSPのベクトルを計算
		dx := deepSeaPointX - cpX
		dy := deepSeaPointY - cpY
		length := math.Sqrt(float64(dx*dx + dy*dy))

		if length == 0 {
			continue
		}

		// 正規化されたベクトル
		normDx := float64(dx) / length
		normDy := float64(dy) / length

		// 垂直方向のベクトル
		perpDx := -normDy
		perpDy := normDx

		// 深海ベルトの長さ（FT * 2）
		deepSeaBeltLength := int(length * 2)

		// 深海ベルトの幅（固定10マス）
		deepSeaBeltWidth := 10

		// 楕円のサイズ（深海ベルトに対して垂直に描画）
		// 楕円の長軸：深海ベルトの幅を超えて垂直方向に伸びる（30〜50マス）
		ellipseMajorAxis := deepSeaBeltWidth * (3 + rng.Intn(3)) // 30〜50マス

		// 楕円の短軸：深海ベルトの長さの1/3〜1/2
		ellipseRatio := 0.33 + rng.Float64()*0.17 // 0.33 〜 0.5
		ellipseMinorAxis := int(float64(deepSeaBeltLength) * ellipseRatio)

		// 深海ベルトの中心点を計算（CP-DSP間の中間点）
		midX := cpX + int(normDx*length/2)
		midY := cpY + int(normDy*length/2)

		// 楕円を描画（深海ベルトに対して垂直方向に伸びる）
		// 長軸 = 垂直方向（perpDx, perpDy）
		// 短軸 = CP-DSP方向（normDx, normDy）
		g.drawEllipse(midX, midY, ellipseMajorAxis, ellipseMinorAxis, perpDx, perpDy, w, h)
	}

	_ = rng // 未使用変数警告を回避
}

// drawEllipse: 楕円形の深海エリアを描画
// (cx, cy): 中心座標
// (length, width): 楕円の長さと幅
// (dirX, dirY): 方向ベクトル（正規化済み）
func (g *Game) drawEllipse(cx, cy, length, width int, dirX, dirY float64, w, h int) {
	halfLength := length / 2
	halfWidth := width / 2

	// ゼロ除算を防ぐ
	if halfLength == 0 || halfWidth == 0 {
		return
	}

	// 垂直方向のベクトル（dirに対して垂直）
	perpDx := -dirY
	perpDy := dirX

	// 訪問済みピクセルを記録
	visited := make(map[int]bool)

	// 楕円の枠線のみを描画（デバッグ用）
	// 角度を0〜360度で回してピクセルを配置
	for angle := 0.0; angle < 360.0; angle += 0.5 {
		rad := angle * math.Pi / 180.0

		// 楕円のパラメトリック方程式
		// x = a * cos(θ), y = b * sin(θ)
		localX := float64(halfLength) * math.Cos(rad)
		localY := float64(halfWidth) * math.Sin(rad)

		// ワールド座標に変換
		wx := cx + int(dirX*localX+perpDx*localY)
		wy := cy + int(dirY*localX+perpDy*localY)

		if wx < 0 || wx >= w || wy < 0 || wy >= h {
			continue
		}

		// 重複訪問チェック
		pixelKey := wy*w + wx
		if visited[pixelKey] {
			continue
		}
		visited[pixelKey] = true

		tile := &g.World2.Tiles[wx][wy]

		// 固定海、土タイルは変更しない
		if tile.Type == W2TileFixedOcean || tile.Type == W2TileSoil {
			continue
		}

		// 枠線のみ描画
		if tile.Type == W2TileVariableOcean {
			tile.Type = W2TileDeepSea
		}
	}
}
