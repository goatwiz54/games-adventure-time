// filename: world2/settings.go
package main

import (
	"bufio"
	"os" // 追加
	"strings"
)

// LoadSettings は settings.txt からキーと値を読み込み、マップとして返す
func LoadSettings(filename string) (map[string]string, error) {
	settings := make(map[string]string)
	
	file, err := os.Open(filename)
	if err != nil {
		// ファイルが存在しない場合はエラーとしない (デフォルト値を使用する)
		if os.IsNotExist(err) {
			return settings, nil 
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		
		// コメント行または空行をスキップ
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// Key: Value の形式で分割
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			settings[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return settings, nil
}