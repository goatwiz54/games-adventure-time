// filename: phase_processors.go
package main

import "math/rand"

// ==================== PhaseProcessor implementations ====================

// PhaseInitProcessor: 初期化フェーズ
type PhaseInitProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseInitProcessor) Initialize(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	// 設定ファイルを読み込む（必要に応じて）
	// TODO: 設定ファイル読み込み処理を追加
	p.BasePhaseProcessor.Initialize(g, layer, layerIdx, w, h, rng, gen)
}

func (p *PhaseInitProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseInit(w, h, rng, gen)
}

func (p *PhaseInitProcessor) GetName() string {
	return "Init"
}

// PhaseSeaProcessor: 可変海の初期化フェーズ
type PhaseSeaProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseSeaProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseSea(w, h, rng, gen)
}

func (p *PhaseSeaProcessor) GetName() string {
	return "Sea (Variable Ocean)"
}

// PhaseFixedSeaProcessor: 固定海の設定フェーズ
type PhaseFixedSeaProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseFixedSeaProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseFixedSea(w, h, rng, gen)
}

func (p *PhaseFixedSeaProcessor) GetName() string {
	return "Fixed Sea (Border)"
}

// PhaseMaskGenProcessor: マスク生成フェーズ
type PhaseMaskGenProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseMaskGenProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseMaskGen(w, h, rng, gen)
}

func (p *PhaseMaskGenProcessor) GetName() string {
	return "Mask Gen (TargetSoilCount)"
}

// PhaseSoilStartProcessor: 土の開始フェーズ
type PhaseSoilStartProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseSoilStartProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseSoilProgress(w, h, rng, gen)
}

func (p *PhaseSoilStartProcessor) GetName() string {
	return "2. Soil Start"
}

// PhaseSoilProgressProcessor: 土の進行フェーズ（3-11）
type PhaseSoilProgressProcessor struct {
	BasePhaseProcessor
	Step int
}

func (p *PhaseSoilProgressProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseSoilProgress(w, h, rng, gen)
}

func (p *PhaseSoilProgressProcessor) GetName() string {
	return "Soil Progress"
}

// PhaseSoil30PctProcessor: 土壌30%生成フェーズ
type PhaseSoil30PctProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseSoil30PctProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseSoil30Pct(w, h, rng, gen)
}

func (p *PhaseSoil30PctProcessor) GetName() string {
	return "Soil Generation (30%)"
}

// PhaseTectonicShiftProcessor: 地殻変動フェーズ
type PhaseTectonicShiftProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseTectonicShiftProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseTectonicShift(w, h, rng, gen)
}

func (p *PhaseTectonicShiftProcessor) GetName() string {
	return "Tectonic Shift"
}

// PhaseSoilCompleteProcessor: 土壌生成を100%まで完了させる
type PhaseSoilCompleteProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseSoilCompleteProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseSoilComplete(w, h, rng, gen)
}

func (p *PhaseSoilCompleteProcessor) GetName() string {
	return "Soil Generation (100%)"
}

// PhaseBridgeProcessor: 橋フェーズ
type PhaseBridgeProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseBridgeProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseBridge(w, h, rng, gen)
}

func (p *PhaseBridgeProcessor) GetName() string {
	return "Bridge"
}

// PhaseCenteringProcessor: 中央配置フェーズ
type PhaseCenteringProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseCenteringProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseCentering(w, h, rng, gen)
}

func (p *PhaseCenteringProcessor) GetName() string {
	return "Centering"
}

// PhaseIslandsQuadProcessor: 大島生成フェーズ
type PhaseIslandsQuadProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseIslandsQuadProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseIslandsQuad(w, h, rng, gen)
}

func (p *PhaseIslandsQuadProcessor) GetName() string {
	return "Islands (Quad)"
}

// PhaseIslandsRandProcessor: 小島生成フェーズ
type PhaseIslandsRandProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseIslandsRandProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseIslandsRand(w, h, rng, gen)
}

func (p *PhaseIslandsRandProcessor) GetName() string {
	return "Islands (Rand)"
}

// PhaseTransitStartProcessor: 航路開始フェーズ
type PhaseTransitStartProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseTransitStartProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseTransitStart(w, h, rng, gen)
}

func (p *PhaseTransitStartProcessor) GetName() string {
	return "Transit Start"
}

// PhaseIslandShallowAdjustProcessor: 島浅瀬調整フェーズ
type PhaseIslandShallowAdjustProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseIslandShallowAdjustProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseIslandShallowAdjust(w, h, rng, gen)
}

func (p *PhaseIslandShallowAdjustProcessor) GetName() string {
	return "Island Shallow Adjust"
}

// PhaseTransitRouteAProcessor: 航路探索Aフェーズ
type PhaseTransitRouteAProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseTransitRouteAProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseTransitRouteA(w, h, rng, gen)
}

func (p *PhaseTransitRouteAProcessor) GetName() string {
	return "Transit Route A"
}

// PhaseDeepSeaProcessor: 深海フェーズ
type PhaseDeepSeaProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseDeepSeaProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseDeepSea(w, h, rng, gen)
}

func (p *PhaseDeepSeaProcessor) GetName() string {
	return "Deep Sea"
}

// PhaseHarborGenerationProcessor: 港生成フェーズ
type PhaseHarborGenerationProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseHarborGenerationProcessor) Before(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	// 作業タイルを初期化
	layer.WorkTiles = make([][]WorkTile, w)
	for x := 0; x < w; x++ {
		layer.WorkTiles[x] = make([]WorkTile, h)
	}

	// 親クラスのBefore処理を呼び出す
	p.BasePhaseProcessor.Before(g, layer, layerIdx, w, h, rng, gen)
}

func (p *PhaseHarborGenerationProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseHarborGeneration(w, h, rng, gen)
}

func (p *PhaseHarborGenerationProcessor) GetName() string {
	return "Harbor Generation"
}

// PhaseCliffsShallowsProcessor: 崖・浅瀬フェーズ
type PhaseCliffsShallowsProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseCliffsShallowsProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseCliffsShallows(w, h, rng, gen)
}

func (p *PhaseCliffsShallowsProcessor) GetName() string {
	return "Cliffs & Shallows"
}

// PhaseLakesFinalProcessor: 湖生成フェーズ
type PhaseLakesFinalProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseLakesFinalProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	g.PhaseLakesFinal(w, h, rng, gen)
}

func (p *PhaseLakesFinalProcessor) GetName() string {
	return "Lakes Final"
}

// PhaseEndProcessor: 終了フェーズ
type PhaseEndProcessor struct {
	BasePhaseProcessor
}

func (p *PhaseEndProcessor) Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator) {
	gen.IsFinished = true
}

func (p *PhaseEndProcessor) GetName() string {
	return "End"
}

// ==================== Phases配列の初期化 ====================

// InitPhases: Phases配列を初期化
func InitPhases() []PhaseInfo {
	phases := []PhaseInfo{
		{Processor: &PhaseInitProcessor{}},                // 0: 初期化
		{Processor: &PhaseSeaProcessor{}},                 // 1: 可変海の初期化
		{Processor: &PhaseFixedSeaProcessor{}},            // 2: 固定海の設定
		{Processor: &PhaseMaskGenProcessor{}},             // 3: TargetSoilCount計算
		{Processor: &PhaseSoil30PctProcessor{}},           // 4: 土壌30%生成
		{Processor: &PhaseTectonicShiftProcessor{}},       // 5: 地殻変動
		{Processor: &PhaseSoilCompleteProcessor{}},        // 6: 土壌100%生成
		{Processor: &PhaseBridgeProcessor{}},              // 7: 橋の生成
		{Processor: &PhaseCenteringProcessor{}},           // 8: 中央配置調整
		{Processor: &PhaseIslandsQuadProcessor{}},         // 9: 大島生成
		{Processor: &PhaseIslandsRandProcessor{}},         // 10: 小島生成
		{Processor: &PhaseTransitStartProcessor{}},        // 11: 航路開始点設定
		{Processor: &PhaseIslandShallowAdjustProcessor{}}, // 12: 島周辺の浅瀬調整
		{Processor: &PhaseDeepSeaProcessor{}},             // 13: 深海領域生成
		{Processor: &PhaseHarborGenerationProcessor{}},    // 14: 港生成
		{Processor: &PhaseTransitRouteAProcessor{}},       // 15: 航路探索A（深海迂回）
		{Processor: &PhaseCliffsShallowsProcessor{}},      // 16: 崖・浅瀬生成
		{Processor: &PhaseLakesFinalProcessor{}},          // 17: 湖生成
		{Processor: &PhaseEndProcessor{}},                 // 18: 終了
	}

	return phases
}
