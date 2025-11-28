package main

import (
	"math"
	"math/rand"
	"time"
)

// Status はキャラクターのステータスを定義します
type Status struct {
	HP, SP             int
	STR, DEX, VIT, INT int
	AGI, SPD, LUCK     int
}

// Character はゲーム内のキャラクターを表します
type Character struct {
	ID        int
	Name      string
	Race      string // "Human", "Elf", "Dwarf"
	Job       string
	Stats     Status
	BaseWT    int
	Inventory Inventory
}

// Inventory はアイテム管理情報を表します
type Inventory struct {
	TotalWeight float64 // 現在の総重量 (kg)
}

// NewCharacter はランダムなステータスを持つキャラクターを生成します
func NewCharacter(id int, name, race string) *Character {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// BaseWTの決定 (仕様: 275-325, 中央値300)
	// 正規分布に近い形を簡易的に模倣（3つのサイコロを振るイメージ）
	baseWT := 275 + rng.Intn(17) + rng.Intn(17) + rng.Intn(18) // 範囲 275-326程度

	// 仮のステータス生成 (ボーナスポイント等は省略し、ランダム設定)
	c := &Character{
		ID:   id,
		Name: name,
		Race: race,
		Stats: Status{
			HP: 100, SP: 50,
			STR: 10 + rng.Intn(9), // 10-18
			DEX: 10 + rng.Intn(9),
			VIT: 10 + rng.Intn(9),
			INT: 10 + rng.Intn(9),
			AGI: 10 + rng.Intn(9),
			SPD: 10 + rng.Intn(9),
			LUCK: 10 + rng.Intn(9),
		},
		BaseWT: baseWT,
		Inventory: Inventory{
			TotalWeight: 0.0,
		},
	}
	return c
}

// MaxCapacitySlots はインベントリの最大マス数を計算します
// 仕様: (VIT * 3) + DEX
func (c *Character) MaxCapacitySlots() int {
	return (c.Stats.VIT * 3) + c.Stats.DEX
}

// MaxLoadWeight は最大積載量(kg)を計算します
// 仕様: STR * 5
func (c *Character) MaxLoadWeight() float64 {
	return float64(c.Stats.STR * 5)
}

// CalculateGameWT は現在のGameWTを計算します
// 仕様: Base WT + ceil(総重量) - floor(AGI * 2.5) - (INT * 2)
func (c *Character) CalculateGameWT() int {
	weightPenalty := int(math.Ceil(c.Inventory.TotalWeight))
	agiBonus := int(math.Floor(float64(c.Stats.AGI) * 2.5))
	intBonus := c.Stats.INT * 2

	gameWT := c.BaseWT + weightPenalty - agiBonus - intBonus
	
	// 安全のため下限を設定（仕様確認時の議論より1を下限とする）
	if gameWT < 1 {
		gameWT = 1
	}
	return gameWT
}

// TimePenaltyRate は重量過多による時間経過ペナルティ倍率を返します
func (c *Character) TimePenaltyRate() float64 {
	maxLoad := c.MaxLoadWeight()
	if maxLoad == 0 {
		return 2.0 // STR0の場合は動けないので最大のペナルティ
	}
	ratio := c.Inventory.TotalWeight / maxLoad

	if ratio < 0.85 {
		return 1.0
	} else if ratio < 0.90 {
		return 1.1
	} else if ratio < 0.95 {
		return 1.2
	} else if ratio < 1.0 {
		return 1.4
	}
	return 2.0
}