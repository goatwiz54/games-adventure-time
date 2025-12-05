// filename: types.go
package main

import (
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- 設定定数 ---
const (
	ScreenWidth   = 1280
	ScreenHeight  = 720
	TileWidth     = 64
	TileHeight    = 32
	BaseTurnToMin = 6
	MoveSpeed     = 0.20

	WorldWidth    = 100
	WorldHeight   = 80
	WorldTileSize = 32

	World2Width         = 150
	World2Height        = 100
	DefaultWorld2Width  = 150
	DefaultWorld2Height = 100
	World2TileSize      = 16
)

const (
	DirNorth = 0
	DirEast  = 1
	DirSouth = 2
	DirWest  = 3
)

const (
	W2TileVariableOcean = 0
	W2TileSoil          = 1
	W2TileFixedOcean    = 2
	W2TileTransit       = 3
	W2TileCliff         = 4
	W2TileShallow       = 5
	W2TileDeepSea       = 8 // 深海
	W2TileVeryDeepSea   = 9 // 大深海（深海の重複）
)

const (
	SrcNone         = 0
	SrcMain         = 1
	SrcSub          = 2
	SrcMix          = 3
	SrcBridge       = 4
	SrcIsland       = 5
	SrcTransitPath  = 6
	SrcBRouteIsland = 7 // B航路の経由島（緑）
	SrcBRoutePath   = 8 // B航路の航路（暗緑）
)

// WorkTile データタイプ
const (
	WorkDataNone       = 0 // データなし
	WorkDataSpiral     = 1 // 螺旋探索点
	WorkDataHarborCandidate = 2 // 港候補点
)

// WorkTile 表示色
const (
	WorkColorNone        = 0 // 表示なし
	WorkColorYellow      = 1 // 黄色（螺旋探索点）
	WorkColorBrightGreen = 2 // 明るい緑（港候補点）
)

// Input Edit Modes
const (
	EditNone        = 0
	EditSoilMin     = 1
	EditSoilMax     = 2
	EditW2Width     = 3
	EditW2Height    = 4
	EditTransitDist = 5
	EditCentering   = 6
	EditCliffInit   = 7
	EditCliffDec    = 8
	EditShallowDec  = 9
	EditCliffPath   = 10
	EditForceSwitch = 11
	EditMapRatio    = 14
)


var ZoomLevels = []float64{0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5}

// --- ゲーム状態 ---
type GameState int

const (
	StateMenu GameState = iota
	StateWorldLoading
	StateWorld
	StateDungeon
	StateWorld2
)

// --- リソース ---
var (
	TexGrass, TexDirt, TexStone, TexWhite, TexArrow                  *ebiten.Image
	TexW_MountainIcon, TexW_TreeIcon, TexW_CityIcon                  *ebiten.Image
	TexW_Ocean, TexW_Plains, TexW_Forest, TexW_Desert, TexW_Mountain *ebiten.Image
	TexW2_Ocean, TexW2_Soil, TexW2_FixedOcean                        *ebiten.Image
)

type Status struct{ HP, SP, STR, DEX, VIT, INT, AGI, SPD, LUCK int }

type Character struct {
	Name                         string
	Stats                        Status
	BaseWT                       int
	LoadWeight                   float64
	CurrentX, CurrentY, CurrentZ float64
	TargetX, TargetY, TargetZ    int
	IsMoving                     bool
	Facing                       int
}

type Party struct {
	Leader     *Character
	Members    []*Character
	TotalTurns int
	InCombat   bool
	CombatLog  string
}

type Enemy struct {
	ID                           int
	CurrentX, CurrentY, CurrentZ float64
	TargetX, TargetY, TargetZ    int
	IsMoving                     bool
	Facing                       int
	Type, Speed                  int
	Active                       bool
}

type Tile struct {
	Type, Height int
	Explored     bool
}

type Dungeon struct {
	Width, Height int
	Tiles         [][]Tile
	Enemies       []*Enemy
}

type WorldTile struct {
	Biome  int
	Height float64
	IsRoad bool
}
type WorldMap struct {
	Tiles            [][]WorldTile
	CameraX, CameraY float64
	Zoom             float64
}

type Rect struct {
	x, y, w, h int
}

type World2Tile struct {
	Type    int
	Source  int
	IsLake  bool
	IsEmpty bool // レイヤーでこのタイルが空かどうか
}

// WorkTile: 作業用タイル（フェーズの中間データ保存用）
type WorkTile struct {
	DataType int // データタイプ（0=空, 1=螺旋探索点, 2=港候補点, など）
	Color    int // 表示色（0=なし, 1=黄色, 2=明るい緑, など）
}

// PhaseProcessor: 各フェーズの処理を実装するインターフェース
type PhaseProcessor interface {
	// Initialize: レイヤーに入る前の初期化処理（ワールド・ゲーム設定、スナップショット復元）
	// 完了フラグに関わらず常に呼ばれる
	Initialize(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator)

	// Before: レイヤー構造の初期化処理（Walker初期化、レイヤー固有の前処理）
	// 完了フラグがfalseの場合のみ呼ばれる
	Before(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator)

	// Execute: メイン処理
	// 完了フラグがfalseの場合のみ呼ばれる
	Execute(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator)

	// TearDown: 後処理（DeepCopyとフラグ設定の前に実行される）
	// 完了フラグがfalseの場合のみ呼ばれる
	TearDown(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator)

	// PrepareDisplay: レイヤー固有のディスプレイタイル更新処理
	// 完了フラグに関わらず常に呼ばれる
	PrepareDisplay(g *Game, layerIdx int, w, h int)

	// Finalize: 次のフェーズに進む前の最終処理
	// 完了フラグに関わらず常に呼ばれる
	Finalize(g *Game, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator)

	// Reset: レイヤーのリセット処理（Rキー押下時）
	Reset(g *Game, layer *Layer, layerIdx int, w, h int, rng *rand.Rand, gen *World2Generator)

	// GetName: フェーズ名を取得
	GetName() string
}

// PhaseInfo: フェーズ情報
type PhaseInfo struct {
	Processor PhaseProcessor
}

// Layer: 各フェーズのレイヤーデータ
type Layer struct {
	Tiles      [][]World2Tile // このレイヤーのタイル（合成済み）
	IsComplete bool           // 完了フラグ
	PinkRects  []Rect         // スナップショット用
	WorkTiles  [][]WorkTile   // 作業用タイル（各フェーズの中間データ）
}

type WorldMap2 struct {
	Width, Height    int
	Tiles            [][]World2Tile // メインのタイルデータ（描画・処理用）
	Phases           []PhaseInfo    // フェーズ定義配列
	Layers           []Layer        // レイヤー配列（スナップショット保存用）
	CurrentLayerIdx  int            // 現在のレイヤーインデックス
	OffsetX, OffsetY float64
	Zoom             float64
	ShowGrid         bool
	StatsInfo        []string
	PinkRects        []Rect
}

type GenSnapshot struct {
	Tiles     [][]World2Tile
	PhaseName string
	StepID    int

	NewSoils         map[int]bool
	PinkRects        []Rect
	Walkers          []struct{ x, y int }
	CurrentSoilCount int
	Multiplier       float64
	Excluded         map[int]bool
	CurrentSeed      int64

	CliffStreak   int
	ShallowStreak int
}

type World2Generator struct {
	CurrentStep int
	IsFinished  bool
	PhaseName   string
	History     []GenSnapshot

	Rng         *rand.Rand
	CurrentSeed int64

	MaskMain  [][]float64
	MaskSub   [][]float64
	FinalMask [][]float64

	MaskImage *ebiten.Image

	Walkers          []struct{ x, y int }
	TargetSoilCount  int
	CurrentSoilCount int
	Config           GenConfig

	NewSoils map[int]bool

	Multiplier float64
	Excluded   map[int]bool

	CliffStreak   int
	ShallowStreak int

	// 大島データ（Phase_IslandsQuad で生成、Phase_IslandsRandの小島は含まない）
	Islands []IslandData
}

// IslandData は各大島の情報を保持（Phase_IslandsQuadで生成された島のみ）
// 航路探索Aで使用される
type IslandData struct {
	Tiles                  []struct{ x, y int } // 大島を構成するタイルの座標リスト
	MinX, MinY, MaxX, MaxY int                  // 大島がすっぽり入る矩形
	CenterX, CenterY       int                  // 大島矩形の中心点

	// 港関連データ（PhaseHarborGenerationで設定）
	IslandSeaPortX, IslandSeaPortY int // 大島側の港（深海の反対側）
	HarborCandidateAX, HarborCandidateAY int // 最終港候補A（C字型の湾内）
	HarborCandidateBX, HarborCandidateBY int // 港候補B（最初の大陸の土）
	HarborCandidateCX, HarborCandidateCY int // 港候補C（B以降の最初の海）
}

type GenConfig struct {
	MinPct, MaxPct, W, H, TransitDist, Ratio int
	VastOcean, IslandBound                   int
	Centering                                bool
	CliffInit, CliffDec, ShallowDec          float64
	CliffPathLen, ForceSwitch                int
	MainType, SubType                        int
}

type Camera struct {
	X, Y, Angle, TargetAngle float64
	ZoomIndex                int
}

type LoadingState struct {
	StartTime time.Time
	Progress  float64
	Step      int
	Logs      []LoadingLog
}
type LoadingLog struct {
	Msg     string
	AddedAt time.Time
}

type Game struct {
	State                    GameState
	MenuIndex                int
	Loader                   *LoadingState
	World                    *WorldMap
	World2                   *WorldMap2
	Gen2                     *World2Generator
	Dungeon                  *Dungeon
	Party                    *Party
	Camera                   *Camera
	Log                      []string
	Rng                      *rand.Rand
	DebugMode                bool
	MouseStartX, MouseStartY int
	IsDragging               bool
	ArrowTimer               float64

	TotalRoute1Dist float64

	SoilMin         int
	SoilMax         int
	W2Width         int
	W2Height        int
	TransitDist     int
	VastOceanSize   int
	IslandBoundSize int

	MapRatio        int
	EnableCentering bool

	CliffInitVal  float64
	CliffDecVal   float64
	ShallowDecVal float64
	CliffPathLen  int
	ForceSwitch   int

	LastTargetSoil int
	InputMode      int
	InputBuffer    string

	WarningMsg   string
	WarningTimer float64

	AutoProgress    bool // Enterキー押下時に自動で次のフェーズへ進むフラグ
	SuppressMapDraw bool // 自動進行中はマップ描画を抑制し、Phase名のみ表示
}
