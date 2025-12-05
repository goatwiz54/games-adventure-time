// filename: world2_layers.go
package main

import "fmt"

// DeepCopyTiles: タイル配列のディープコピー
func DeepCopyTiles(src [][]World2Tile, w, h int) [][]World2Tile {
	if src == nil {
		return nil
	}

	dst := make([][]World2Tile, w)
	for x := 0; x < w; x++ {
		dst[x] = make([]World2Tile, h)
		copy(dst[x], src[x])
	}

	return dst
}

// DeepCopyRects: Rectスライスのディープコピー
func DeepCopyRects(src []Rect) []Rect {
	if src == nil {
		return nil
	}
	dst := make([]Rect, len(src))
	copy(dst, src)
	return dst
}

// ExecuteCurrentPhase: 現在のフェーズを実行
func (g *Game) ExecuteCurrentPhase() {
	i := g.World2.CurrentLayerIdx

	// 範囲チェック
	if i >= len(g.World2.Phases) {
		g.Gen2.IsFinished = true
		return
	}

	layer := &g.World2.Layers[i]
	phase := g.World2.Phases[i]

	// フェーズ名を設定（番号を付与）
	g.Gen2.PhaseName = fmt.Sprintf("%d. %s", i, phase.Processor.GetName())

	// 1. Initialize: ワールド・ゲーム設定、スナップショット復元（常に実行）
	phase.Processor.Initialize(g, layer, i, g.W2Width, g.W2Height, g.Gen2.Rng, g.Gen2)

	// 2. 未完了の場合のみ Before → Execute → TearDown を実行
	if !layer.IsComplete {
		// Before: レイヤー構造初期化
		phase.Processor.Before(g, layer, i, g.W2Width, g.W2Height, g.Gen2.Rng, g.Gen2)

		// Execute: メイン処理
		phase.Processor.Execute(g, i, g.W2Width, g.W2Height, g.Gen2.Rng, g.Gen2)

		// TearDown: 後処理（各プロセッサがDeepCopyと完了フラグ設定を行う）
		phase.Processor.TearDown(g, layer, i, g.W2Width, g.W2Height, g.Gen2.Rng, g.Gen2)
	}

	// 3. PrepareDisplay: ディスプレイタイル更新処理（常に実行）
	phase.Processor.PrepareDisplay(g, i, g.W2Width, g.W2Height)

	// 4. Finalize: 次のフェーズに進む前の最終処理（常に実行）
	phase.Processor.Finalize(g, i, g.W2Width, g.W2Height, g.Gen2.Rng, g.Gen2)
}

// OnPageDown: 次のフェーズへ
func (g *Game) OnPageDown() {
	if g.World2.CurrentLayerIdx < len(g.World2.Phases)-1 {
		g.World2.CurrentLayerIdx++
		// スナップショット復元はInitialize()内で行われる
		g.ExecuteCurrentPhase()
	}
}

// OnPageUp: 前のフェーズへ
func (g *Game) OnPageUp() {
	if g.World2.CurrentLayerIdx > 0 {
		g.World2.CurrentLayerIdx--
		// スナップショット復元とフェーズ名更新はInitialize()とExecuteCurrentPhase()内で行われる
		g.ExecuteCurrentPhase()
	}
}

// OnEnterKey: 自動進行（最後まで実行）
func (g *Game) OnEnterKey() {
	// 最後のフェーズならリセットして再開
	if g.World2.CurrentLayerIdx >= len(g.World2.Phases)-1 {
		g.OnResetGeneration()
		return
	}

	g.AutoProgress = true

	for g.World2.CurrentLayerIdx < len(g.World2.Phases)-1 {
		g.OnPageDown()
	}

	g.AutoProgress = false
	g.Gen2.IsFinished = true
}

// OnResetGeneration: リセット（最初から再生成）
func (g *Game) OnResetGeneration() {
	// 全てのレイヤーのReset()を呼ぶ
	for i := 0; i < len(g.World2.Layers); i++ {
		layer := &g.World2.Layers[i]
		phase := g.World2.Phases[i]
		phase.Processor.Reset(g, layer, i, g.W2Width, g.W2Height, g.Gen2.Rng, g.Gen2)
	}

	// インデックスを0に戻す
	g.World2.CurrentLayerIdx = 0

	// Tilesを初期状態に戻す（空）
	for x := 0; x < g.World2.Width; x++ {
		for y := 0; y < g.World2.Height; y++ {
			g.World2.Tiles[x][y] = World2Tile{IsEmpty: true}
		}
	}

	// PinkRectsをクリア
	g.World2.PinkRects = nil

	// Generatorの状態をリセット
	g.Gen2.IsFinished = false
	g.Gen2.CurrentStep = 0
	g.Gen2.CurrentSoilCount = 0
	g.Gen2.Walkers = nil
	g.Gen2.NewSoils = make(map[int]bool)
	g.Gen2.Islands = nil

	// 最初のフェーズを実行
	g.ExecuteCurrentPhase()
}
