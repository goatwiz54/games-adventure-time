// filename: world2/world2.go
package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// InitWorld2Generator は main.go/menu.go から参照されるため、ここに残す
func (g *Game) InitWorld2Generator() {
	// Validation
	if g.W2Width < 40 || g.W2Height < 40 {
		g.WarningMsg = "Size >= 40x40"
		g.WarningTimer = 3.0
		return
	}
	if g.W2Width > 500 || g.W2Height > 500 {
		g.WarningMsg = "Size <= 500x500"
		g.WarningTimer = 3.0
		return
	}

	g.World2 = &WorldMap2{
		Width:     g.W2Width,
		Height:    g.W2Height,
		Tiles:     make([][]World2Tile, g.W2Width),
		OffsetX:   float64(g.W2Width * World2TileSize / 2),
		OffsetY:   float64(g.W2Height * World2TileSize / 2),
		PinkRects: []Rect{},
	}

	mapPixelW := float64(g.W2Width * World2TileSize)
	mapPixelH := float64(g.W2Height * World2TileSize)
	scaleW := float64(ScreenWidth) / mapPixelW
	scaleH := float64(ScreenHeight) / mapPixelH
	if scaleW < scaleH {
		g.World2.Zoom = scaleW
	} else {
		g.World2.Zoom = scaleH
	}
	g.World2.Zoom *= 0.9

	// シード初期化
	startSeed := rand.Int63()

	gen := &World2Generator{
		CurrentStep: 0,
		IsFinished:  false,
		PhaseName:   "0. Init (Fixed Ocean)",
		History:     []GenSnapshot{},
		Rng:         rand.New(rand.NewSource(startSeed)),
		CurrentSeed: startSeed,
		Config: GenConfig{
			MinPct: g.SoilMin, MaxPct: g.SoilMax, W: g.W2Width, H: g.W2Height,
			TransitDist: g.TransitDist,
			VastOcean:   g.VastOceanSize, IslandBound: g.IslandBoundSize,
			// MapTypeMain, MapTypeSub は固定値 1
			MainType: 1, SubType: 1, Ratio: g.MapRatio,
			Centering: g.EnableCentering,
			CliffInit: g.CliffInitVal, CliffDec: g.CliffDecVal, ShallowDec: g.ShallowDecVal,
			CliffPathLen: g.CliffPathLen,
			ForceSwitch:  g.ForceSwitch,
		},
		Multiplier: g.CliffInitVal,
		Excluded:   make(map[int]bool),
		NewSoils:   make(map[int]bool),
		Islands:    []IslandData{}, // 島データを初期化
	}

	for x := 0; x < g.W2Width; x++ {
		g.World2.Tiles[x] = make([]World2Tile, g.W2Height)
		for y := 0; y < g.W2Height; y++ {
			if x < 3 || x >= g.W2Width-3 || y < 3 || y >= g.W2Height-3 {
				g.World2.Tiles[x][y] = World2Tile{Type: W2TileFixedOcean}
			} else {
				g.World2.Tiles[x][y] = World2Tile{Type: W2TileVariableOcean}
			}
		}
	}

	g.Gen2 = gen
	g.TotalRoute1Dist = 0.0 // 初期化
	g.Gen2.UpdateMaskImage(g.W2Width, g.W2Height)
	g.SaveSnapshot()
}

func (g *Game) SaveSnapshot() {
	w, h := g.Gen2.Config.W, g.Gen2.Config.H
	tilesCopy := make([][]World2Tile, w)
	for x := 0; x < w; x++ {
		tilesCopy[x] = make([]World2Tile, h)
		copy(tilesCopy[x], g.World2.Tiles[x])
	}

	newSoilsCopy := make(map[int]bool)
	for k, v := range g.Gen2.NewSoils {
		newSoilsCopy[k] = v
	}

	exCopy := make(map[int]bool)
	for k, v := range g.Gen2.Excluded {
		exCopy[k] = v
	}

	pinkCopy := make([]Rect, len(g.World2.PinkRects))
	copy(pinkCopy, g.World2.PinkRects)

	walkersCopy := make([]struct{ x, y int }, len(g.Gen2.Walkers))
	copy(walkersCopy, g.Gen2.Walkers)

	g.Gen2.History = append(g.Gen2.History, GenSnapshot{
		Tiles:            tilesCopy,
		PhaseName:        g.Gen2.PhaseName,
		StepID:           g.Gen2.CurrentStep,
		NewSoils:         newSoilsCopy,
		Excluded:         exCopy,
		Multiplier:       g.Gen2.Multiplier,
		PinkRects:        pinkCopy,
		Walkers:          walkersCopy,
		CurrentSoilCount: g.Gen2.CurrentSoilCount,
		CurrentSeed:      g.Gen2.CurrentSeed,
		CliffStreak:      g.Gen2.CliffStreak,
		ShallowStreak:    g.Gen2.ShallowStreak,
	})
}

func (g *Game) UndoStep() {
	if len(g.Gen2.History) > 1 {
		g.Gen2.History = g.Gen2.History[:len(g.Gen2.History)-1]
		last := g.Gen2.History[len(g.Gen2.History)-1]

		for x := 0; x < g.Gen2.Config.W; x++ {
			copy(g.World2.Tiles[x], last.Tiles[x])
		}

		g.Gen2.PhaseName = last.PhaseName
		g.Gen2.CurrentStep = last.StepID
		g.Gen2.IsFinished = false

		g.Gen2.NewSoils = make(map[int]bool)
		for k, v := range last.NewSoils {
			g.Gen2.NewSoils[k] = v
		}

		g.Gen2.Excluded = make(map[int]bool)
		for k, v := range last.Excluded {
			g.Gen2.Excluded[k] = v
		}

		g.World2.PinkRects = make([]Rect, len(last.PinkRects))
		copy(g.World2.PinkRects, last.PinkRects)

		g.Gen2.Walkers = make([]struct{ x, y int }, len(last.Walkers))
		copy(g.Gen2.Walkers, last.Walkers)

		g.Gen2.CurrentSoilCount = last.CurrentSoilCount
		g.Gen2.Multiplier = last.Multiplier
		g.Gen2.CurrentSeed = last.CurrentSeed
		g.Gen2.Rng.Seed(g.Gen2.CurrentSeed)
		g.Gen2.CliffStreak = last.CliffStreak
		g.Gen2.ShallowStreak = last.ShallowStreak
	}
}

// NextStep のラッパー
func (g *Game) NextStep() {
	if g.Gen2.IsFinished {
		return
	}

	gen := g.Gen2
	w, h := gen.Config.W, gen.Config.H

	gen.Rng.Seed(gen.CurrentSeed)
	rng := gen.Rng

	gen.NewSoils = make(map[int]bool)
	g.World2.PinkRects = []Rect{}

	switch gen.CurrentStep {
	case Phase_Init:
		g.PhaseInit(w, h, rng, gen)
	case Phase_MaskGen:
		g.PhaseMaskGen(w, h, rng, gen)
	case Phase_SoilStart, 3, 4, 5, 6, 7, 8, 9, 10, Phase_SoilProgressEnd:
		g.PhaseSoilProgress(w, h, rng, gen)
	case Phase_Bridge:
		g.PhaseBridge(w, h, rng, gen)
	case Phase_Centering:
		g.PhaseCentering(w, h, rng, gen)
	case Phase_IslandsQuad:
		g.PhaseIslandsQuad(w, h, rng, gen)
	case Phase_IslandsRand:
		g.PhaseIslandsRand(w, h, rng, gen)
	case Phase_Transit_Start:
		g.PhaseTransitStart(w, h, rng, gen)
	case Phase_IslandShallowAdjust:
		g.PhaseIslandShallowAdjust(w, h, rng, gen)
	case Phase_Transit_Route1:
		g.PhaseTransitRoute1(w, h, rng, gen)
	case Phase_Transit_Route2_Calc:
		g.PhaseTransitRoute2Calc(w, h, rng, gen)
	case Phase_Transit_Route2_Draw:
		g.PhaseTransitRoute2Draw(w, h, rng, gen)
	case Phase_Transit_RouteA:
		g.PhaseTransitRouteA(w, h, rng, gen)
	case Phase_DeepSea:
		g.PhaseDeepSea(w, h, rng, gen)
	case Phase_CliffsShallows:
		g.PhaseCliffsShallows(w, h, rng, gen)
	case Phase_LakesFinal:
		g.PhaseLakesFinal(w, h, rng, gen)
	}

	gen.CurrentStep++                 // 各フェーズの処理メソッド内で次のフェーズに移行するロジックを削除したため、ここでインクリメント
	gen.CurrentSeed = gen.Rng.Int63() // Save next seed
	g.SaveSnapshot()
}

// UpdateWorld2 は main.go から参照されるため、ここに残す
func (g *Game) UpdateWorld2() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.State = StateMenu
	}

	// F1キー: 設定ファイル読み込みとリセット
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		if settings, err := LoadSettings(SettingsFilename); err == nil {
			g.ApplySettings(settings)
			g.InitWorld2Generator() // 設定適用後、生成をリセット
		} else if !os.IsNotExist(err) {
			g.WarningMsg = fmt.Sprintf("Error loading settings: %v", err)
			g.WarningTimer = 3.0
		}
	}

	// ** Rキーは最優先でリセット **
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.InitWorld2Generator() // Reset
		return nil
	}

	// Enter キーの処理 (UI編集モード以外)
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && g.InputMode == EditNone {

		// Phase 10 (IsFinished) の場合はリセットして自動進行開始
		if g.Gen2.IsFinished {
			g.InitWorld2Generator()
			g.AutoProgress = true
			g.SuppressMapDraw = true
		} else {
			// 自動進行モードを開始
			g.AutoProgress = true
			g.SuppressMapDraw = true
		}
	}

	// 自動進行モード: 1フレームごとに1フェーズ進める
	if g.AutoProgress && !g.Gen2.IsFinished {
		g.NextStep()
	}

	// 生成完了時に自動進行を停止し、マップ描画を再開
	if g.Gen2.IsFinished && g.AutoProgress {
		g.AutoProgress = false
		g.SuppressMapDraw = false
	}

	// PgDn/PgUp (UI編集モード以外)
	if g.InputMode == EditNone {
		if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
			g.NextStep()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
			g.UndoStep()
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		g.World2.ShowGrid = !g.World2.ShowGrid
	}

	// --- UI入力モードの開始 (マウス) ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		newMode := EditNone
		if mx >= 10 && mx <= 210 {
			if my >= 100 && my <= 130 {
				newMode = EditSoilMin
				g.InputBuffer = fmt.Sprintf("%d", g.SoilMin)
			} else if my >= 140 && my <= 170 {
				newMode = EditSoilMax
				g.InputBuffer = fmt.Sprintf("%d", g.SoilMax)
			} else if my >= 180 && my <= 210 {
				newMode = EditW2Width
				g.InputBuffer = fmt.Sprintf("%d", g.W2Width)
			} else if my >= 220 && my <= 250 {
				newMode = EditW2Height
				g.InputBuffer = fmt.Sprintf("%d", g.W2Height)
			} else if my >= 260 && my <= 290 {
				newMode = EditTransitDist
				g.InputBuffer = fmt.Sprintf("%d", g.TransitDist)
			} else if my >= 300 && my <= 330 {
				newMode = EditMapRatio
				g.InputBuffer = fmt.Sprintf("%d", g.MapRatio)
			} else if my >= 340 && my <= 370 {
				g.EnableCentering = !g.EnableCentering // Centering はトグルなのでここで処理を完結
			} else if my >= 380 && my <= 410 {
				newMode = EditCliffInit
				g.InputBuffer = fmt.Sprintf("%.1f", g.CliffInitVal)
			} else if my >= 420 && my <= 450 {
				newMode = EditCliffDec
				g.InputBuffer = fmt.Sprintf("%.2f", g.CliffDecVal)
			} else if my >= 460 && my <= 490 {
				newMode = EditShallowDec
				g.InputBuffer = fmt.Sprintf("%.2f", g.ShallowDecVal)
			} else if my >= 500 && my <= 530 {
				newMode = EditCliffPath
				g.InputBuffer = fmt.Sprintf("%d", g.CliffPathLen)
			} else if my >= 540 && my <= 570 {
				newMode = EditForceSwitch
				g.InputBuffer = fmt.Sprintf("%d", g.ForceSwitch)
			}
		}

		// トグルではない入力ボックスをクリックした時のみモードを変更
		if newMode != g.InputMode && newMode != EditNone && newMode != EditCentering {
			g.InputMode = newMode
		} else if newMode == g.InputMode {
			// 同じボックスを再度クリックしたら解除
			g.InputMode = EditNone
		} else if newMode == EditCentering {
			// Centering をクリックした場合、モードは EditNone に戻る (トグル処理は既に完了)
			g.InputMode = EditNone
		} else {
			// どこでもない場所をクリックしたら解除
			g.InputMode = EditNone
		}
	}

	// --- UI入力モード中のキー入力 (数字/小数点/BS/Enter) ---
	if g.InputMode != EditNone && g.InputMode != EditCentering {
		// Enterキーが押されたら、モードを解除し、新しい値を適用
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			valInt, errInt := strconv.Atoi(g.InputBuffer)
			valFloat, errFloat := strconv.ParseFloat(g.InputBuffer, 64)

			if errInt == nil {
				switch g.InputMode {
				case EditSoilMin:
					g.SoilMin = valInt
				case EditSoilMax:
					g.SoilMax = valInt
				case EditW2Width:
					g.W2Width = valInt
				case EditW2Height:
					g.W2Height = valInt
				case EditTransitDist:
					g.TransitDist = valInt
				case EditMapRatio:
					if valInt >= 0 && valInt <= 10 {
						g.MapRatio = valInt
					}
				case EditCliffPath:
					if valInt > 0 {
						g.CliffPathLen = valInt
					}
				case EditForceSwitch:
					g.ForceSwitch = valInt
				}
			}
			if errFloat == nil {
				switch g.InputMode {
				case EditCliffInit:
					g.CliffInitVal = valFloat
				case EditCliffDec:
					g.CliffDecVal = valFloat
				case EditShallowDec:
					g.ShallowDecVal = valFloat
				}
			}
			g.InitWorld2Generator()
			g.InputMode = EditNone
			return nil // 処理完了
		}

		for k := ebiten.Key0; k <= ebiten.Key9; k++ {
			if inpututil.IsKeyJustPressed(k) {
				g.InputBuffer += string('0' + (k - ebiten.Key0))
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
			g.InputBuffer += "."
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			if len(g.InputBuffer) > 0 {
				g.InputBuffer = g.InputBuffer[:len(g.InputBuffer)-1]
			}
		}
	}

	// --- カメラ操作 (UI入力モード外) ---
	if g.InputMode == EditNone {
		moveSpd := 10.0 / g.World2.Zoom
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			g.World2.OffsetX -= moveSpd
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			g.World2.OffsetX += moveSpd
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			g.World2.OffsetY -= moveSpd
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			g.World2.OffsetY += moveSpd
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
				dx := float64(g.MouseStartX-cx) / g.World2.Zoom
				dy := float64(g.MouseStartY-cy) / g.World2.Zoom
				g.World2.OffsetX += dx
				g.World2.OffsetY += dy
				g.MouseStartX, g.MouseStartY = cx, cy
			}
		} else {
			g.IsDragging = false
		}

		_, dy := ebiten.Wheel()
		if ebiten.IsKeyPressed(ebiten.KeyControl) && dy != 0 {
			mx, my := ebiten.CursorPosition()
			worldMx := (float64(mx)-ScreenWidth/2)/g.World2.Zoom + g.World2.OffsetX
			worldMy := (float64(my)-ScreenHeight/2)/g.World2.Zoom + g.World2.OffsetY
			if dy > 0 {
				g.World2.Zoom *= 1.1
			} else {
				g.World2.Zoom /= 1.1
			}
			g.World2.OffsetX = worldMx - (float64(mx)-ScreenWidth/2)/g.World2.Zoom
			g.World2.OffsetY = worldMy - (float64(my)-ScreenHeight/2)/g.World2.Zoom
		}
	}

	return nil
}

func (g *Game) DrawWorld2(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	w := g.World2.Width
	h := g.World2.Height

	// 自動進行中はマップ描画をスキップ
	if !g.SuppressMapDraw {
		startX := int(g.World2.OffsetX/float64(World2TileSize)) - int(float64(ScreenWidth)/g.World2.Zoom/float64(World2TileSize))/2 - 2
		startY := int(g.World2.OffsetY/float64(World2TileSize)) - int(float64(ScreenHeight)/g.World2.Zoom/float64(World2TileSize))/2 - 2
		endX := startX + int(float64(ScreenWidth)/g.World2.Zoom/float64(World2TileSize)) + 4
		endY := startY + int(float64(ScreenHeight)/g.World2.Zoom/float64(World2TileSize)) + 4

		for x := startX; x <= endX; x++ {
			for y := startY; y <= endY; y++ {
				if x < 0 || x >= w || y < 0 || y >= h {
					continue
				}
				tile := g.World2.Tiles[x][y]

				sx := (float64(x)*float64(World2TileSize)-g.World2.OffsetX)*g.World2.Zoom + ScreenWidth/2
				sy := (float64(y)*float64(World2TileSize)-g.World2.OffsetY)*g.World2.Zoom + ScreenHeight/2
				size := float64(World2TileSize) * g.World2.Zoom

				var c color.Color
				// --- タイルカラー判定 ---
				switch tile.Type {
				case W2TileSoil:
					if g.Gen2.NewSoils[y*w+x] {
						c = color.RGBA{210, 180, 140, 255}
					} else {
						switch tile.Source {
						case SrcMain:
							c = color.RGBA{180, 100, 80, 255}
						case SrcSub:
							c = color.RGBA{100, 160, 80, 255}
						case SrcMix:
							c = color.RGBA{160, 100, 160, 255}
						case SrcBridge:
							c = color.RGBA{150, 150, 160, 255}
						case SrcIsland:
							c = color.RGBA{230, 190, 100, 255}
						case SrcBRouteIsland:
							c = color.RGBA{100, 180, 100, 255} // B航路の経由島（緑）
						default:
							c = color.RGBA{139, 69, 19, 255}
						}
					}
				case W2TileVariableOcean:
					if tile.IsLake {
						c = color.RGBA{60, 100, 200, 255}
					} else {
						c = color.RGBA{30, 60, 180, 255}
					}
					// 航路の色付け
					if tile.Source == SrcTransitPath {
						c = color.RGBA{40, 80, 160, 255} // A航路（青系）
					}
					if tile.Source == SrcBRoutePath {
						c = color.RGBA{50, 120, 80, 255} // B航路（暗緑）
					}
				case W2TileFixedOcean:
					c = color.RGBA{10, 20, 80, 255}
				case W2TileTransit:
					if tile.Source == SrcBRouteIsland {
						c = color.RGBA{100, 180, 100, 255} // B航路の経由島（緑）
					} else {
						c = color.RGBA{200, 180, 80, 255} // A航路の経由島（黄色）
					}
				case W2TileCliff:
					c = color.RGBA{80, 40, 10, 255}
					if g.Gen2.NewSoils[y*w+x] {
						c = color.RGBA{120, 60, 30, 255}
					}
				case W2TileShallow:
					c = color.RGBA{60, 160, 200, 255}
					if g.Gen2.NewSoils[y*w+x] {
						c = color.RGBA{100, 200, 255, 255}
					}
				case W2TileDeepSea:
					c = color.RGBA{255, 255, 0, 255} // 深海（デバッグ用：黄色）
				case W2TileVeryDeepSea:
					c = color.RGBA{255, 128, 0, 255} // 大深海（デバッグ用：オレンジ）
				}
				ebitenutil.DrawRect(screen, sx, sy, size+1, size+1, c)
				// --- タイルカラー判定 終 ---

				if g.World2.ShowGrid || tile.Type == W2TileTransit {
					if g.World2.ShowGrid {
						ebitenutil.DrawRect(screen, sx, sy, size, 1, color.RGBA{255, 255, 255, 50})
						ebitenutil.DrawRect(screen, sx, sy, 1, size, color.RGBA{255, 255, 255, 50})
					}
					if tile.Type == W2TileTransit && size > 10 {
						text.Draw(screen, "経", basicfont.Face7x13, int(sx), int(sy+10), color.Black)
					}
				}

				// Gen Mask Imageの描画
				if g.Gen2.CurrentStep <= Phase_SoilStart && g.Gen2.FinalMask != nil {
					val := g.Gen2.FinalMask[x][y]
					if val > 0 {
						gray := uint8(val * 255)
						ebitenutil.DrawRect(screen, sx, sy, size+1, size+1, color.RGBA{gray, gray, gray, 100})
					}
				}
			}
		}

		// --- ポーズ中の強調描画 (航路探索Aでも使用) ---
		if len(g.World2.PinkRects) > 0 {
			for _, rect := range g.World2.PinkRects {
				sx := (float64(rect.x)*float64(World2TileSize)-g.World2.OffsetX)*g.World2.Zoom + ScreenWidth/2
				sy := (float64(rect.y)*float64(World2TileSize)-g.World2.OffsetY)*g.World2.Zoom + ScreenHeight/2
				w := float64(rect.w) * g.World2.Zoom
				h := float64(rect.h) * g.World2.Zoom

				// 赤の半透明の矩形を描画（航路探索A用）
				ebitenutil.DrawRect(screen, sx, sy, w, h, color.RGBA{255, 80, 80, 128})
			}
		}
		// --- 強調描画 終 ---
	} // if !g.SuppressMapDraw の閉じ括弧

	vectorY := 20
	text.Draw(screen, "Phase: "+g.Gen2.PhaseName, basicfont.Face7x13, 220, 20, color.White)
	if g.LastTargetSoil > 0 {
		text.Draw(screen, fmt.Sprintf("Target Soil: %d%%", g.LastTargetSoil), basicfont.Face7x13, 220, 40, color.White)
	}

	for _, s := range g.World2.StatsInfo {
		text.Draw(screen, s, basicfont.Face7x13, 10, vectorY, color.White)
		vectorY += 15
	}

	if g.WarningTimer > 0 {
		msgWidth := len(g.WarningMsg) * 7
		ebitenutil.DrawRect(screen, float64(ScreenWidth/2-msgWidth/2-10), float64(ScreenHeight/2-20), float64(msgWidth+20), 40, color.RGBA{200, 0, 0, 200})
		text.Draw(screen, g.WarningMsg, basicfont.Face7x13, ScreenWidth/2-msgWidth/2, ScreenHeight/2+5, color.White)
	}

	drawInputBox := func(y int, label string, val interface{}, mode int) {
		boxColor := color.RGBA{50, 50, 50, 200}
		if g.InputMode == mode {
			boxColor = color.RGBA{100, 100, 50, 200}
		}
		ebitenutil.DrawRect(screen, 10, float64(y), 200, 30, boxColor)

		var txt string
		switch v := val.(type) {
		case int:
			txt = fmt.Sprintf("%s: %d", label, v)
		case float64:
			txt = fmt.Sprintf("%s: %.2f", label, v)
		}

		if g.InputMode == mode {
			txt = fmt.Sprintf("%s: %s_", label, g.InputBuffer)
		}
		text.Draw(screen, txt, basicfont.Face7x13, 20, y+20, color.White)
	}

	drawInputBox(100, "Min Soil %", g.SoilMin, EditSoilMin)
	drawInputBox(140, "Max Soil %", g.SoilMax, EditSoilMax)
	drawInputBox(180, "Width", g.W2Width, EditW2Width)
	drawInputBox(220, "Height", g.W2Height, EditW2Height)
	drawInputBox(260, "Transit Dist", g.TransitDist, EditTransitDist)

	drawInputBox(300, "Ratio", g.MapRatio, EditMapRatio)

	cenColor := color.RGBA{50, 0, 0, 200}
	cenText := "OFF"
	if g.EnableCentering {
		cenColor = color.RGBA{0, 100, 0, 200}
		cenText = "ON"
	}
	ebitenutil.DrawRect(screen, 10, 340, 200, 30, cenColor)
	text.Draw(screen, "Centering: "+cenText, basicfont.Face7x13, 20, 360, color.White)

	drawInputBox(380, "Cliff Init", g.CliffInitVal, EditCliffInit)
	drawInputBox(420, "Cliff Dec", g.CliffDecVal, EditCliffDec)
	drawInputBox(460, "Shallow Dec", g.ShallowDecVal, EditShallowDec)
	drawInputBox(500, "Cliff Path", g.CliffPathLen, EditCliffPath)
	drawInputBox(540, "Force Turn", g.ForceSwitch, EditForceSwitch)

	text.Draw(screen, "[PgDn] Next, [PgUp] Back, [Enter] All", basicfont.Face7x13, 10, 670, color.White)
	text.Draw(screen, "[Drag]: Move, [Ctrl+Wheel]: Zoom, [R]: Reset", basicfont.Face7x13, 10, ScreenHeight-20, color.White)
}
