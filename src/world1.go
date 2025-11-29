// filename: world1.go
package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

func (g *Game) StartWorldLoading() {
	g.State = StateWorldLoading
	g.Loader = &LoadingState{ StartTime: time.Now(), Logs: []LoadingLog{} }
	g.AddLoadingLog("Starting World Generation...")
}

func (g *Game) AddLoadingLog(msg string) {
	g.Loader.Logs = append(g.Loader.Logs, LoadingLog{Msg: msg, AddedAt: time.Now()})
}

func (g *Game) UpdateLoading() {
	elapsed := time.Since(g.Loader.StartTime).Seconds()
	g.Loader.Progress = elapsed / 3.0
	if g.Loader.Progress > 1.0 { g.Loader.Progress = 1.0 }
	
	if elapsed > 0.5 && g.Loader.Step == 0 {
		g.AddLoadingLog("Generating Map Structure...")
		g.World = &WorldMap{ Tiles: make([][]WorldTile, WorldWidth), CameraX: float64(WorldWidth * WorldTileSize / 2), CameraY: float64(WorldHeight * WorldTileSize / 2), Zoom: 1.0 }
		for x := 0; x < WorldWidth; x++ { g.World.Tiles[x] = make([]WorldTile, WorldHeight) }
		g.Loader.Step++
	}
	if elapsed > 1.0 && g.Loader.Step == 1 {
		g.AddLoadingLog("Generating Terrain Data...")
		for x := 0; x < WorldWidth; x++ {
			for y := 0; y < WorldHeight; y++ {
				nx := float64(x)*0.1; ny := float64(y)*0.1; val := math.Sin(nx)*math.Cos(ny)+math.Sin(nx*2.5)*0.5
				biome := 0; if val>0.6{biome=4} else if val>0.3{biome=2} else if val>0.0{biome=1} else if val>-0.2{biome=3}
				isRoad := false; if (x==WorldWidth/2 || y==WorldHeight/2) && biome!=0 { isRoad=true }
				g.World.Tiles[x][y] = WorldTile{Biome: biome, Height: val, IsRoad: isRoad}
			}
		}
		g.Loader.Step++
	}
	if elapsed > 3.0 { g.State = StateWorld }
}

func (g *Game) UpdateWorld() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) { g.State = StateMenu }
	moveSpd := 10.0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) { g.World.CameraX -= moveSpd }
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) { g.World.CameraX += moveSpd }
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) { g.World.CameraY -= moveSpd }
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) { g.World.CameraY += moveSpd }
	worldPixelW := float64(WorldWidth * WorldTileSize); if g.World.CameraX < 0 { g.World.CameraX += worldPixelW }; if g.World.CameraX >= worldPixelW { g.World.CameraX -= worldPixelW }
	worldPixelH := float64(WorldHeight * WorldTileSize); if g.World.CameraY < 0 { g.World.CameraY = 0 }; if g.World.CameraY > worldPixelH { g.World.CameraY = worldPixelH }
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) { g.DebugMode = !g.DebugMode }
	return nil
}

func (g *Game) DrawLoading(screen *ebiten.Image) {
	screen.Fill(color.Black)
	barW, barH := 400.0, 10.0; x, y := float64(ScreenWidth)/2 - barW/2, float64(ScreenHeight) - 100.0
	ebitenutil.DrawRect(screen, x, y, barW, barH, color.RGBA{50, 50, 50, 255})
	ebitenutil.DrawRect(screen, x, y, barW * g.Loader.Progress, barH, color.RGBA{0, 200, 0, 255})
	showCount := 0
	for i := len(g.Loader.Logs) - 1; i >= 0; i-- {
		l := g.Loader.Logs[i]; elapsed := time.Since(l.AddedAt).Seconds(); alpha := 1.0; if elapsed > 1.5 { alpha = 1.0 - (elapsed-1.5)*2 }
		if alpha < 0 { continue }; if showCount >= 3 { break }
		c := color.RGBA{200, 200, 200, uint8(255 * alpha)}
		text.Draw(screen, l.Msg, basicfont.Face7x13, int(x), int(y-30.0-float64(showCount)*20.0), c); showCount++
	}
}

func (g *Game) DrawWorld(screen *ebiten.Image) {
	screen.Fill(color.RGBA{210, 190, 160, 255})
	curveFactor := 0.00005 
	rangeX := ScreenWidth/WorldTileSize/2 + 2; rangeY := ScreenHeight/WorldTileSize/2 + 4
	camX, camY := g.World.CameraX, g.World.CameraY
	for y := -rangeY; y <= rangeY; y++ {
		for x := -rangeX; x <= rangeX; x++ {
			wx := (int(camX)/WorldTileSize + x); wy := (int(camY)/WorldTileSize + y)
			if wy < 0 || wy >= WorldHeight { continue }
			normWX := (wx % WorldWidth + WorldWidth) % WorldWidth
			tile := g.World.Tiles[normWX][wy]
			sx := float64(x * WorldTileSize) + (float64(int(camX)%WorldTileSize) * -1) + ScreenWidth/2
			sy := float64(y * WorldTileSize) + (float64(int(camY)%WorldTileSize) * -1) + ScreenHeight/2
			dx := sx - ScreenWidth/2; dy := sy - ScreenHeight/2; distSq := dx*dx + dy*dy; sy += distSq * curveFactor
			var img *ebiten.Image
			switch tile.Biome {
			case 0: img = TexW_Ocean 
			case 1: img = TexW_Plains
			case 2: img = TexW_Forest
			case 3: img = TexW_Desert
			case 4: img = TexW_Mountain
			}
			op := &ebiten.DrawImageOptions{}
			scale := float64(WorldTileSize) / 32.0; op.GeoM.Scale(scale, scale); op.GeoM.Translate(sx, sy)
			if sx < -WorldTileSize || sx > ScreenWidth || sy < -WorldTileSize || sy > ScreenHeight { continue }
			if img != nil { screen.DrawImage(img, op) } else { ebitenutil.DrawRect(screen, sx, sy, float64(WorldTileSize), float64(WorldTileSize), color.Gray{128}) }
			if tile.Biome == 2 { screen.DrawImage(TexW_TreeIcon, op) }
			if tile.Biome == 4 { screen.DrawImage(TexW_MountainIcon, op) }
			if tile.IsRoad { ebitenutil.DrawRect(screen, sx+float64(WorldTileSize)*0.4, sy+float64(WorldTileSize)*0.4, 4, 4, color.RGBA{100,50,0,200}) }
		}
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("World Map 1\nPos: %.0f, %.0f", camX, camY))
}