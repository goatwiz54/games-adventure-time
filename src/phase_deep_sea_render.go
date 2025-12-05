// filename: phase_deep_sea_display.go
package main

// PrepareDisplay: 深海フェーズのディスプレイタイル更新処理
func (p *PhaseDeepSeaProcessor) PrepareDisplay(g *Game, layerIdx int, w, h int) {
	// 深海フェーズでは特別なディスプレイタイル更新は不要
	// PinkRectsはworld2.goのDrawWorld2で直接描画される
}
