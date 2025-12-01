// filename: main.go
package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// 初期設定ファイル名
const SettingsFilename = "settings.txt"

// NewGame は app_init.go で定義されたヘルパー関数
func NewGame() *Game {
	g := initializeNewGame()
	
	// 設定ファイルを読み込み、初期値を上書き
	if settings, err := LoadSettings(SettingsFilename); err == nil {
		g.ApplySettings(settings)
	} else if !os.IsNotExist(err) {
		log.Printf("Error loading settings: %v", err)
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