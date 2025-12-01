// filename: app_init.go
package main

import (
	"math/rand"
	"time"
)

// initializeNewGame は Game 構造体の初期値を設定する
func initializeNewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	return &Game{ 
		State: StateMenu, 
		MenuIndex: 0, 
		Rng: rng,
		
		SoilMin: 20,
		SoilMax: 28,
		W2Width: DefaultWorld2Width,
		W2Height: DefaultWorld2Height,
		TransitDist: 15,
		VastOceanSize: 25,
		IslandBoundSize: 15,
		MapRatio:    10,
		EnableCentering: true,

		CliffInitVal:  10.0,
		CliffDecVal:   0.1,
		ShallowDecVal: 0.25,
		CliffPathLen:  5, 
		ForceSwitch:   5, 
	}
}