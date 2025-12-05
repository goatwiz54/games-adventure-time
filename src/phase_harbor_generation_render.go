// filename: phase_harbor_generation_render.go
package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// PrepareDisplay: 港生成フェーズのディスプレイタイル更新処理
func (p *PhaseHarborGenerationProcessor) PrepareDisplay(g *Game, layerIdx int, w, h int) {
	// 作業タイルのデータをワールドマップタイルに反映（可視化のため）
	layer := &g.World2.Layers[layerIdx]

	// 作業タイルが初期化されていない場合はスキップ
	if layer.WorkTiles == nil {
		return
	}

	// 各作業タイルをチェックして、表示用のマーカーを追加
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			workTile := layer.WorkTiles[x][y]

			if workTile.DataType == WorkDataNone {
				continue
			}

			// 作業タイルのデータに応じて表示処理
			// ここでは、g.World2.Tiles に直接書き込むのではなく、
			// 描画時に色付きの点を描画する
		}
	}
}

// DrawWorkTiles: 作業タイルの描画（港候補点と螺旋探索点）
func DrawWorkTiles(screen *ebiten.Image, g *Game, layerIdx int) {
	w := g.World2.Width
	h := g.World2.Height

	if layerIdx < 0 || layerIdx >= len(g.World2.Layers) {
		return
	}

	layer := &g.World2.Layers[layerIdx]
	if layer.WorkTiles == nil {
		return
	}

	// 各作業タイルを描画
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			workTile := layer.WorkTiles[x][y]

			if workTile.DataType == WorkDataNone {
				continue
			}

			// ワールド座標をスクリーン座標に変換
			screenX := float32(x)*float32(World2TileSize)*float32(g.World2.Zoom) + float32(g.World2.OffsetX)
			screenY := float32(y)*float32(World2TileSize)*float32(g.World2.Zoom) + float32(g.World2.OffsetY)

			// 表示色を取得
			var drawColor color.Color
			switch workTile.Color {
			case WorkColorYellow:
				drawColor = color.RGBA{255, 255, 0, 128} // 黄色（半透明）
			case WorkColorBrightGreen:
				drawColor = color.RGBA{0, 255, 0, 255} // 明るい緑（不透明）
			default:
				continue
			}

			// 点を描画（小さい円）
			radius := float32(2) * float32(g.World2.Zoom)
			if workTile.Color == WorkColorBrightGreen {
				radius = float32(4) * float32(g.World2.Zoom) // 港候補は大きく
			}

			vector.DrawFilledCircle(
				screen,
				screenX+float32(World2TileSize)*float32(g.World2.Zoom)/2,
				screenY+float32(World2TileSize)*float32(g.World2.Zoom)/2,
				radius,
				drawColor,
				false,
			)
		}
	}
}
