// filename: main.go
package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	g := &Game{ 
		State: StateMenu, 
		MenuIndex: 0, 
		Rng: rng,
		// World 2 Init
		SoilMin: 20,
		SoilMax: 28,
		W2Width: DefaultWorld2Width,
		W2Height: DefaultWorld2Height,
		TransitDist: 15,
		VastOceanSize: 25,
		IslandBoundSize: 15,
		
		MapTypeMain: 1, 
		MapTypeSub:  1, 
		MapRatio:    10,
		EnableCentering: true,

		// Cliff Params Default
		CliffInitVal:  10.0,
		CliffDecVal:   0.1,
		ShallowDecVal: 0.25,
		CliffPathLen:  5, // New
		ForceSwitch:   5, // New
	}
	g.InitDungeon()
	return g
}

func (g *Game) Update() error {
	if g.WarningTimer > 0 { g.WarningTimer -= 1.0 / 60.0 }
	switch g.State {
	case StateMenu: return g.UpdateMenu()
	case StateWorldLoading: g.UpdateLoading(); return nil
	case StateWorld: return g.UpdateWorld()
	case StateDungeon: return g.UpdateDungeon()
	case StateWorld2: return g.UpdateWorld2()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.State {
	case StateMenu: g.DrawMenu(screen)
	case StateWorldLoading: g.DrawLoading(screen)
	case StateWorld: g.DrawWorld(screen)
	case StateDungeon: g.DrawDungeon(screen)
	case StateWorld2: g.DrawWorld2(screen)
	}
	
	DrawMemoryStats(screen)
}

func (g *Game) Layout(w, h int) (int, int) { return ScreenWidth, ScreenHeight }

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Tactics Dungeon: Cliff Logic Update")
	if err := ebiten.RunGame(NewGame()); err != nil { log.Fatal(err) }
}