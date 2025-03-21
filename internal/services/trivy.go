package services

import (
	"scan_backend/models"
	"scan_backend/internal/utils"

	"encoding/json"
	"fmt"
	"os/exec"
)

type TrivyResult struct {
	ArtifactName string                 `json:"ArtifactName"`
	Metadata     map[string]interface{} `json:"Metadata"`
	Results      []struct {
		Target          string `json:"Target"`
		Class           string `json:"Class"`
		Type            string `json:"Type"`
		Vulnerabilities []struct {
			VulnerabilityID string `json:"VulnerabilityID"`
			PkgName         string `json:"PkgName"`
			Severity        string `json:"Severity"`
			InstalledVersion string `json:"InstalledVersion"`
			FixedVersion     string `json:"FixedVersion"`
			Status           string `json:"Status"`
			PrimaryURL       string `json:"PrimaryURL"`
			DataSource       map[string]string `json:"DataSource"`
			Title        string   `json:"Title"`
			Description  string   `json:"Description"`
			CweIDs       []string `json:"CweIDs"`
			VendorSeverity map[string]int `json:"VendorSeverity"`
			CVSS             map[string]map[string]interface{} `json:"CVSS"`
			References       []string `json:"References"`
			PublishedDate    string   `json:"PublishedDate"`
			LastModifiedDate string   `json:"LastModifiedDate"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

func RunTrivy(scanID int, options map[string]interface{}) error {
	image, ok := options["image"].(string)
	if !ok {
		return fmt.Errorf("invalid options: 'image' is required for Trivy")
	}

	cmd := exec.Command("trivy", "image", "--format", "json", image, "--quiet")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running Trivy: %v", err)
	}

	var result TrivyResult
	if err := json.Unmarshal(out, &result); err != nil {
		return fmt.Errorf("error parsing Trivy result: %v", err)
	}

	return SaveTrivyResult(scanID, &result)
}

func SaveTrivyResult(scanID int, result *TrivyResult) error {
	// ScanTargetを保存
	target := models.ScanTarget{
		ScanSummaryID:  uint(scanID),
		TargetName:     result.ArtifactName,
		TargetMetadata: utils.ToJSON(result.Metadata),
	}
	if err := db.Create(&target).Error; err != nil {
		return fmt.Errorf("failed to save scan target: %v", err)
	}

	// ScanControlを保存し、ScanTargetと関連付け
	for _, res := range result.Results {
		control := models.ScanControl{
			ScanTargetID:   target.ID,
			ControlID:      res.Type,
			ControlName:    res.Class,
			ControlStatus:  "Fail", // 状態は適宜設定
			ControlDetails: utils.ToJSON(res.Vulnerabilities),
		}

		// スキャンごとに独立したエントリを保存
		if err := db.Create(&control).Error; err != nil {
			return fmt.Errorf("failed to save scan control: %v", err)
		}
	}

	return nil
}

