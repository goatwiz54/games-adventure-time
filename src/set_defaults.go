// filename: world2/set_defaults.go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

// applyIntSetting は設定マップから整数値を読み込み、gのフィールドに適用する
func applyIntSetting(settings map[string]string, key string, target *int) {
	if valStr, ok := settings[key]; ok {
		if valInt, err := strconv.Atoi(valStr); err == nil {
			*target = valInt
		} else {
			fmt.Printf("Warning: Setting '%s' is not an integer: %v\n", key, err)
		}
	}
}

// applyFloatSetting は設定マップから浮動小数値を読み込み、gのフィールドに適用する
func applyFloatSetting(settings map[string]string, key string, target *float64) {
	if valStr, ok := settings[key]; ok {
		if valFloat, err := strconv.ParseFloat(valStr, 64); err == nil {
			*target = valFloat
		} else {
			fmt.Printf("Warning: Setting '%s' is not a float: %v\n", key, err)
		}
	}
}

// applyBoolSetting は設定マップから真偽値を読み込み、gのフィールドに適用する
func applyBoolSetting(settings map[string]string, key string, target *bool) {
	if valStr, ok := settings[key]; ok {
		valStr = strings.ToLower(valStr)
		if valStr == "true" || valStr == "on" || valStr == "1" {
			*target = true
		} else if valStr == "false" || valStr == "off" || valStr == "0" {
			*target = false
		} else {
			fmt.Printf("Warning: Setting '%s' is not a boolean: %s\n", key, valStr)
		}
	}
}

// ApplySettings は読み込んだ設定値をGame構造体フィールドに適用する
func (g *Game) ApplySettings(settings map[string]string) {
	// Int Settings
	applyIntSetting(settings, "SoilMin", &g.SoilMin)
	applyIntSetting(settings, "SoilMax", &g.SoilMax)
	applyIntSetting(settings, "W2Width", &g.W2Width)
	applyIntSetting(settings, "W2Height", &g.W2Height)
	applyIntSetting(settings, "TransitDist", &g.TransitDist)
	applyIntSetting(settings, "VastOceanSize", &g.VastOceanSize)
	applyIntSetting(settings, "IslandBoundSize", &g.IslandBoundSize)
	applyIntSetting(settings, "MapRatio", &g.MapRatio)
	applyIntSetting(settings, "CliffPathLen", &g.CliffPathLen)
	applyIntSetting(settings, "ForceSwitch", &g.ForceSwitch)

	// Float Settings
	applyFloatSetting(settings, "CliffInitVal", &g.CliffInitVal)
	applyFloatSetting(settings, "CliffDec", &g.CliffDecVal)
	applyFloatSetting(settings, "ShallowDec", &g.ShallowDecVal)

	// Bool Settings
	applyBoolSetting(settings, "Centering", &g.EnableCentering)
	
	fmt.Println("Settings applied from file. SoilMin:", g.SoilMin)
}