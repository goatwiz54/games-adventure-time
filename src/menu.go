// filename: menu.go
package main

import (
	"image/color"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

func (g *Game) UpdateMenu() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) { g.MenuIndex--; if g.MenuIndex < 0 { g.MenuIndex = 2 } }
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) { g.MenuIndex++; if g.MenuIndex > 2 { g.MenuIndex = 0 } }
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if g.MenuIndex == 0 {
			g.StartWorldLoading()
		} else if g.MenuIndex == 1 {
			g.State = StateDungeon
		} else if g.MenuIndex == 2 {
			// サイズ指定付きで生成
			g.World2 = GenerateWorld2(g.Rng, g.SoilMin, g.SoilMax, g.W2Width, g.W2Height, g)
			g.State = StateWorld2
		}
	}
	return nil
}

func (g *Game) DrawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 20, 255})
	text.Draw(screen, "TACTICS DUNGEON RPG", basicfont.Face7x13, ScreenWidth/2-80, ScreenHeight/3, color.White)
	var c1, c2, c3 color.Color = color.White, color.White, color.White
	if g.MenuIndex == 0 { c1 = color.RGBA{255, 200, 0, 255} }
	if g.MenuIndex == 1 { c2 = color.RGBA{255, 200, 0, 255} }
	if g.MenuIndex == 2 { c3 = color.RGBA{255, 200, 0, 255} }
	text.Draw(screen, "> 1. World Map 1 (Globe)", basicfont.Face7x13, ScreenWidth/2-100, ScreenHeight/2, c1)
	text.Draw(screen, "> 2. Dungeon", basicfont.Face7x13, ScreenWidth/2-100, ScreenHeight/2+30, c2)
	text.Draw(screen, "> 3. World Map 2 (Flat/Zoom)", basicfont.Face7x13, ScreenWidth/2-100, ScreenHeight/2+60, c3)
}