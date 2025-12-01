// filename: monitor.go
package main

import (
	"fmt"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func DrawMemoryStats(screen *ebiten.Image) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// bToMb converts bytes to megabytes
	bToMb := func(b uint64) uint64 {
		return b / 1024 / 1024
	}

	// Alloc: 現在ヒープに割り当てられているオブジェクトのバイト数
	// TotalAlloc: これまでに割り当てられたヒープオブジェクトの累積バイト数
	// Sys: OSから取得したメモリの合計バイト数
	// NumGC: 完了したGCサイクル数

	alloc := bToMb(m.Alloc)
	total := bToMb(m.TotalAlloc)
	sys := bToMb(m.Sys)
	
	// アプリが確保したメモリに対する使用率
	usagePercent := 0.0
	if sys > 0 {
		usagePercent = float64(m.Alloc) / float64(m.Sys) * 100
	}

	msg := fmt.Sprintf("FPS: %0.2f\nTPS: %0.2f\n\n[Memory]\nCurrent: %v MB\nTotal: %v MB\nSystem: %v MB\nUsage: %.2f%%", 
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		alloc, total, sys, usagePercent,
	)

	// 右上に表示
	ebitenutil.DebugPrintAt(screen, msg, ScreenWidth-150, 10)
}