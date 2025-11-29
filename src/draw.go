// filename: draw.go
package main

import (
	"fmt" // Added
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

func (g *Game) GetScale() float64 {
	return ZoomLevels[g.Camera.ZoomIndex] / 0.8
}

func (g *Game) CenterCamera() {
	if g.State == StateDungeon {
		leader := g.Party.Leader
		s := g.GetScale()
		isoX, isoY := IsoToScreen(leader.CurrentX, leader.CurrentY, leader.CurrentZ, s, g.Camera.Angle)
		g.Camera.X = -(isoX - ScreenWidth/2)
		g.Camera.Y = -(isoY - ScreenHeight/2)
	}
}

func IsoToScreen(x, y, z float64, scale float64, angle float64) (float64, float64) {
	cos := math.Cos(angle); sin := math.Sin(angle)
	rx := x*cos - y*sin; ry := x*sin + y*cos
	tw := float64(TileWidth) * scale; th := float64(TileHeight) * scale
	sx := (rx - ry) * (tw / 2.0); sy := (rx + ry) * (th / 2.0); sy -= z * 16.0 * scale
	return sx, sy
}

func (g *Game) DrawDungeon(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 25, 30, 255})
	s := g.GetScale()
	ang := g.Camera.Angle

	type RenderItem struct { Type int; Depth float64; Obj interface{}; X, Y int }
	var items []RenderItem
	getSortDepth := func(x, y float64) float64 { _, sy := IsoToScreen(x, y, 0, s, ang); return sy }

	for x := 0; x < g.Dungeon.Width; x++ {
		for y := 0; y < g.Dungeon.Height; y++ {
			t := &g.Dungeon.Tiles[x][y]
			if t.Type == 0 { continue }
			vis := g.GetVisibility(x, y)
			if vis < 3 && !g.DebugMode { t.Explored = true }
			if !t.Explored && !g.DebugMode { continue }
			items = append(items, RenderItem{Type: 0, Depth: getSortDepth(float64(x), float64(y)), Obj: t, X: x, Y: y})
		}
	}
	leader := g.Party.Leader
	charDepth := getSortDepth(leader.CurrentX, leader.CurrentY) + float64(TileHeight/2)*s
	items = append(items, RenderItem{Type: 1, Depth: charDepth, Obj: leader})
	for _, e := range g.Dungeon.Enemies {
		if !e.Active { continue }
		vis := g.GetVisibility(int(math.Round(e.CurrentX)), int(math.Round(e.CurrentY)))
		if vis < 3 || g.DebugMode {
			eDepth := getSortDepth(e.CurrentX, e.CurrentY) + float64(TileHeight/2)*s
			items = append(items, RenderItem{Type: 2, Depth: eDepth, Obj: e})
		}
	}
	sort.Slice(items, func(i, j int) bool { if math.Abs(items[i].Depth - items[j].Depth) > 0.1 { return items[i].Depth < items[j].Depth }; return items[i].Type < items[j].Type })

	arrowAlpha := 0.0
	if g.ArrowTimer < 0.3 { arrowAlpha = g.ArrowTimer / 0.3 } else if g.ArrowTimer < 5.0 { arrowAlpha = 1.0 } else if g.ArrowTimer < 8.0 { arrowAlpha = 1.0 - (g.ArrowTimer - 5.0) / 3.0 }

	for _, item := range items {
		switch item.Type {
		case 0: 
			t := item.Obj.(*Tile); sx, sy := IsoToScreen(float64(item.X), float64(item.Y), float64(t.Height), s, ang); sx += g.Camera.X; sy += g.Camera.Y; vis := g.GetVisibility(item.X, item.Y)
			drawTexturedBlock(screen, sx, sy, t.Height, s)
			if vis == 3 && t.Explored { drawOverlay(screen, sx, sy, t.Height, s, color.RGBA{0, 0, 0, 180}) } else if vis > 0 { alpha := uint8(0); if vis==1{alpha=80}; if vis==2{alpha=160}; drawOverlay(screen, sx, sy, t.Height, s, color.RGBA{0, 0, 0, alpha}) }
		case 1: 
			p := item.Obj.(*Character); sx, sy := IsoToScreen(p.CurrentX, p.CurrentY, p.CurrentZ, s, ang); drawUnit(screen, sx+g.Camera.X, sy+g.Camera.Y, s, color.RGBA{50, 100, 255, 255}, p.Facing, g.Camera.Angle, true, arrowAlpha)
		case 2:
			e := item.Obj.(*Enemy); c := color.RGBA{220, 50, 50, 255}; if e.Type == 2 { c = color.RGBA{220, 100, 100, 255} }
			sx, sy := IsoToScreen(e.CurrentX, e.CurrentY, e.CurrentZ, s, ang); pcVis := g.GetVisibility(int(math.Round(e.CurrentX)), int(math.Round(e.CurrentY))); drawUnit(screen, sx+g.Camera.X, sy+g.Camera.Y, s, c, e.Facing, g.Camera.Angle, pcVis <= 1, arrowAlpha)
		}
	}
	g.DrawDungeonUI(screen)
}

func drawTexturedBlock(screen *ebiten.Image, x, y float64, h int, scale float64) {
	w := float32(float64(TileWidth) * scale); hh := float32(float64(TileHeight) * scale); cx, cy := float32(x), float32(y); depth := float32(float64(h)*16.0*scale + 10.0*scale)
	drawTexturedQuad(screen, TexDirt, cx, cy+hh, cx+w/2, cy+hh/2, cx+w/2, cy+hh/2+depth, cx, cy+hh+depth, color.RGBA{120, 120, 120, 255})
	drawTexturedQuad(screen, TexDirt, cx, cy+hh, cx-w/2, cy+hh/2, cx-w/2, cy+hh/2+depth, cx, cy+hh+depth, color.RGBA{200, 200, 200, 255})
	topTex := TexGrass; if h == 0 { topTex = TexStone }
	drawTexturedQuad(screen, topTex, cx, cy, cx+w/2, cy+hh/2, cx, cy+hh, cx-w/2, cy+hh/2, color.White)
}

func drawTexturedQuad(screen *ebiten.Image, tex *ebiten.Image, x1, y1, x2, y2, x3, y3, x4, y4 float32, clr color.Color) {
	r, g, b, a := clr.RGBA(); cr, cg, cb, ca := float32(r)/65535, float32(g)/65535, float32(b)/65535, float32(a)/65535
	w, h := float32(tex.Bounds().Dx()), float32(tex.Bounds().Dy())
	var vs []ebiten.Vertex
	vs = append(vs, ebiten.Vertex{DstX: x1, DstY: y1, SrcX: 0, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca})
	vs = append(vs, ebiten.Vertex{DstX: x2, DstY: y2, SrcX: w, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca})
	vs = append(vs, ebiten.Vertex{DstX: x3, DstY: y3, SrcX: w, SrcY: h, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca})
	vs = append(vs, ebiten.Vertex{DstX: x4, DstY: y4, SrcX: 0, SrcY: h, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca})
	is := []uint16{0, 1, 2, 2, 3, 0}; screen.DrawTriangles(vs, is, tex, nil)
}

func drawOverlay(screen *ebiten.Image, x, y float64, h int, scale float64, c color.Color) {
	w := float32(float64(TileWidth) * scale); hh := float32(float64(TileHeight) * scale); cx, cy := float32(x), float32(y)
	drawTexturedQuad(screen, TexWhite, cx, cy, cx+w/2, cy+hh/2, cx, cy+hh, cx-w/2, cy+hh/2, c)
}

func drawUnit(screen *ebiten.Image, x, y float64, scale float64, c color.Color, facing int, camAngle float64, showArrow bool, arrowAlpha float64) {
	s := float32(scale)
	ebitenutil.DrawRect(screen, x-6*float64(s), y+14*float64(s), 12*float64(s), 4*float64(s), color.RGBA{0,0,0,100})
	ebitenutil.DrawRect(screen, x-4*float64(s), y-14*float64(s), 8*float64(s), 22*float64(s), c)
	if showArrow && arrowAlpha > 0 {
		op := &ebiten.DrawImageOptions{}; op.GeoM.Translate(-64, -64); baseAng := (315.0 + float64(facing)*90.0) * math.Pi / 180; finalRad := baseAng + camAngle
		op.GeoM.Rotate(finalRad); op.GeoM.Scale(1.0, 0.5); op.GeoM.Scale(float64(s), float64(s)); op.GeoM.Scale(0.35, 0.35)
		op.GeoM.Translate(x, y-28*float64(s)); op.ColorScale.ScaleAlpha(float32(arrowAlpha)); op.Filter = ebiten.FilterLinear
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
	cx, cy := float64(ScreenWidth)-60, float64(ScreenHeight)-60; dirs := []string{"N", "E", "S", "W"}; radius := 40.0
	for i, d := range dirs {
		baseAng := (315.0 + float64(i)*90.0) * math.Pi / 180; drawAng := baseAng + g.Camera.Angle
		dx := math.Cos(drawAng) * radius; dy := math.Sin(drawAng) * radius
		text.Draw(screen, d, basicfont.Face7x13, int(cx+dx)-4, int(cy+dy)+4, color.White)
	}
	for i, l := range g.Log { text.Draw(screen, l, basicfont.Face7x13, 10, ScreenHeight-100+(i*15), color.RGBA{220, 220, 220, 255}) }
}