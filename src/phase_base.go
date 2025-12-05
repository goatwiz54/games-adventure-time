// filename: phase_base.go
package main

import (
	"math/rand"
)

// BasePhaseProcessor: 全てのフェーズプロセッサの基本実装
// デフォルトの空実装を提供し、各プロセッサは必要なメソッドのみをオーバーライドする
type BasePhaseProcessor struct{}

// Initialize: デフォルト実装（スナップショット復元）
func (p *BasePhaseProcessor) Initialize(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	// レイヤーが完了済みの場合、スナップショットから復元
	if layer.IsComplete {
		g.World2.Tiles = DeepCopyTiles(layer.Tiles, w, h)
		g.World2.PinkRects = DeepCopyRects(layer.PinkRects)
	}
}

// Before: デフォルト実装（何もしない）
func (p *BasePhaseProcessor) Before(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	// 各フェーズで必要に応じてオーバーライドする
}

// TearDown: デフォルト実装（DeepCopyと完了フラグ設定）
func (p *BasePhaseProcessor) TearDown(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	// スナップショット作成
	layer.Tiles = DeepCopyTiles(g.World2.Tiles, w, h)
	layer.PinkRects = DeepCopyRects(g.World2.PinkRects)

	// 完了フラグを立てる
	layer.IsComplete = true
}

// PrepareDisplay: デフォルト実装（何もしない）
func (p *BasePhaseProcessor) PrepareDisplay(g *Game, layerIdx int, w, h int) {
	// 各フェーズで必要に応じてオーバーライドする
}

// Finalize: デフォルト実装（何もしない）
func (p *BasePhaseProcessor) Finalize(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	// 各フェーズで必要に応じてオーバーライドする
}

// Reset: デフォルト実装（完了フラグをクリア、スナップショットを破棄）
func (p *BasePhaseProcessor) Reset(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	layer.IsComplete = false
	layer.Tiles = nil
	layer.PinkRects = nil
}
