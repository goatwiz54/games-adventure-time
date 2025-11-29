// filename: world2.go
package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

func GenerateWorld2(rng *rand.Rand, minPercent, maxPercent, width, height int, game *Game) *WorldMap2 {
	
	// バリデーション
	if width < 20 || height < 20 {
		if game != nil {
			game.WarningMsg = "Error: Size must be at least 20x20!"
			game.WarningTimer = 3.0
		}
		return nil
	}
	if width > 500 || height > 500 {
		if game != nil {
			game.WarningMsg = "Error: Size too large! (Max 500)"
			game.WarningTimer = 3.0
		}
		return nil
	}

	w2 := &WorldMap2{
		Width:  width,
		Height: height,
		Tiles:  make([][]World2Tile, width),
		OffsetX: float64(width * World2TileSize / 2),
		OffsetY: float64(height * World2TileSize / 2),
	}
	
	// 初期ズーム計算
	mapPixelW := float64(width * World2TileSize)
	mapPixelH := float64(height * World2TileSize)
	scaleW := float64(ScreenWidth) / mapPixelW
	scaleH := float64(ScreenHeight) / mapPixelH
	if scaleW < scaleH { w2.Zoom = scaleW } else { w2.Zoom = scaleH }
	w2.Zoom *= 0.9 

	// 1. 初期化
	totalTiles := width * height
	fixedCount := 0
	
	for x := 0; x < width; x++ {
		w2.Tiles[x] = make([]World2Tile, height)
		for y := 0; y < height; y++ {
			if x < 3 || x >= width-3 || y < 3 || y >= height-3 {
				w2.Tiles[x][y] = World2Tile{Type: W2TileFixedOcean}
				fixedCount++
			} else {
				w2.Tiles[x][y] = World2Tile{Type: W2TileVariableOcean}
			}
		}
	}

	// 補助関数
	placeSoil := func(x, y int) bool {
		if x >= 0 && x < width && y >= 0 && y < height {
			if w2.Tiles[x][y].Type == W2TileVariableOcean {
				w2.Tiles[x][y].Type = W2TileSoil
				return true
			}
		}
		return false
	}

	// 2. 目標土数
	if minPercent > maxPercent { minPercent, maxPercent = maxPercent, minPercent }
	if minPercent < 0 { minPercent = 0 }
	if maxPercent > 100 { maxPercent = 100 }
	targetPercent := minPercent
	if maxPercent > minPercent { targetPercent = minPercent + rng.Intn(maxPercent - minPercent + 1) }
	if game != nil { game.LastTargetSoil = targetPercent }

	targetSoilCount := int(float64(totalTiles) * float64(targetPercent) / 100.0)
	currentSoilCount := 0
	
	// 3. 大陸生成
	cx, cy := width/2, height/2
	walkers := 10
	type Walker struct { x, y int }
	activeWalkers := make([]Walker, walkers)
	for i := 0; i < walkers; i++ {
		activeWalkers[i] = Walker{x: cx + rng.Intn(10) - 5, y: cy + rng.Intn(10) - 5}
	}

	safetyCounter := 0
	for currentSoilCount < targetSoilCount && safetyCounter < 2000000 {
		safetyCounter++
		for i := 0; i < walkers; i++ {
			if placeSoil(activeWalkers[i].x, activeWalkers[i].y) { currentSoilCount++ }
			dir := rng.Intn(4)
			switch dir { case 0: activeWalkers[i].y--; case 1: activeWalkers[i].x++; case 2: activeWalkers[i].y++; case 3: activeWalkers[i].x-- }
			wx, wy := activeWalkers[i].x, activeWalkers[i].y
			if wx < 3 || wx >= width-3 || wy < 3 || wy >= height-3 {
				activeWalkers[i].x = cx + rng.Intn(20) - 10
				activeWalkers[i].y = cy + rng.Intn(20) - 10
			}
		}
	}

	// 4. 島生成 (リストに保持)
	r := rng.Intn(100)
	islandCount := 0
	if r < 50 { islandCount = 2 } else if r < 70 { islandCount = 3 + rng.Intn(2) } else if r < 90 { islandCount = 1 }

	type Point struct { x, y int }
	islandCenters := []Point{}

	for i := 0; i < islandCount; i++ {
		var ix, iy int; found := false
		for attempt := 0; attempt < 100; attempt++ {
			ix = rng.Intn(width); iy = rng.Intn(height)
			if w2.Tiles[ix][iy].Type == W2TileVariableOcean {
				dist := math.Sqrt(math.Pow(float64(ix-cx), 2) + math.Pow(float64(iy-cy), 2))
				if dist > float64(width)*0.35 { found = true; break }
			}
		}
		if found {
			islandSize := 20 + rng.Intn(40)
			islandCenters = append(islandCenters, Point{ix, iy})
			// 島生成
			for j := 0; j < islandSize; j++ {
				if placeSoil(ix, iy) { currentSoilCount++ }
				dir := rng.Intn(4); switch dir { case 0: iy--; case 1: ix++; case 2: iy++; case 3: ix-- }
			}
		}
	}

	// 5. 経由島 (Transit Island) の生成 - ステップストーン方式
	
	// 距離計算ヘルパー
	calcDist := func(x1, y1, x2, y2 int) float64 {
		return math.Sqrt(math.Pow(float64(x1-x2), 2) + math.Pow(float64(y1-y2), 2))
	}

	// 最も近い「土(大陸/既存島)」を探す関数
	findNearestSoil := func(tx, ty int) (int, int) {
		minDist := 999999.0
		nx, ny := tx, ty
		// 全走査は重いので、ランダムサンプリング + 中心方向探索で簡易化
		// ここでは簡易的に「マップ上のすべての土」と比較する（サイズが小さいので許容）
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				// Transitも土台として使える
				if w2.Tiles[x][y].Type == W2TileSoil || w2.Tiles[x][y].Type == W2TileTransit {
					// 自分自身（島の構成タイル）は除外したいが、
					// 簡易的に「距離が近すぎる(島の一部)」は除外
					d := calcDist(tx, ty, x, y)
					if d < 8.0 { continue } // 島半径より大きい値
					
					if d < minDist {
						minDist = d
						nx, ny = x, y
					}
				}
			}
		}
		return nx, ny
	}

	// 各島について接続チェック
	for _, center := range islandCenters {
		sx, sy := findNearestSoil(center.x, center.y)
		
		// 土が見つからなければスキップ（通常ありえない）
		if sx == center.x && sy == center.y { continue }

		// 現在地(Start) -> 目的地(Goal)
		currX, currY := float64(sx), float64(sy)
		destX, destY := float64(center.x), float64(center.y)

		// 初期距離チェック
		totalDist := calcDist(int(currX), int(currY), int(destX), int(destY))
		
		// 10マス以上離れていたら架橋プロセス開始
		if totalDist >= 10 {
			// ループ回数制限
			for safety := 0; safety < 50; safety++ {
				// ベクトル計算
				vecX := destX - currX
				vecY := destY - currY
				vecLen := math.Sqrt(vecX*vecX + vecY*vecY)
				if vecLen == 0 { break }

				// 3-5マス進む
				step := 3.0 + float64(rng.Intn(3)) // 3, 4, 5
				
				// 次の座標
				nextX := currX + (vecX/vecLen)*step
				nextY := currY + (vecY/vecLen)*step
				
				// 座標整数化
				ix, iy := int(nextX), int(nextY)

				// 経由島配置 (最大2x2)
				// 既に土や経由島がある場所は上書きしない（海のみ）
				// ただしVariableOceanのみ
				transitCreated := false
				for dy := 0; dy < 2; dy++ {
					for dx := 0; dx < 2; dx++ {
						tx, ty := ix+dx, iy+dy
						if tx >= 3 && tx < width-3 && ty >= 3 && ty < height-3 {
							if w2.Tiles[tx][ty].Type == W2TileVariableOcean {
								w2.Tiles[tx][ty].Type = W2TileTransit
								transitCreated = true
							}
						}
					}
				}

				// 現在地更新
				currX, currY = nextX, nextY

				// 残り距離チェック
				remainDist := calcDist(int(currX), int(currY), int(destX), int(destY))
				
				// 7マス未満になれば終了
				if remainDist < 7 {
					break
				}
				
				// 何も作れなかった場合（大陸の上を通ったなど）、位置だけ進めて次へ
			}
		}
	}

	// 6. 湖の判定 (Flood Fill)
	reached := make([][]bool, width)
	for x := range reached { reached[x] = make([]bool, height) }
	queue := []struct{x,y int}{}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if w2.Tiles[x][y].Type == W2TileFixedOcean {
				reached[x][y] = true
				queue = append(queue, struct{x,y int}{x, y})
			}
		}
	}

	dx := []int{0, 1, 0, -1}
	dy := []int{-1, 0, 1, 0}
	for len(queue) > 0 {
		p := queue[0]; queue = queue[1:]
		for i := 0; i < 4; i++ {
			nx, ny := p.x+dx[i], p.y+dy[i]
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				// 経由島も「陸」扱いなので、水流を止める
				isLand := (w2.Tiles[nx][ny].Type == W2TileSoil || w2.Tiles[nx][ny].Type == W2TileTransit)
				if !reached[nx][ny] && !isLand {
					reached[nx][ny] = true
					queue = append(queue, struct{x,y int}{nx, ny})
				}
			}
		}
	}

	// 7. 最終集計
	lakeCount := 0
	variableOceanCount := 0 
	transitCount := 0
	// currentSoilCountは再計算
	currentSoilCount = 0

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			t := w2.Tiles[x][y].Type
			if t == W2TileVariableOcean {
				if !reached[x][y] {
					w2.Tiles[x][y].IsLake = true
					lakeCount++
				} else {
					variableOceanCount++
				}
			} else if t == W2TileSoil {
				currentSoilCount++
			} else if t == W2TileTransit {
				transitCount++
			}
		}
	}

	w2.StatsInfo = []string{
		fmt.Sprintf("Size:   %d x %d", width, height),
		fmt.Sprintf("Range:  %d%% - %d%%", minPercent, maxPercent),
		fmt.Sprintf("Target: %d%%", targetPercent),
		"-----------------",
		fmt.Sprintf("Fix Ocean: %4d (%4.1f%%)", fixedCount, float64(fixedCount)/float64(totalTiles)*100),
		fmt.Sprintf("Var Ocean: %4d (%4.1f%%)", variableOceanCount, float64(variableOceanCount)/float64(totalTiles)*100),
		fmt.Sprintf("Lake:      %4d (%4.1f%%)", lakeCount, float64(lakeCount)/float64(totalTiles)*100),
		fmt.Sprintf("Soil:      %4d (%4.1f%%)", currentSoilCount, float64(currentSoilCount)/float64(totalTiles)*100),
		fmt.Sprintf("Transit:   %4d (%4.1f%%)", transitCount, float64(transitCount)/float64(totalTiles)*100),
	}
	return w2
}

func (g *Game) UpdateWorld2() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) { g.State = StateMenu }
	
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		newWorld := GenerateWorld2(g.Rng, g.SoilMin, g.SoilMax, g.W2Width, g.W2Height, g)
		if newWorld != nil {
			g.World2 = newWorld
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) { g.World2.ShowGrid = !g.World2.ShowGrid }

	// UI Input Logic
	if inpututil.IsMouseButton