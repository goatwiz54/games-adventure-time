package main

import (
	"math/rand"
	"time"
)

// MapTile は1マスの情報を持ちます
type MapTile struct {
	Type   int // 0:None(壁), 1:Floor(床)
	Height int // 高さ (0-5程度)
}

// DungeonMap はマップ全体を管理します
type DungeonMap struct {
	Width, Height int
	Tiles         [][]MapTile
}

// GenerateDungeon は部屋と通路を持つマップを生成します
func GenerateDungeon(w, h int) *DungeonMap {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	d := &DungeonMap{
		Width:  w,
		Height: h,
		Tiles:  make([][]MapTile, w),
	}
	for x := 0; x < w; x++ {
		d.Tiles[x] = make([]MapTile, h)
		for y := 0; y < h; y++ {
			// 初期化：すべて壁、高さ0
			d.Tiles[x][y] = MapTile{Type: 0, Height: 0}
		}
	}

	// 簡易的な部屋生成ロジック
	numRooms := 5 + rng.Intn(5)
	var rooms []struct{ x, y, w, h, z int }

	for i := 0; i < numRooms; i++ {
		rw := 4 + rng.Intn(6)
		rh := 4 + rng.Intn(6)
		rx := 1 + rng.Intn(w-rw-2)
		ry := 1 + rng.Intn(h-rh-2)
		rz := rng.Intn(3) // 部屋ごとの基本の高さ (0-2)

		// 部屋を描画
		for x := rx; x < rx+rw; x++ {
			for y := ry; y < ry+rh; y++ {
				d.Tiles[x][y].Type = 1
				d.Tiles[x][y].Height = rz
			}
		}

		// 前の部屋と通路をつなぐ（簡易版）
		if len(rooms) > 0 {
			prev := rooms[len(rooms)-1]
			// 中心座標
			cx1, cy1 := prev.x+prev.w/2, prev.y+prev.h/2
			cx2, cy2 := rx+rw/2, ry+rh/2
			
			// 通路の高さは低い方に合わせるか、平均を取る
			pathZ := prev.z
			if rz < pathZ { pathZ = rz }

			// 横移動
			start, end := cx1, cx2
			if cx1 > cx2 { start, end = cx2, cx1 }
			for x := start; x <= end; x++ {
				d.Tiles[x][cy1].Type = 1
				if d.Tiles[x][cy1].Height < pathZ { d.Tiles[x][cy1].Height = pathZ }
			}
			// 縦移動
			start, end = cy1, cy2
			if cy1 > cy2 { start, end = cy2, cy1 }
			for y := start; y <= end; y++ {
				d.Tiles[cx2][y].Type = 1
				if d.Tiles[cx2][y].Height < pathZ { d.Tiles[cx2][y].Height = pathZ }
			}
		}
		rooms = append(rooms, struct{ x, y, w, h, z int }{rx, ry, rw, rh, rz})
	}
	
	return d
}