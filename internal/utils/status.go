package utils

func CalculateChecklistItemStatus(controls []map[string]interface{}) string {
	hasFailed := false
	hasSkipped := false
	allPassed := true

	for _, control := range controls {
		status := control["status"].(string)
		if status == "failed" {
			hasFailed = true
			allPassed = false
		} else if status == "skipped" {
			hasSkipped = true
			allPassed = false
		} else if status == "passed" {
			// Do nothing, it's already considered passed
		}
	}

	if hasFailed {
		return "failed"
	}
	if hasSkipped && !hasFailed {
		return "skipped"
	}
	if allPassed {
		return "passed"
	}
	return "unknown" // デフォルトのケース
}
