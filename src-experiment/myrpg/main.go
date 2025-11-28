package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// 設定定数
const (
	ScreenWidth  = 800
	ScreenHeight = 600
	TileWidth    = 64  // 描画上のタイルの幅
	TileHeight   = 32  // 描画上のタイルの高さ（クォータービューのひし形）
)

type Game struct {
	dungeon *DungeonMap
	player  *Character
	cameraX float64
	cameraY float64
}

func (g *Game) Update() error {
	// カメラ移動（矢印キー）
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft)  { g.cameraX += 4 }
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) { g.cameraX -= 4 }
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp)    { g.cameraY += 4 }
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown)  { g.cameraY -= 4 }

	// デバッグ：スペースキーで重量変更テスト
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.player.Inventory.TotalWeight += 0.5
		if g.player.Inventory.TotalWeight > g.player.MaxLoadWeight() * 1.2 {
			g.player.Inventory.TotalWeight = 0
		}
	}
	
	// デバッグ：Rキーでダンジョン再生成
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.dungeon = GenerateDungeon(20, 20)
	}

	return nil
}

// IsoToScreen はマップ座標(x,y,z)をスクリーン座標(sx,sy)に変換します
func IsoToScreen(x, y, z int) (float64, float64) {
	// クォータービュー変換式
	sx := float64(x-y) * (TileWidth / 2.0)
	sy := float64(x+y) * (TileHeight / 2.0)
	
	// 高さ(z)による補正（上にずらす）
	sy -= float64(z) * 20.0 
	
	return sx, sy
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 背景塗りつぶし
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// マップ描画オフセット（画面中央に表示するため）
	offsetX := float64(ScreenWidth/2) + g.cameraX
	offsetY := float64(100) + g.cameraY

	// マップの描画（奥から手前へ描画しないと重なりがおかしくなる）
	for x := 0; x < g.dungeon.Width; x++ {
		for y := 0; y < g.dungeon.Height; y++ {
			tile := g.dungeon.Tiles[x][y]
			if tile.Type == 0 {
				continue // 壁（今回は床だけ描画）
			}

			sx, sy := IsoToScreen(x, y, tile.Height)
			
			// 簡易的なタイルの描画（ひし形の代わりに矩形を変形させて表現もできるが、今回はラインで描画）
			// 実際には画像リソースを用意して DrawImage するのが一般的
			
			// 床の上面
			pathColor := color.RGBA{100 + uint8(tile.Height*40), 100 + uint8(tile.Height*40), 100, 255}
			
			// 擬似的なひし形描画
			pX, pY := sx + offsetX, sy + offsetY
			
			// Ebitenで図形を描くための頂点定義などは複雑になるため、
			// デバッグ文字として床を描画してみる
			ebitenutil.DebugPrintAt(screen, "[]", int(pX), int(pY))
			
			// あるいは小さな矩形で代用
			// vector.DrawFilledRect(screen, float32(pX), float32(pY), 30, 15, pathColor, false) 
			// ※vectorパッケージが必要になるため、今回は簡易テキストと高さ情報のみ表示
			
			// 高さを視覚化するための棒
			ebitenutil.DrawLine(screen, pX, pY, pX, pY+float64(tile.Height)*10, pathColor)
		}
	}

	// UI情報の表示
	msg := fmt.Sprintf(
		"Player: %s (%s %s)\n"+
		"HP:%d SP:%d\n"+
		"STR:%d INT:%d AGI:%d\n"+
		"Weight: %.1f / %.1f kg (Penalty: x%.1f)\n"+
		"Base WT: %d\n"+
		"Game WT: %d\n\n"+
		"[Arrow]: Move Camera, [Space]: Add Weight, [R]: Re-Gen Map",
		g.player.Name, g.player.Race, g.player.Job,
		g.player.Stats.HP, g.player.Stats.SP,
		g.player.Stats.STR, g.player.Stats.INT, g.player.Stats.AGI,
		g.player.Inventory.TotalWeight, g.player.MaxLoadWeight(), g.player.TimePenaltyRate(),
		g.player.BaseWT,
		g.player.CalculateGameWT(),
	)
	text.Draw(screen, msg, basicfont.Face7x13, 10, 20, color.White)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	// ゲーム初期化
	g := &Game{
		dungeon: GenerateDungeon(20, 20),
		player:  NewCharacter(1, "Hero", "Human"),
	}

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("My Dungeon RPG Prototype")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}