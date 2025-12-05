// filename: phase_soil_complete.go
package main

import (
	"fmt"
	"math/rand"
)

// PhaseSoilComplete: 土壌生成を100%まで完了させる（Tectonic Shift後の残り70%）
func (g *Game) PhaseSoilComplete(w, h int, rng *rand.Rand, gen *World2Generator) {
	// ターゲット土壌数を取得（PhaseMaskGenで計算済み）
	target := gen.TargetSoilCount

	// 土壌配置関数：指定座標に土壌を配置
	placeSoil := func(x, y int, srcOverride int) bool {
		// 固定海の外周3マスを避ける
		if x >= 3 && x < w-3 && y >= 3 && y < h-3 {
			// 可変海の場合のみ土壌に変換
			if g.World2.Tiles[x][y].Type == W2TileVariableOcean {
				g.World2.Tiles[x][y].Type = W2TileSoil

				// ソース（起源）を設定
				if srcOverride != SrcNone {
					g.World2.Tiles[x][y].Source = srcOverride
				} else {
					g.World2.Tiles[x][y].Source = SrcMain
				}

				// 新規土壌として記録
				gen.NewSoils[y*w+x] = true
				return true
			}
		}
		return false
	}

	// スポーン位置を決定する関数：マップ中心付近にランダムに配置
	findSpawn := func() (int, int) {
		cx, cy := w/2, h/2
		// 中心から±10マスの範囲でランダム配置
		return cx + rng.Intn(20) - 10, cy + rng.Intn(20) - 10
	}

	// Walkerの初期化チェック（Tectonic Shift後なので既に存在する場合はスキップ）
	if len(gen.Walkers) == 0 {
		walkers := 50
		gen.Walkers = make([]struct{ x, y int }, walkers)
		for i := 0; i < walkers; i++ {
			sx, sy := findSpawn()
			gen.Walkers[i].x, gen.Walkers[i].y = sx, sy
		}
	}

	// 土壌生成のメインループ：ターゲット数まで繰り返す
	safety := 0 // 無限ループ防止用カウンター
	for gen.CurrentSoilCount < target && safety < 500000 {
		safety++

		// 全てのWalkerを処理
		for i := range gen.Walkers {
			// 現在位置に土壌を配置（可能なら）
			if placeSoil(gen.Walkers[i].x, gen.Walkers[i].y, SrcNone) {
				gen.CurrentSoilCount++
			}

			// 次の移動方向を決定（スコアベース）
			dir := rng.Intn(4)
			bestScore := -1.0
			bestDir := dir
			dxs := []int{0, 1, 0, -1}  // 方向ベクトルX（上、右、下、左）
			dys := []int{-1, 0, 1, 0}  // 方向ベクトルY（上、右、下、左）

			// 4方向のスコアを計算して最適方向を選択
			for d := 0; d < 4; d++ {
				nx, ny := gen.Walkers[i].x+dxs[d], gen.Walkers[i].y+dys[d]
				score := 0.0

				// FinalMaskベースのスコア計算
				if nx >= 0 && nx < w && ny >= 0 && ny < h {
					score = gen.FinalMask[nx][ny]
				}

				// ランダム性を加えて自然な形状にする
				score += rng.Float64() * 0.5
				if score > bestScore {
					bestScore = score
					bestDir = d
				}
			}

			// スコアが低すぎる、または固定海の境界に達した場合はリスポーン
			if bestScore < 0.1 || gen.Walkers[i].x < 3 || gen.Walkers[i].x >= w-3 || gen.Walkers[i].y < 3 || gen.Walkers[i].y >= h-3 {
				nx, ny := findSpawn()
				gen.Walkers[i].x, gen.Walkers[i].y = nx, ny
			} else {
				// 最適方向に移動
				gen.Walkers[i].x += dxs[bestDir]
				gen.Walkers[i].y += dys[bestDir]
			}
		}
	}

	// フェーズ名を最終的な生成数で更新（パーセント表示付き）
	gen.PhaseName = fmt.Sprintf("2. Soil Generation: 100%% (%d/%d)", gen.CurrentSoilCount, target)
}
