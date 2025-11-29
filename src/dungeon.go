// filename: dungeon.go
package main

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (c *Character) CalculateGameWT() int { weightPenalty := int(math.Ceil(c.LoadWeight)); agiBonus := int(math.Floor(float64(c.Stats.AGI) * 2.5)); intBonus := c.Stats.INT * 2; gameWT := c.BaseWT + weightPenalty - agiBonus - intBonus; if gameWT < 1 { return 1 }; return gameWT }
func (c *Character) MaxLoadWeight() float64 { return float64(c.Stats.STR * 5) }
func (c *Character) WeightPenalty() float64 { max := c.MaxLoadWeight(); if max == 0 { return 2.0 }; ratio := c.LoadWeight / max; if ratio < 0.85 { return 1.0 }; if ratio < 0.90 { return 1.1 }; if ratio < 0.95 { return 1.2 }; if ratio < 1.00 { return 1.4 }; return 2.0 }
func (p *Party) ExplorationWT() int { if len(p.Members) == 0 { return 100 }; total := 0; for _, m := range p.Members { total += m.CalculateGameWT() }; return int(math.Ceil(float64(total) / float64(len(p.Members)))) }
func abs(x int) int { if x < 0 { return -x }; return x }
func (g *Game) GetVisibility(tx, ty int) int { if g.DebugMode { return 0 }; leader := g.Party.Leader; px, py := leader.TargetX, leader.TargetY; dist := int(math.Abs(float64(tx-px)) + math.Abs(float64(ty-py))); dx, dy := tx-px, ty-py; isFront, isBack, isSide := false, false, false; switch leader.Facing { case DirNorth: if dy < 0 && abs(dx) <= abs(dy) { isFront = true } else if dy > 0 && abs(dx) <= abs(dy) { isBack = true } else { isSide = true }; case DirSouth: if dy > 0 && abs(dx) <= abs(dy) { isFront = true } else if dy < 0 && abs(dx) <= abs(dy) { isBack = true } else { isSide = true }; case DirWest:  if dx < 0 && abs(dy) <= abs(dx) { isFront = true } else if dx > 0 && abs(dy) <= abs(dx) { isBack = true } else { isSide = true }; case DirEast:  if dx > 0 && abs(dy) <= abs(dx) { isFront = true } else if dx < 0 && abs(dy) <= abs(dx) { isBack = true } else { isSide = true } }; effectiveDist := dist; if isFront { effectiveDist -= 2 } else if isSide { effectiveDist -= 1 } else if isBack { effectiveDist += 1 }; if effectiveDist <= 2 { return 0 }; if effectiveDist <= 3 { return 1 }; if effectiveDist <= 4 { return 2 }; return 3 }
func (e *Enemy) CheckPlayerVisibility(px, py int) int { dist := abs(e.TargetX-px) + abs(e.TargetY-py); dx, dy := px-e.TargetX, py-e.TargetY; isFront, isBack, isSide := false, false, false; switch e.Facing { case DirNorth: if dy < 0 && abs(dx) <= abs(dy) { isFront = true } else if dy > 0 { isBack = true } else { isSide = true }; case DirSouth: if dy > 0 && abs(dx) <= abs(dy) { isFront = true } else if dy < 0 { isBack = true } else { isSide = true }; case DirWest:  if dx < 0 && abs(dy) <= abs(dx) { isFront = true } else if dx > 0 { isBack = true } else { isSide = true }; case DirEast:  if dx > 0 && abs(dy) <= abs(dx) { isFront = true } else if dx < 0 { isBack = true } else { isSide = true } }; ed := dist; if isFront { ed -= 2 } else if isSide { ed -= 1 } else if isBack { ed += 1 }; if ed <= 2 { return 0 }; if ed <= 3 { return 1 }; return 3 }

func (g *Game) InitDungeon() {
	w, h := 40, 40
	d := GenerateDungeon(w, h, g.Rng)
	leader := &Character{Name: "Denim", Stats: Status{AGI: 14}, BaseWT: 290, LoadWeight: 12.0, Facing: DirSouth}
	px, py := 0, 0
	for x := 0; x < w; x++ { for y := 0; y < h; y++ { if d.Tiles[x][y].Type == 1 { px, py = x, y; goto Found } } }
Found:
	leader.TargetX, leader.TargetY = px, py; leader.TargetZ = d.Tiles[px][py].Height
	leader.CurrentX, leader.CurrentY = float64(px), float64(py); leader.CurrentZ = float64(leader.TargetZ)
	g.Dungeon = d; g.Party = &Party{Leader: leader, Members: []*Character{leader}}
	g.Camera = &Camera{ZoomIndex: 3}; g.Log = []string{"Quest Started."}; g.ArrowTimer = 0.3; g.CenterCamera()
}

func GenerateDungeon(w, h int, rng *rand.Rand) *Dungeon {
	d := &Dungeon{Width: w, Height: h, Tiles: make([][]Tile, w)}
	for x := 0; x < w; x++ { d.Tiles[x] = make([]Tile, h); for y := 0; y < h; y++ { d.Tiles[x][y] = Tile{Type: 0, Height: 0, Explored: false} } }
	var rooms []struct{x,y,w,h,z int}
	count := 12 + rng.Intn(6)
	for i:=0; i<count; i++ {
		rw, rh := 4+rng.Intn(4), 4+rng.Intn(4); rx, ry := 1+rng.Intn(w-rw-2), 1+rng.Intn(h-rh-2); rz := rng.Intn(4)
		if len(rooms) > 0 { rz = rooms[len(rooms)-1].z + rng.Intn(3) - 1; if rz < 0 {rz=0}; if rz>5{rz=5} }
		for x:=rx; x<rx+rw; x++ { for y:=ry; y<ry+rh; y++ { d.Tiles[x][y] = Tile{Type: 1, Height: rz} } }
		if len(rooms) > 0 {
			prev := rooms[len(rooms)-1]; cx1, cy1 := prev.x+prev.w/2, prev.y+prev.h/2; cx2, cy2 := rx+rw/2, ry+rh/2; pathZ := prev.z; if rz < pathZ { pathZ = rz }
			sx, ex := cx1, cx2; if sx > ex { sx, ex = ex, sx }; for x:=sx; x<=ex; x++ { d.Tiles[x][cy1] = Tile{Type: 1, Height: pathZ} }
			sy, ey := cy1, cy2; if sy > ey { sy, ey = ey, sy }; for y:=sy; y<=ey; y++ { d.Tiles[cx2][y] = Tile{Type: 1, Height: pathZ} }
		}
		rooms = append(rooms, struct{x,y,w,h,z int}{rx,ry,rw,rh,rz})
	}
	enemyCount := 10 + rng.Intn(5)
	for i := 0; i < enemyCount; i++ {
		for attempt := 0; attempt < 100; attempt++ {
			ex := rng.Intn(w); ey := rng.Intn(h)
			if d.Tiles[ex][ey].Type == 1 {
				ez := d.Tiles[ex][ey].Height; etype, spd := 1, 1; if rng.Float64() < 0.3 { etype = 2; spd = 2 }
				d.Enemies = append(d.Enemies, &Enemy{ID: i, TargetX: ex, TargetY: ey, TargetZ: ez, CurrentX: float64(ex), CurrentY: float64(ey), CurrentZ: float64(ez), Type: etype, Speed: spd, Active: true, Facing: DirWest})
				break
			}
		}
	}
	return d
}

func (g *Game) refreshArrow() { if g.ArrowTimer >= 8.0 { g.ArrowTimer = 0.0 } else { g.ArrowTimer = 0.3 } }

func (g *Game) UpdateDungeon() error {
	g.ArrowTimer += 1.0 / 60.0
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) { g.State = StateMenu; return nil }
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) { g.DebugMode = !g.DebugMode }
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) { g.refreshArrow() }
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			g.Camera.ZoomIndex--; if g.Camera.ZoomIndex < 0 { g.Camera.ZoomIndex = len(ZoomLevels)-1 }
		} else {
			g.Camera.ZoomIndex++; if g.Camera.ZoomIndex >= len(ZoomLevels) { g.Camera.ZoomIndex = 0 }
		}
		g.CenterCamera()
	}
	cameraChanged := false
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) { g.Camera.TargetAngle -= math.Pi / 2; cameraChanged = true }
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) { g.Camera.TargetAngle += math.Pi / 2; cameraChanged = true }
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		cx, cy := ebiten.CursorPosition()
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) { g.IsDragging = true; g.MouseStartX, g.MouseStartY = cx, cy }
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) { g.IsDragging = false }
		if g.IsDragging {
			dx := float64(cx - g.MouseStartX); dy := float64(cy - g.MouseStartY)
			if math.Abs(dx) > 0 || math.Abs(dy) > 0 { cameraChanged = true }
			g.Camera.X += dx; g.Camera.Y += dy; g.MouseStartX, g.MouseStartY = cx, cy
		}
	} else { g.IsDragging = false }
	if cameraChanged { g.refreshArrow() }
	diff := g.Camera.TargetAngle - g.Camera.Angle
	if math.Abs(diff) > 0.01 { g.Camera.Angle += diff * 0.1; g.CenterCamera() } else { g.Camera.Angle = g.Camera.TargetAngle }

	animate := func(curr *float64, target int) bool {
		d := float64(target) - *curr
		if math.Abs(d) < 0.05 { *curr = float64(target); return false }
		*curr += d * MoveSpeed; return true
	}
	leader := g.Party.Leader
	imX := animate(&leader.CurrentX, leader.TargetX); imY := animate(&leader.CurrentY, leader.TargetY); imZ := animate(&leader.CurrentZ, leader.TargetZ)
	leader.IsMoving = imX || imY || imZ
	anyEnemyMoving := false
	for _, e := range g.Dungeon.Enemies {
		if !e.Active { continue }
		emX := animate(&e.CurrentX, e.TargetX); emY := animate(&e.CurrentY, e.TargetY); emZ := animate(&e.CurrentZ, e.TargetZ)
		e.IsMoving = emX || emY || emZ; if e.IsMoving { anyEnemyMoving = true }
	}
	if leader.IsMoving || anyEnemyMoving || g.Party.InCombat {
		if leader.IsMoving && !g.IsDragging { g.CenterCamera() }
		if g.Party.InCombat && inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.Party.InCombat = false; g.Log = append(g.Log, "Combat Reset.")
			for _, e := range g.Dungeon.Enemies { if abs(leader.TargetX-e.TargetX)+abs(leader.TargetY-e.TargetY) <= 1 { e.Active = false } }
		}
		return nil
	}
	inputDx, inputDy := 0, 0; pressed := false
	if !ebiten.IsKeyPressed(ebiten.KeyControl) {
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) { inputDy = -1; pressed = true }
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) { inputDy = 1; pressed = true }
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) { inputDx = -1; pressed = true }
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) { inputDx = 1; pressed = true }
	}
	if pressed {
		g.refreshArrow()
		snapAngle := math.Round(g.Camera.TargetAngle / (math.Pi/2)) * (math.Pi/2)
		cos := math.Cos(-snapAngle); sin := math.Sin(-snapAngle)
		fdx := float64(inputDx)*cos - float64(inputDy)*sin; fdy := float64(inputDx)*sin + float64(inputDy)*cos
		dx := int(math.Round(fdx)); dy := int(math.Round(fdy)); newFacing := leader.Facing
		if dx == 0 && dy == -1 { newFacing = DirNorth }
		if dx == 0 && dy == 1  { newFacing = DirSouth }
		if dx == -1 && dy == 0 { newFacing = DirWest }
		if dx == 1 && dy == 0  { newFacing = DirEast }
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if leader.Facing != newFacing { leader.Facing = newFacing; wt := int(math.Ceil(float64(g.Party.ExplorationWT()) * 0.1 * leader.WeightPenalty())); if wt < 1 { wt = 1 }; g.Party.TotalTurns += wt }
			return nil
		}
		nx, ny := leader.TargetX+dx, leader.TargetY+dy
		if nx >= 0 && nx < g.Dungeon.Width && ny >= 0 && ny < g.Dungeon.Height {
			tile := g.Dungeon.Tiles[nx][ny]; curTile := g.Dungeon.Tiles[leader.TargetX][leader.TargetY]
			if tile.Type == 1 && abs(tile.Height-curTile.Height) <= 1 {
				leader.TargetX = nx; leader.TargetY = ny; leader.TargetZ = tile.Height; leader.Facing = newFacing; leader.IsMoving = true
				baseWT := g.Party.ExplorationWT(); penalty := leader.WeightPenalty()
				cost := int(math.Ceil(float64(baseWT) * penalty)); g.Party.TotalTurns += cost
				for _, e := range g.Dungeon.Enemies { if e.Active && abs(leader.TargetX-e.TargetX)+abs(leader.TargetY-e.TargetY) <= 1 { g.StartCombat(e); return nil } }
				g.ProcessEnemyTurn()
			}
		}
	}
	return nil
}