package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// --- 1. 設定定数 ---
const (
	ScreenWidth   = 1280
	ScreenHeight  = 720
	TileWidth     = 64
	TileHeight    = 32
	BaseTurnToMin = 6
	MoveSpeed     = 0.20

	// World Map Constants
	WorldWidth    = 100
	WorldHeight   = 60
	WorldTileSize = 48
)

const (
	DirNorth = 0
	DirEast  = 1
	DirSouth = 2
	DirWest  = 3
)

// ゲームの状態定義
type GameState int

const (
	StateMenu         GameState = iota
	StateWorldLoading           // ワールド生成中
	StateWorld                  // ワールドマップ操作
	StateDungeon                // ダンジョン探索
)

var ZoomLevels = []float64{0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5}

// --- 2. グローバルリソース ---
var (
	TexGrass, TexDirt, TexStone, TexWhite, TexArrow                  *ebiten.Image
	TexW_Ocean, TexW_Plains, TexW_Forest, TexW_Desert, TexW_Mountain *ebiten.Image
)

func init() {
	TexWhite = ebiten.NewImage(1, 1)
	TexWhite.Fill(color.White)

	// 矢印画像生成
	size := 128
	TexArrow = ebiten.NewImage(size, size)
	orange := color.RGBA{255, 140, 0, 255}

	isInArrow := func(x, y float64) bool {
		if x >= 0 && x <= 64 && y >= 48 && y <= 80 {
			return true
		}
		if x >= 64 && x <= 128 {
			topY := 0.75*x - 32
			botY := -0.75*x + 160
			if y >= topY && y <= botY {
				return true
			}
		}
		return false
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			hits := 0
			for sy := 0.0; sy < 1.0; sy += 0.5 {
				for sx := 0.0; sx < 1.0; sx += 0.5 {
					if isInArrow(float64(x)+sx, float64(y)+sy) {
						hits++
					}
				}
			}
			if hits > 0 {
				c := orange
				c.A = uint8(float64(hits) / 4.0 * 255)
				TexArrow.Set(x, y, c)
			}
		}
	}

	// テクスチャ生成ヘルパー
	genTex := func(baseColor color.RGBA, noiseAmount int, patternType int) *ebiten.Image {
		img := ebiten.NewImage(32, 32)
		pix := make([]byte, 32*32*4)
		rng := rand.New(rand.NewSource(1))
		br, bg, bb := int(baseColor.R), int(baseColor.G), int(baseColor.B)

		for i := 0; i < 32*32; i++ {
			x, y := i%32, i/32
			noise := rng.Intn(noiseAmount*2) - noiseAmount
			if patternType == 1 {
				if rng.Intn(10) == 0 {
					noise -= 20
				}
			} else if patternType == 2 {
				if y%8 == 0 || (y%16 < 8 && x%16 == 0) || (y%16 >= 8 && (x+8)%16 == 0) {
					noise -= 30
				}
			} else {
				if rng.Intn(8) == 0 {
					noise += 15
				}
			}
			if x == 0 || y == 0 {
				noise += 40
			} else if x == 31 || y == 31 {
				noise -= 40
			}

			r, g, b := clamp(br+noise), clamp(bg+noise), clamp(bb+noise)
			pix[4*i], pix[4*i+1], pix[4*i+2], pix[4*i+3] = uint8(r), uint8(g), uint8(b), 255
		}
		img.WritePixels(pix)
		return img
	}

	// ダンジョン用
	TexGrass = genTex(color.RGBA{85, 125, 70, 255}, 10, 0)
	TexDirt = genTex(color.RGBA{100, 80, 60, 255}, 15, 1)
	TexStone = genTex(color.RGBA{100, 100, 110, 255}, 10, 2)

	// ワールドマップ用
	fillImg := func(c color.Color) *ebiten.Image {
		img := ebiten.NewImage(32, 32)
		img.Fill(c)
		return img
	}
	TexW_Ocean = fillImg(color.RGBA{20, 60, 120, 255})
	TexW_Plains = fillImg(color.RGBA{100, 160, 80, 255})
	TexW_Forest = fillImg(color.RGBA{40, 100, 50, 255})
	TexW_Desert = fillImg(color.RGBA{200, 180, 100, 255})
	TexW_Mountain = fillImg(color.RGBA{120, 110, 100, 255})
}

func clamp(x int) int {
	if x < 0 {
		return 0
	}
	if x > 255 {
		return 255
	}
	return x
}

// --- 3. データ構造 ---

type Status struct {
	HP, SP, STR, DEX, VIT, INT, AGI, SPD, LUCK int
}

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
	Biome  int // 0:Ocean, 1:Plains, 2:Forest, 3:Desert, 4:Mountain
	Height float64
}

type WorldMap struct {
	Tiles            [][]WorldTile
	CameraX, CameraY float64
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
	Dungeon                  *Dungeon
	Party                    *Party
	Camera                   *Camera
	Log                      []string
	Rng                      *rand.Rand
	DebugMode                bool
	MouseStartX, MouseStartY int
	IsDragging               bool
	ArrowTimer               float64
}

// --- 4. ロジック ---

func (c *Character) CalculateGameWT() int {
	weightPenalty := int(math.Ceil(c.LoadWeight))
	agiBonus := int(math.Floor(float64(c.Stats.AGI) * 2.5))
	intBonus := c.Stats.INT * 2
	gameWT := c.BaseWT + weightPenalty - agiBonus - intBonus
	if gameWT < 1 {
		return 1
	}
	return gameWT
}

func (c *Character) MaxLoadWeight() float64 { return float64(c.Stats.STR * 5) }

func (c *Character) WeightPenalty() float64 {
	max := c.MaxLoadWeight()
	if max == 0 {
		return 2.0
	}
	ratio := c.LoadWeight / max
	if ratio < 0.85 {
		return 1.0
	}
	if ratio < 0.90 {
		return 1.1
	}
	if ratio < 0.95 {
		return 1.2
	}
	if ratio < 1.00 {
		return 1.4
	}
	return 2.0
}

func (p *Party) ExplorationWT() int {
	if len(p.Members) == 0 {
		return 100
	}
	total := 0
	for _, m := range p.Members {
		total += m.CalculateGameWT()
	}
	return int(math.Ceil(float64(total) / float64(len(p.Members))))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *Game) GetVisibility(tx, ty int) int {
	if g.DebugMode {
		return 0
	}
	leader := g.Party.Leader
	px, py := leader.TargetX, leader.TargetY
	dist := int(math.Abs(float64(tx-px)) + math.Abs(float64(ty-py)))
	dx, dy := tx-px, ty-py
	isFront, isBack, isSide := false, false, false

	switch leader.Facing {
	case DirNorth:
		if dy < 0 && abs(dx) <= abs(dy) {
			isFront = true
		} else if dy > 0 && abs(dx) <= abs(dy) {
			isBack = true
		} else {
			isSide = true
		}
	case DirSouth:
		if dy > 0 && abs(dx) <= abs(dy) {
			isFront = true
		} else if dy < 0 && abs(dx) <= abs(dy) {
			isBack = true
		} else {
			isSide = true
		}
	case DirWest:
		if dx < 0 && abs(dy) <= abs(dx) {
			isFront = true
		} else if dx > 0 && abs(dy) <= abs(dx) {
			isBack = true
		} else {
			isSide = true
		}
	case DirEast:
		if dx > 0 && abs(dy) <= abs(dx) {
			isFront = true
		} else if dx < 0 && abs(dy) <= abs(dx) {
			isBack = true
		} else {
			isSide = true
		}
	}

	effectiveDist := dist
	if isFront {
		effectiveDist -= 2
	} else if isSide {
		effectiveDist -= 1
	} else if isBack {
		effectiveDist += 1
	}
	if effectiveDist <= 2 {
		return 0
	}
	if effectiveDist <= 3 {
		return 1
	}
	if effectiveDist <= 4 {
		return 2
	}
	return 3
}

func (e *Enemy) CheckPlayerVisibility(px, py int) int {
	dist := abs(e.TargetX-px) + abs(e.TargetY-py)
	dx, dy := px-e.TargetX, py-e.TargetY
	isFront, isBack, isSide := false, false, false
	switch e.Facing {
	case DirNorth:
		if dy < 0 && abs(dx) <= abs(dy) {
			isFront = true
		} else if dy > 0 {
			isBack = true
		} else {
			isSide = true
		}
	case DirSouth:
		if dy > 0 && abs(dx) <= abs(dy) {
			isFront = true
		} else if dy < 0 {
			isBack = true
		} else {
			isSide = true
		}
	case DirWest:
		if dx < 0 && abs(dy) <= abs(dx) {
			isFront = true
		} else if dx > 0 {
			isBack = true
		} else {
			isSide = true
		}
	case DirEast:
		if dx > 0 && abs(dy) <= abs(dx) {
			isFront = true
		} else if dx < 0 {
			isBack = true
		} else {
			isSide = true
		}
	}
	ed := dist
	if isFront {
		ed -= 2
	} else if isSide {
		ed -= 1
	} else if isBack {
		ed += 1
	}
	if ed <= 2 {
		return 0
	}
	if ed <= 3 {
		return 1
	}
	return 3
}

// --- 5. 初期化と生成ロジック ---

func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	g := &Game{
		State:     StateMenu,
		MenuIndex: 0,
		Rng:       rng,
	}
	g.InitDungeon()
	return g
}

func (g *Game) StartWorldLoading() {
	g.State = StateWorldLoading
	g.Loader = &LoadingState{
		StartTime: time.Now(),
		Progress:  0.0,
		Step:      0,
		Logs:      []LoadingLog{},
	}
	g.AddLoadingLog("Starting World Generation...")
}

func (g *Game) AddLoadingLog(msg string) {
	g.Loader.Logs = append(g.Loader.Logs, LoadingLog{Msg: msg, AddedAt: time.Now()})
}

// ローディング中の更新処理 (3秒かけてステップを進める)
func (g *Game) UpdateLoading() {
	elapsed := time.Since(g.Loader.StartTime).Seconds()

	g.Loader.Progress = elapsed / 3.0
	if g.Loader.Progress > 1.0 {
		g.Loader.Progress = 1.0
	}

	// 0.5s: 構造生成
	if elapsed > 0.5 && g.Loader.Step == 0 {
		g.AddLoadingLog("Generating Map Structure...")
		g.World = &WorldMap{
			Tiles:   make([][]WorldTile, WorldWidth),
			CameraX: float64(WorldWidth * WorldTileSize / 2),
			CameraY: float64(WorldHeight * WorldTileSize / 2),
		}
		for x := 0; x < WorldWidth; x++ {
			g.World.Tiles[x] = make([]WorldTile, WorldHeight)
		}
		g.Loader.Step++
	}
	// 1.0s: 地形生成
	if elapsed > 1.0 && g.Loader.Step == 1 {
		g.AddLoadingLog("Generating Terrain Data...")
		for x := 0; x < WorldWidth; x++ {
			for y := 0; y < WorldHeight; y++ {
				// ノイズ計算
				nx := float64(x) * 0.1
				ny := float64(y) * 0.1
				val := math.Sin(nx)*math.Cos(ny) + math.Sin(nx*2.5)*0.5

				biome := 0 // Ocean
				if val > 0.6 {
					biome = 4 // Mountain
				} else if val > 0.3 {
					biome = 2 // Forest
				} else if val > 0.0 {
					biome = 1 // Plains
				} else if val > -0.2 {
					biome = 3 // Desert
				}

				g.World.Tiles[x][y] = WorldTile{Biome: biome, Height: val}
			}
		}
		g.Loader.Step++
	}
	// 1.5s: サンガード
	if elapsed > 1.5 && g.Loader.Step == 2 {
		g.AddLoadingLog("Sunguard... Creating...")
		g.Loader.Step++
	}
	// 2.0s: シルバーフォレスト
	if elapsed > 2.0 && g.Loader.Step == 3 {
		g.AddLoadingLog("Silver Forest... Creating...")
		g.Loader.Step++
	}
	// 2.5s: 仕上げ
	if elapsed > 2.5 && g.Loader.Step == 4 {
		g.AddLoadingLog("Finalizing Map...")
		g.Loader.Step++
	}

	// 3.0s: 完了
	if elapsed > 3.0 {
		g.State = StateWorld
	}
}

func (g *Game) InitDungeon() {
	w, h := 40, 40
	d := GenerateDungeon(w, h, g.Rng)
	leader := &Character{Name: "Denim", Stats: Status{AGI: 14}, BaseWT: 290, LoadWeight: 12.0, Facing: DirSouth}
	px, py := 0, 0
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if d.Tiles[x][y].Type == 1 {
				px, py = x, y
				goto Found
			}
		}
	}
Found:
	leader.TargetX, leader.TargetY = px, py
	leader.TargetZ = d.Tiles[px][py].Height
	leader.CurrentX, leader.CurrentY = float64(px), float64(py)
	leader.CurrentZ = float64(leader.TargetZ)
	g.Dungeon = d
	g.Party = &Party{Leader: leader, Members: []*Character{leader}}
	g.Camera = &Camera{ZoomIndex: 3}
	g.Log = []string{"Quest Started."}
	g.ArrowTimer = 0.3
	g.CenterCamera()
}

func GenerateDungeon(w, h int, rng *rand.Rand) *Dungeon {
	d := &Dungeon{Width: w, Height: h, Tiles: make([][]Tile, w)}
	for x := 0; x < w; x++ {
		d.Tiles[x] = make([]Tile, h)
		for y := 0; y < h; y++ {
			d.Tiles[x][y] = Tile{Type: 0, Height: 0, Explored: false}
		}
	}
	var rooms []struct{ x, y, w, h, z int }
	count := 12 + rng.Intn(6)
	for i := 0; i < count; i++ {
		rw, rh := 4+rng.Intn(4), 4+rng.Intn(4)
		rx, ry := 1+rng.Intn(w-rw-2), 1+rng.Intn(h-rh-2)
		rz := rng.Intn(4)
		if len(rooms) > 0 {
			rz = rooms[len(rooms)-1].z + rng.Intn(3) - 1
			if rz < 0 {
				rz = 0
			}
			if rz > 5 {
				rz = 5
			}
		}
		for x := rx; x < rx+rw; x++ {
			for y := ry; y < ry+rh; y++ {
				d.Tiles[x][y] = Tile{Type: 1, Height: rz}
			}
		}
		if len(rooms) > 0 {
			prev := rooms[len(rooms)-1]
			cx1, cy1 := prev.x+prev.w/2, prev.y+prev.h/2
			cx2, cy2 := rx+rw/2, ry+rh/2
			pathZ := prev.z
			if rz < pathZ {
				pathZ = rz
			}
			sx, ex := cx1, cx2
			if sx > ex {
				sx, ex = ex, sx
			}
			for x := sx; x <= ex; x++ {
				d.Tiles[x][cy1] = Tile{Type: 1, Height: pathZ}
			}
			sy, ey := cy1, cy2
			if sy > ey {
				sy, ey = ey, sy
			}
			for y := sy; y <= ey; y++ {
				d.Tiles[cx2][y] = Tile{Type: 1, Height: pathZ}
			}
		}
		rooms = append(rooms, struct{ x, y, w, h, z int }{rx, ry, rw, rh, rz})
	}
	enemyCount := 10 + rng.Intn(5)
	for i := 0; i < enemyCount; i++ {
		for attempt := 0; attempt < 100; attempt++ {
			ex := rng.Intn(w)
			ey := rng.Intn(h)
			if d.Tiles[ex][ey].Type == 1 {
				ez := d.Tiles[ex][ey].Height
				etype, spd := 1, 1
				if rng.Float64() < 0.3 {
					etype = 2
					spd = 2
				}
				d.Enemies = append(d.Enemies, &Enemy{
					ID: i, TargetX: ex, TargetY: ey, TargetZ: ez,
					CurrentX: float64(ex), CurrentY: float64(ey), CurrentZ: float64(ez),
					Type: etype, Speed: spd, Active: true, Facing: DirWest,
				})
				break
			}
		}
	}
	return d
}

// --- 6. 更新ループ ---

func (g *Game) Update() error {
	switch g.State {
	case StateMenu:
		return g.UpdateMenu()
	case StateWorldLoading:
		g.UpdateLoading()
		return nil
	case StateWorld:
		return g.UpdateWorld()
	case StateDungeon:
		return g.UpdateDungeon()
	}
	return nil
}

func (g *Game) UpdateMenu() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.MenuIndex--
		if g.MenuIndex < 0 {
			g.MenuIndex = 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.MenuIndex++
		if g.MenuIndex > 1 {
			g.MenuIndex = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if g.MenuIndex == 0 {
			g.StartWorldLoading()
		} else {
			g.State = StateDungeon
		}
	}
	return nil
}

func (g *Game) UpdateWorld() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.State = StateMenu
	}

	moveSpd := 10.0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.World.CameraX -= moveSpd
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.World.CameraX += moveSpd
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.World.CameraY -= moveSpd
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.World.CameraY += moveSpd
	}

	worldPixelW := float64(WorldWidth * WorldTileSize)
	if g.World.CameraX < 0 {
		g.World.CameraX += worldPixelW
	}
	if g.World.CameraX >= worldPixelW {
		g.World.CameraX -= worldPixelW
	}

	worldPixelH := float64(WorldHeight * WorldTileSize)
	if g.World.CameraY < 0 {
		g.World.CameraY = 0
	}
	if g.World.CameraY > worldPixelH {
		g.World.CameraY = worldPixelH
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		g.DebugMode = !g.DebugMode
	}

	return nil
}

func (g *Game) refreshArrow() {
	if g.ArrowTimer >= 8.0 {
		g.ArrowTimer = 0.0
	} else {
		g.ArrowTimer = 0.3
	}
}

func (g *Game) UpdateDungeon() error {
	g.ArrowTimer += 1.0 / 60.0
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.State = StateMenu
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.DebugMode = !g.DebugMode
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		g.refreshArrow()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			g.Camera.ZoomIndex--
			if g.Camera.ZoomIndex < 0 {
				g.Camera.ZoomIndex = len(ZoomLevels) - 1
			}
		} else {
			g.Camera.ZoomIndex++
			if g.Camera.ZoomIndex >= len(ZoomLevels) {
				g.Camera.ZoomIndex = 0
			}
		}
		g.CenterCamera()
	}
	cameraChanged := false
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			g.Camera.TargetAngle -= math.Pi / 2
			cameraChanged = true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			g.Camera.TargetAngle += math.Pi / 2
			cameraChanged = true
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		cx, cy := ebiten.CursorPosition()
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.IsDragging = true
			g.MouseStartX, g.MouseStartY = cx, cy
		}
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			g.IsDragging = false
		}
		if g.IsDragging {
			dx := float64(cx - g.MouseStartX)
			dy := float64(cy - g.MouseStartY)
			if math.Abs(dx) > 0 || math.Abs(dy) > 0 {
				cameraChanged = true
			}
			g.Camera.X += dx
			g.Camera.Y += dy
			g.MouseStartX, g.MouseStartY = cx, cy
		}
	} else {
		g.IsDragging = false
	}
	if cameraChanged {
		g.refreshArrow()
	}
	diff := g.Camera.TargetAngle - g.Camera.Angle
	if math.Abs(diff) > 0.01 {
		g.Camera.Angle += diff * 0.1
		g.CenterCamera()
	} else {
		g.Camera.Angle = g.Camera.TargetAngle
	}

	animate := func(curr *float64, target int) bool {
		d := float64(target) - *curr
		if math.Abs(d) < 0.05 {
			*curr = float64(target)
			return false
		}
		*curr += d * MoveSpeed
		return true
	}
	leader := g.Party.Leader
	imX := animate(&leader.CurrentX, leader.TargetX)
	imY := animate(&leader.CurrentY, leader.TargetY)
	imZ := animate(&leader.CurrentZ, leader.TargetZ)
	leader.IsMoving = imX || imY || imZ
	anyEnemyMoving := false
	for _, e := range g.Dungeon.Enemies {
		if !e.Active {
			continue
		}
		emX := animate(&e.CurrentX, e.TargetX)
		emY := animate(&e.CurrentY, e.TargetY)
		emZ := animate(&e.CurrentZ, e.TargetZ)
		e.IsMoving = emX || emY || emZ
		if e.IsMoving {
			anyEnemyMoving = true
		}
	}
	if leader.IsMoving || anyEnemyMoving || g.Party.InCombat {
		if leader.IsMoving && !g.IsDragging {
			g.CenterCamera()
		}
		if g.Party.InCombat && inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.Party.InCombat = false
			g.Log = append(g.Log, "Combat Reset.")
			for _, e := range g.Dungeon.Enemies {
				if abs(leader.TargetX-e.TargetX)+abs(leader.TargetY-e.TargetY) <= 1 {
					e.Active = false
				}
			}
		}
		return nil
	}
	inputDx, inputDy := 0, 0
	pressed := false
	if !ebiten.IsKeyPressed(ebiten.KeyControl) {
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			inputDy = -1
			pressed = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			inputDy = 1
			pressed = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			inputDx = -1
			pressed = true
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			inputDx = 1
			pressed = true
		}
	}
	if pressed {
		g.refreshArrow()
		snapAngle := math.Round(g.Camera.TargetAngle/(math.Pi/2)) * (math.Pi / 2)
		cos := math.Cos(-snapAngle)
		sin := math.Sin(-snapAngle)
		fdx := float64(inputDx)*cos - float64(inputDy)*sin
		fdy := float64(inputDx)*sin + float64(inputDy)*cos
		dx := int(math.Round(fdx))
		dy := int(math.Round(fdy))
		newFacing := leader.Facing
		if dx == 0 && dy == -1 {
			newFacing = DirNorth
		}
		if dx == 0 && dy == 1 {
			newFacing = DirSouth
		}
		if dx == -1 && dy == 0 {
			newFacing = DirWest
		}
		if dx == 1 && dy == 0 {
			newFacing = DirEast
		}
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if leader.Facing != newFacing {
				leader.Facing = newFacing
				wt := int(math.Ceil(float64(g.Party.ExplorationWT()) * 0.1 * leader.WeightPenalty()))
				if wt < 1 {
					wt = 1
				}
				g.Party.TotalTurns += wt
			}
			return nil
		}
		nx, ny := leader.TargetX+dx, leader.TargetY+dy
		if nx >= 0 && nx < g.Dungeon.Width && ny >= 0 && ny < g.Dungeon.Height {
			tile := g.Dungeon.Tiles[nx][ny]
			curTile := g.Dungeon.Tiles[leader.TargetX][leader.TargetY]
			if tile.Type == 1 && abs(tile.Height-curTile.Height) <= 1 {
				leader.TargetX = nx
				leader.TargetY = ny
				leader.TargetZ = tile.Height
				leader.Facing = newFacing
				leader.IsMoving = true
				baseWT := g.Party.ExplorationWT()
				penalty := leader.WeightPenalty()
				cost := int(math.Ceil(float64(baseWT) * penalty))
				g.Party.TotalTurns += cost
				for _, e := range g.Dungeon.Enemies {
					if e.Active && abs(leader.TargetX-e.TargetX)+abs(leader.TargetY-e.TargetY) <= 1 {
						g.StartCombat(e)
						return nil
					}
				}
				g.ProcessEnemyTurn()
			}
		}
	}
	return nil
}

func (g *Game) ProcessEnemyTurn() {
	leader := g.Party.Leader
	px, py := leader.TargetX, leader.TargetY
	for _, e := range g.Dungeon.Enemies {
		if !e.Active {
			continue
		}
		for s := 0; s < e.Speed; s++ {
			vis := e.CheckPlayerVisibility(px, py)
			dist := abs(px-e.TargetX) + abs(py-e.TargetY)
			action := 2
			r := g.Rng.Intn(100)
			if vis == 0 {
				if dist <= 2 {
					if r < 55 {
						action = 0
					} else if r < 85 {
						action = 1
					} else {
						action = 2
					}
					if action == 2 && g.Rng.Intn(2) == 0 {
						r2 := g.Rng.Intn(100)
						if r2 < 55 {
							action = 0
						} else if r2 < 85 {
							action = 1
						}
					}
				} else {
					if r < 30 {
						action = 0
					} else {
						action = 2
					}
				}
			} else if vis == 1 {
				if r < 30 {
					action = 0
				} else if r < 45 {
					action = 1
				} else {
					action = 2
				}
			}
			tx, ty := e.TargetX, e.TargetY
			if action == 0 {
				if px > tx {
					tx++
				} else if px < tx {
					tx--
				}
				if py > ty {
					ty++
				} else if py < ty {
					ty--
				}
				if abs(px-e.TargetX) > abs(py-e.TargetY) {
					ty = e.TargetY
				} else {
					tx = e.TargetX
				}
			} else if action == 2 {
				dir := g.Rng.Intn(4)
				switch dir {
				case 0:
					ty--
				case 1:
					ty++
				case 2:
					tx--
				case 3:
					tx++
				}
			}
			if tx >= 0 && tx < g.Dungeon.Width && ty >= 0 && ty < g.Dungeon.Height {
				nt := g.Dungeon.Tiles[tx][ty]
				ct := g.Dungeon.Tiles[e.TargetX][e.TargetY]
				if nt.Type == 1 && abs(nt.Height-ct.Height) <= 1 {
					if tx > e.TargetX {
						e.Facing = DirEast
					}
					if tx < e.TargetX {
						e.Facing = DirWest
					}
					if ty > e.TargetY {
						e.Facing = DirSouth
					}
					if ty < e.TargetY {
						e.Facing = DirNorth
					}
					e.TargetX = tx
					e.TargetY = ty
					e.TargetZ = nt.Height
					e.IsMoving = true
				}
			}
			if abs(px-e.TargetX)+abs(py-e.TargetY) <= 1 {
				g.StartCombat(e)
				break
			}
		}
	}
}

func (g *Game) StartCombat(e *Enemy) {
	g.Party.InCombat = true
	g.Party.CombatLog = "ENCOUNTER! Press [R] to Reset."
	g.Log = append(g.Log, "Battle Started!")
}

// --- 7. 描画 ---

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.State {
	case StateMenu:
		g.DrawMenu(screen)
	case StateWorldLoading:
		g.DrawLoading(screen)
	case StateWorld:
		g.DrawWorld(screen)
	case StateDungeon:
		g.DrawDungeon(screen)
	}
}

func (g *Game) DrawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 20, 255})
	title := "TACTICS DUNGEON RPG"
	text.Draw(screen, title, basicfont.Face7x13, ScreenWidth/2-80, ScreenHeight/3, color.White)
	var c1, c2 color.Color = color.White, color.White
	if g.MenuIndex == 0 {
		c1 = color.RGBA{255, 200, 0, 255}
	}
	if g.MenuIndex == 1 {
		c2 = color.RGBA{255, 200, 0, 255}
	}
	text.Draw(screen, "> 1. World Map", basicfont.Face7x13, ScreenWidth/2-60, ScreenHeight/2, c1)
	text.Draw(screen, "> 2. Dungeon", basicfont.Face7x13, ScreenWidth/2-60, ScreenHeight/2+30, c2)
}

func (g *Game) DrawLoading(screen *ebiten.Image) {
	screen.Fill(color.Black)
	barW, barH := 400.0, 10.0
	x, y := float64(ScreenWidth)/2-barW/2, float64(ScreenHeight)-100.0
	ebitenutil.DrawRect(screen, x, y, barW, barH, color.RGBA{50, 50, 50, 255})
	ebitenutil.DrawRect(screen, x, y, barW*g.Loader.Progress, barH, color.RGBA{0, 200, 0, 255})

	logCount := len(g.Loader.Logs)
	showCount := 0
	for i := logCount - 1; i >= 0; i-- {
		l := g.Loader.Logs[i]
		elapsed := time.Since(l.AddedAt).Seconds()
		alpha := 1.0
		if elapsed > 1.5 {
			alpha = 1.0 - (elapsed-1.5)*2
		}
		if alpha < 0 {
			continue
		}
		if showCount >= 3 {
			break
		}
		c := color.RGBA{200, 200, 200, uint8(255 * alpha)}
		py := y - 30.0 - float64(showCount)*20.0
		text.Draw(screen, l.Msg, basicfont.Face7x13, int(x), int(py), c)
		showCount++
	}
}

func (g *Game) DrawWorld(screen *ebiten.Image) {
	screen.Fill(color.RGBA{5, 5, 20, 255})
	curveFactor := 0.00005
	rangeX := ScreenWidth/WorldTileSize/2 + 2
	rangeY := ScreenHeight/WorldTileSize/2 + 4
	camX, camY := g.World.CameraX, g.World.CameraY

	for y := -rangeY; y <= rangeY; y++ {
		for x := -rangeX; x <= rangeX; x++ {
			wx := (int(camX)/WorldTileSize + x)
			wy := (int(camY)/WorldTileSize + y)
			if wy < 0 || wy >= WorldHeight {
				continue
			}
			normWX := (wx%WorldWidth + WorldWidth) % WorldWidth
			tile := g.World.Tiles[normWX][wy]
			sx := float64(x*WorldTileSize) + (float64(int(camX)%WorldTileSize) * -1) + ScreenWidth/2
			sy := float64(y*WorldTileSize) + (float64(int(camY)%WorldTileSize) * -1) + ScreenHeight/2
			dx := sx - ScreenWidth/2
			dy := sy - ScreenHeight/2
			distSq := dx*dx + dy*dy
			sy += distSq * curveFactor
			var img *ebiten.Image
			switch tile.Biome {
			case 0:
				img = TexW_Ocean
			case 1:
				img = TexW_Plains
			case 2:
				img = TexW_Forest
			case 3:
				img = TexW_Desert
			case 4:
				img = TexW_Mountain
			}
			op := &ebiten.DrawImageOptions{}
			scale := float64(WorldTileSize) / 32.0
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(sx, sy)
			if sx < -WorldTileSize || sx > ScreenWidth || sy < -WorldTileSize || sy > ScreenHeight {
				continue
			}
			screen.DrawImage(img, op)
			if g.DebugMode {
				msg := fmt.Sprintf("%d", tile.Biome)
				text.Draw(screen, msg, basicfont.Face7x13, int(sx), int(sy), color.White)
			}
		}
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("World Map (Arrow Keys to Scroll)\nPos: %.0f, %.0f", camX, camY))
}

func (g *Game) DrawDungeon(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 25, 30, 255})
	s := g.GetScale()
	ang := g.Camera.Angle

	type RenderItem struct {
		Type  int
		Depth float64
		Obj   interface{}
		X, Y  int
	}
	var items []RenderItem

	getSortDepth := func(x, y float64) float64 {
		_, sy := IsoToScreen(x, y, 0, s, ang)
		return sy
	}

	for x := 0; x < g.Dungeon.Width; x++ {
		for y := 0; y < g.Dungeon.Height; y++ {
			t := &g.Dungeon.Tiles[x][y]
			if t.Type == 0 {
				continue
			}
			vis := g.GetVisibility(x, y)
			if vis < 3 && !g.DebugMode {
				t.Explored = true
			}
			if !t.Explored && !g.DebugMode {
				continue
			}
			items = append(items, RenderItem{Type: 0, Depth: getSortDepth(float64(x), float64(y)), Obj: t, X: x, Y: y})
		}
	}

	leader := g.Party.Leader
	charDepth := getSortDepth(leader.CurrentX, leader.CurrentY) + float64(TileHeight/2)*s
	items = append(items, RenderItem{Type: 1, Depth: charDepth, Obj: leader})

	for _, e := range g.Dungeon.Enemies {
		if !e.Active {
			continue
		}
		vis := g.GetVisibility(int(math.Round(e.CurrentX)), int(math.Round(e.CurrentY)))
		if vis < 3 || g.DebugMode {
			eDepth := getSortDepth(e.CurrentX, e.CurrentY) + float64(TileHeight/2)*s
			items = append(items, RenderItem{Type: 2, Depth: eDepth, Obj: e})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if math.Abs(items[i].Depth-items[j].Depth) > 0.1 {
			return items[i].Depth < items[j].Depth
		}
		return items[i].Type < items[j].Type
	})

	arrowAlpha := 0.0
	if g.ArrowTimer < 0.3 {
		arrowAlpha = g.ArrowTimer / 0.3
	} else if g.ArrowTimer < 5.0 {
		arrowAlpha = 1.0
	} else if g.ArrowTimer < 8.0 {
		arrowAlpha = 1.0 - (g.ArrowTimer-5.0)/3.0
	}

	for _, item := range items {
		switch item.Type {
		case 0:
			t := item.Obj.(*Tile)
			sx, sy := IsoToScreen(float64(item.X), float64(item.Y), float64(t.Height), s, ang)
			sx += g.Camera.X
			sy += g.Camera.Y
			vis := g.GetVisibility(item.X, item.Y)
			drawTexturedBlock(screen, sx, sy, t.Height, s)
			if vis == 3 && t.Explored {
				drawOverlay(screen, sx, sy, t.Height, s, color.RGBA{0, 0, 0, 180})
			} else if vis > 0 {
				alpha := uint8(0)
				if vis == 1 {
					alpha = 80
				}
				if vis == 2 {
					alpha = 160
				}
				drawOverlay(screen, sx, sy, t.Height, s, color.RGBA{0, 0, 0, alpha})
			}
		case 1:
			p := item.Obj.(*Character)
			sx, sy := IsoToScreen(p.CurrentX, p.CurrentY, p.CurrentZ, s, ang)
			drawUnit(screen, sx+g.Camera.X, sy+g.Camera.Y, s, color.RGBA{50, 100, 255, 255}, p.Facing, g.Camera.Angle, true, arrowAlpha)
		case 2:
			e := item.Obj.(*Enemy)
			c := color.RGBA{220, 50, 50, 255}
			if e.Type == 2 {
				c = color.RGBA{220, 100, 100, 255}
			}
			sx, sy := IsoToScreen(e.CurrentX, e.CurrentY, e.CurrentZ, s, ang)
			pcVis := g.GetVisibility(int(math.Round(e.CurrentX)), int(math.Round(e.CurrentY)))
			drawUnit(screen, sx+g.Camera.X, sy+g.Camera.Y, s, c, e.Facing, g.Camera.Angle, pcVis <= 1, arrowAlpha)
		}
	}
	g.DrawDungeonUI(screen)
}

// ----------------------------------------------------------------------
// Missing Helper Functions (Added here)
// ----------------------------------------------------------------------

func (g *Game) GetScale() float64 {
	return ZoomLevels[g.Camera.ZoomIndex] / 0.8
}

func (g *Game) CenterCamera() {
	// Only center for Dungeon state in this version
	if g.State == StateDungeon {
		leader := g.Party.Leader
		s := g.GetScale()
		isoX, isoY := IsoToScreen(leader.CurrentX, leader.CurrentY, leader.CurrentZ, s, g.Camera.Angle)
		g.Camera.X = -(isoX - ScreenWidth/2)
		g.Camera.Y = -(isoY - ScreenHeight/2)
	}
}

func IsoToScreen(x, y, z float64, scale float64, angle float64) (float64, float64) {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	rx := x*cos - y*sin
	ry := x*sin + y*cos
	tw := float64(TileWidth) * scale
	th := float64(TileHeight) * scale
	sx := (rx - ry) * (tw / 2.0)
	sy := (rx + ry) * (th / 2.0)
	sy -= z * 16.0 * scale
	return sx, sy
}

func drawTexturedBlock(screen *ebiten.Image, x, y float64, h int, scale float64) {
	w := float32(float64(TileWidth) * scale)
	hh := float32(float64(TileHeight) * scale)
	cx, cy := float32(x), float32(y)
	depth := float32(float64(h)*16.0*scale + 10.0*scale)
	drawTexturedQuad(screen, TexDirt, cx, cy+hh, cx+w/2, cy+hh/2, cx+w/2, cy+hh/2+depth, cx, cy+hh+depth, color.RGBA{120, 120, 120, 255})
	drawTexturedQuad(screen, TexDirt, cx, cy+hh, cx-w/2, cy+hh/2, cx-w/2, cy+hh/2+depth, cx, cy+hh+depth, color.RGBA{200, 200, 200, 255})
	topTex := TexGrass
	if h == 0 {
		topTex = TexStone
	}
	drawTexturedQuad(screen, topTex, cx, cy, cx+w/2, cy+hh/2, cx, cy+hh, cx-w/2, cy+hh/2, color.White)
}

func drawTexturedQuad(screen *ebiten.Image, tex *ebiten.Image, x1, y1, x2, y2, x3, y3, x4, y4 float32, clr color.Color) {
	r, g, b, a := clr.RGBA()
	cr, cg, cb, ca := float32(r)/65535, float32(g)/65535, float32(b)/65535, float32(a)/65535
	w, h := float32(tex.Bounds().Dx()), float32(tex.Bounds().Dy())
	var vs []ebiten.Vertex
	vs = append(vs, ebiten.Vertex{DstX: x1, DstY: y1, SrcX: 0, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca})
	vs = append(vs, ebiten.Vertex{DstX: x2, DstY: y2, SrcX: w, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca})
	vs = append(vs, ebiten.Vertex{DstX: x3, DstY: y3, SrcX: w, SrcY: h, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca})
	vs = append(vs, ebiten.Vertex{DstX: x4, DstY: y4, SrcX: 0, SrcY: h, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca})
	is := []uint16{0, 1, 2, 2, 3, 0}
	screen.DrawTriangles(vs, is, tex, nil)
}

func drawOverlay(screen *ebiten.Image, x, y float64, h int, scale float64, c color.Color) {
	w := float32(float64(TileWidth) * scale)
	hh := float32(float64(TileHeight) * scale)
	cx, cy := float32(x), float32(y)
	drawTexturedQuad(screen, TexWhite, cx, cy, cx+w/2, cy+hh/2, cx, cy+hh, cx-w/2, cy+hh/2, c)
}

func drawUnit(screen *ebiten.Image, x, y float64, scale float64, c color.Color, facing int, camAngle float64, showArrow bool, arrowAlpha float64) {
	s := float32(scale)
	ebitenutil.DrawRect(screen, x-6*float64(s), y+14*float64(s), 12*float64(s), 4*float64(s), color.RGBA{0, 0, 0, 100})
	ebitenutil.DrawRect(screen, x-4*float64(s), y-14*float64(s), 8*float64(s), 22*float64(s), c)

	if showArrow && arrowAlpha > 0 {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-64, -64)
		baseAng := (315.0 + float64(facing)*90.0) * math.Pi / 180
		finalRad := baseAng + camAngle
		op.GeoM.Rotate(finalRad)
		op.GeoM.Scale(1.0, 0.5)
		op.GeoM.Scale(float64(s), float64(s))
		op.GeoM.Scale(0.35, 0.35)
		op.GeoM.Translate(x, y-28*float64(s))
		op.ColorScale.ScaleAlpha(float32(arrowAlpha))
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(TexArrow, op)
	}
}

func (g *Game) DrawDungeonUI(screen *ebiten.Image) {
	if g.Party.InCombat {
		ebitenutil.DrawRect(screen, 0, ScreenHeight/2-40, ScreenWidth, 80, color.RGBA{150, 0, 0, 200})
		text.Draw(screen, g.Party.CombatLog, basicfont.Face7x13, ScreenWidth/2-100, ScreenHeight/2+5, color.White)
		return
	}
	totalMins := g.Party.TotalTurns * BaseTurnToMin
	zoomVal := ZoomLevels[g.Camera.ZoomIndex]
	uiStr := fmt.Sprintf("Zoom: %.1fx [F12]  Time: %d min\n[F3]: Show Arrow  [ESC]: Menu", zoomVal, totalMins)
	ebitenutil.DrawRect(screen, 0, 0, 200, 60, color.RGBA{0, 0, 0, 180})
	text.Draw(screen, uiStr, basicfont.Face7x13, 10, 20, color.White)

	cx, cy := float64(ScreenWidth)-60, float64(ScreenHeight)-60
	dirs := []string{"N", "E", "S", "W"}
	radius := 40.0
	for i, d := range dirs {
		baseAng := (315.0 + float64(i)*90.0) * math.Pi / 180
		drawAng := baseAng + g.Camera.Angle
		dx := math.Cos(drawAng) * radius
		dy := math.Sin(drawAng) * radius
		text.Draw(screen, d, basicfont.Face7x13, int(cx+dx)-4, int(cy+dy)+4, color.White)
	}

	for i, l := range g.Log {
		text.Draw(screen, l, basicfont.Face7x13, 10, ScreenHeight-100+(i*15), color.RGBA{220, 220, 220, 255})
	}
}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Tactics Dungeon: World Fixed")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
