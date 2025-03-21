// デプロイ後(運用・監視)フェーズは現状スキャンは実装せず、ツール(falco)からの通知を受け取って処理を行う
package services

import (
	"fmt"
)

func RunPostDeployScan(scanID int, tool string, options map[string]interface{}) error {
	// デプロイ後(運用・監視)フェーズの処理は未実装
	return fmt.Errorf("Post-Deploy phase is not yet implemented")
}