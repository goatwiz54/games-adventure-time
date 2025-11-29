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

	// World Map 1 Constants
	WorldWidth    = 100
	WorldHeight   = 80
	WorldTileSize = 32

	// World Map 2 Constants
	World2Width    = 100
	World2Height   = 100
	World2TileSize = 16
)

const (
	DirNorth = 0
	DirEast  = 1
	DirSouth = 2
	DirWest  = 3
)

// World 2 Tile Types
const (
	W2TileVariableOcean = 0 // 可変海
	W2TileSoil          = 1 // 土
	W2TileFixedOcean    = 2 // 固定海
	W2TileTransit       = 3 // 経由島 (New)
)

// Input Edit Modes
const (
	EditNone = 0
	EditSoilMin = 1
	EditSoilMax = 2
	EditW2Width = 3
	EditW2Height = 4
)

var ZoomLevels = []float64{0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5}

// --- ゲーム状態 ---
type GameState int

const (
	StateMenu         GameState = iota
	StateWorldLoading           // World 1 生成
	StateWorld                  // World 1
	StateDungeon                // ダンジョン
	StateWorld2                 // World 2
)

// --- リソース ---
var (
	TexGrass, TexDirt, TexStone, TexWhite, TexArrow    *ebiten.Image
	TexW_MountainIcon, TexW_TreeIcon, TexW_CityIcon    *ebiten.Image
	TexW_Ocean, TexW_Plains, TexW_Forest, TexW_Desert, TexW_Mountain *ebiten.Image
	TexW2_Ocean, TexW2_Soil, TexW2_FixedOcean          *ebiten.Image
)

// --- キャラクター・パーティ ---
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

// --- マップデータ ---
type Tile struct {
	Type, Height int
	Explored     bool
}

type Dungeon struct {
	Width, Height int
	Tiles         [][]Tile
	Enemies       []*Enemy
}

// World 1
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

// World 2
type World2Tile struct {
	Type   int
	IsLake bool
}
type WorldMap2 struct {
	Width, Height    int
	Tiles            [][]World2Tile
	OffsetX, OffsetY float64
	Zoom             float64
	ShowGrid         bool
	StatsInfo        []string
}

// --- システム ---
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
	Dungeon *Dungeon
	Party *Party
	Camera *Camera
	Log []string
	Rng *rand.Rand
	DebugMode bool
	MouseStartX, MouseStartY int
	IsDragging bool
	ArrowTimer float64
	
	// World 2 Generation Settings
	SoilMin   int
	SoilMax   int
	W2Width   int
	W2Height  int
	LastTargetSoil int
	InputMode int
	InputBuffer string

	// Warning
	WarningMsg   string
	WarningTimer float64
}