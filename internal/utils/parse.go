package utils

import (

	"fmt"
	"encoding/json"
	"log"
)

// JSON文字列をGoの構造にパースするユーティリティ関数
func ParseJSONString(jsonStr string) interface{} {
	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		log.Printf("JSONパースエラー: %v\nJSON: %s", err, jsonStr)
		// エラー時のデフォルト値
		return map[string]interface{}{
			"error": fmt.Sprintf("Invalid JSON: %v", err),
		}
	}
	return result
}

func ToJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}