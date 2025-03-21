package services

import (
	"fmt"
)

func RunDeployScan(scanID int, tool string, options map[string]interface{}) error {
	switch tool {
	case "Kubescape":
		return RunKubescape(scanID, options)
	default:
		return fmt.Errorf("unsupported tool for Deploy phase: %s", tool)
	}
}
