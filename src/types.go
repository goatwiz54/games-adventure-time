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
)

const (
	SrcNone          = 0
	SrcMain          = 1
	SrcSub           = 2
	SrcMix           = 3
	SrcBridge        = 4
	SrcIsland        = 5
	SrcTransitPath   = 6
	SrcBRouteIsland  = 7  // B航路の経由島（緑）
	SrcBRoutePath    = 8  // B航路の航路（暗緑）
)

// Input Edit Modes
const (
	EditNone = 0
	EditSoilMin = 1
	EditSoilMax = 2
	EditW2Width = 3
	EditW2Height = 4
	EditTransitDist = 5
	EditCentering   = 6
	EditCliffInit   = 7
	EditCliffDec    = 8
	EditShallowDec  = 9
	EditCliffPath   = 10 
	EditForceSwitch = 11 
	EditMapRatio = 14
)

// World2 Generation Phases (Step ID)
const (
	Phase_Init             = 0
	Phase_MaskGen          = 1
	Phase_SoilStart        = 2
	Phase_SoilProgressEnd  = 11
	Phase_Bridge           = 13
	Phase_Centering        = 14
	Phase_IslandsQuad      = 15
	Phase_IslandsRand      = 16
	
	// Transit Phase の拡張
	Phase_Transit_Start    = 17
	Phase_IslandShallowAdjust = 18
	Phase_Transit_Route1   = 19
	Phase_Transit_Route2_Calc = 20
	Phase_Transit_Route2_Draw = 21

	Phase_CliffsShallows   = 22
	Phase_LakesFinal       = 23 
)

var ZoomLevels = []float64{0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5}

// --- ゲーム状態 ---
type GameState int

const (
	StateMenu         GameState = iota
	StateWorldLoading           
	StateWorld                  
	StateDungeon                
	StateWorld2                 
)

// --- リソース ---
var (
	TexGrass, TexDirt, TexStone, TexWhite, TexArrow    *ebiten.Image
	TexW_MountainIcon, TexW_TreeIcon, TexW_CityIcon    *ebiten.Image
	TexW_Ocean, TexW_Plains, TexW_Forest, TexW_Desert, TexW_Mountain *ebiten.Image
	TexW2_Ocean, TexW2_Soil, TexW2_FixedOcean          *ebiten.Image
)

type Status struct{ HP, SP, STR, DEX, VIT, INT, AGI, SPD, LUCK int }

type Character struct {
	Name       string
	Stats      Status
	BaseWT     int
	LoadWeight float64
	CurrentX, CurrentY, CurrentZ float64
	TargetX, TargetY, TargetZ    int
	IsMoving   bool
	Facing     int
}

type Party struct {
	Leader     *Character
	Members    []*Character
	TotalTurns int
	InCombat   bool
	CombatLog  string
}

type Enemy struct {
	ID         int
	CurrentX, CurrentY, CurrentZ float64
	TargetX, TargetY, TargetZ    int
	IsMoving   bool
	Facing     int
	Type, Speed int
	Active     bool
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
	Type   int
	Source int
	IsLake bool
}

type WorldMap2 struct {
	Width, Height    int
	Tiles            [][]World2Tile
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
	Walkers          []struct{x, y int}
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
	
	Rng             *rand.Rand
	CurrentSeed     int64

	MaskMain        [][]float64
	MaskSub         [][]float64
	FinalMask       [][]float64
	
	MaskImage       *ebiten.Image 

	Walkers         []struct{ x, y int }
	TargetSoilCount int
	CurrentSoilCount int
	Config          GenConfig
	
	NewSoils map[int]bool
	
	Multiplier float64
	Excluded   map[int]bool
	
	CliffStreak   int
	ShallowStreak int
}

type GenConfig struct {
	MinPct, MaxPct, W, H, TransitDist, Ratio int
	VastOcean, IslandBound int
	Centering bool
	CliffInit, CliffDec, ShallowDec float64
	CliffPathLen, ForceSwitch int
	MainType, SubType int 
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
	State GameState
	MenuIndex int
	Loader *LoadingState
	World *WorldMap
	World2 *WorldMap2
	Gen2 *World2Generator
	Dungeon *Dungeon
	Party *Party
	Camera *Camera
	Log []string
	Rng *rand.Rand
	DebugMode bool
	MouseStartX, MouseStartY int
	IsDragging bool
	ArrowTimer float64
	
	TotalRoute1Dist float64
	
	SoilMin     int
	SoilMax     int
	W2Width     int
	W2Height    int
	TransitDist int
	VastOceanSize   int
	IslandBoundSize int
	
	MapRatio    int 
	EnableCentering bool
	
	CliffInitVal   float64
	CliffDecVal    float64
	ShallowDecVal  float64
	CliffPathLen   int
	ForceSwitch    int
	
	LastTargetSoil int
	InputMode int
	InputBuffer string

	WarningMsg   string
	WarningTimer float64

	AutoProgress bool // Enterキー押下時に自動で次のフェーズへ進むフラグ
	SuppressMapDraw bool // 自動進行中はマップ描画を抑制し、Phase名のみ表示
}