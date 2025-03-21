package services

import (
	"fmt"
)

func RunBuildScan(scanID int, tool string, options map[string]interface{}) error {
	switch tool {
	case "Trivy":
		return RunTrivy(scanID, options)
	default:
		return fmt.Errorf("unsupported tool for Build phase: %s", tool)
	}
}