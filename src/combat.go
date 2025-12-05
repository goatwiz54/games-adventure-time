// filename: combat.go
package main

// 敵のターン処理 (AI)
func (g *Game) ProcessEnemyTurn() {
	leader := g.Party.Leader
	px, py := leader.TargetX, leader.TargetY
	
	for _, e := range g.Dungeon.Enemies {
		if !e.Active { continue }
		
		for s := 0; s < e.Speed; s++ {
			// 視界判定
			vis := e.CheckPlayerVisibility(px, py)
			dist := abs(px-e.TargetX) + abs(py-e.TargetY)
			
			// AIロジック
			action := 2 // Random
			r := g.Rng.Intn(100)

			if vis == 0 { // Clear (見えている)
				if dist <= 2 {
					// 至近距離: 追跡55, 停止30, ランダム15
					if r < 55 { action = 0 } else if r < 85 { action = 1 } else { action = 2 }
					// 殺意補正: ランダムなら50%で再抽選
					if action == 2 && g.Rng.Intn(2) == 0 {
						r2 := g.Rng.Intn(100)
						if r2 < 55 { action = 0 } else if r2 < 85 { action = 1 }
					}
				} else {
					// 遠距離: 追跡30, ランダム70
					if r < 30 { action = 0 } else { action = 2 }
				}
			} else if vis == 1 { // Dim (気配)
				// 追跡30, 停止15, ランダム55
				if r < 30 { action = 0 } else if r < 45 { action = 1 } else { action = 2 }
			} else { // Dark
				action = 2
			}

			tx, ty := e.TargetX, e.TargetY
			if action == 0 { // Chase
				if px > tx { tx++ } else if px < tx { tx-- }
				if py > ty { ty++ } else if py < ty { ty-- }
				// 軸合わせ
				if abs(px-e.TargetX) > abs(py-e.TargetY) { ty = e.TargetY } else { tx = e.TargetX }
			} else if action == 2 { // Random
				dir := g.Rng.Intn(4)
				switch dir {
				case 0: ty--
				case 1: ty++
				case 2: tx--
				case 3: tx++
				}
			}

			// 移動実行
			if tx >= 0 && tx < g.Dungeon.Width && ty >= 0 && ty < g.Dungeon.Height {
				nt := g.Dungeon.Tiles[tx][ty]
				ct := g.Dungeon.Tiles[e.TargetX][e.TargetY]
				
				// 段差チェック
				if nt.Type == 1 && abs(nt.Height-ct.Height) <= 1 {
					// 向き更新
					if tx > e.TargetX { e.Facing = DirEast }
					if tx < e.TargetX { e.Facing = DirWest }
					if ty > e.TargetY { e.Facing = DirSouth }
					if ty < e.TargetY { e.Facing = DirNorth }
					
					e.TargetX = tx
					e.TargetY = ty
					e.TargetZ = nt.Height
					e.IsMoving = true
				}
			}
			
			// 接触チェック
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