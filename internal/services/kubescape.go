package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"scan_backend/models"
	"scan_backend/internal/utils"
)

const kubescapeBaseURL = "http://kubescape.kubescape.svc.cluster.local:8080"

// デフォルトのスキャン設定
var defaultScanBody = map[string]interface{}{
	"excludedNamespaces":   []string{"kube-system"},
	"failThreshold":        42,
	"complianceThreshold":  42,
	"format":               "json",
	"hostScanner":          false,
	"includeNamespaces":    []string{"default"},
	"keepLocal":            true,
	"submit":               false,
	"targetNames":          []string{"all"},
	"targetType":           "framework",
	"useCachedArtifacts":   false,
}


func RunKubescape(scanID int, options map[string]interface{}) error {
	// スキャン初期化リクエストを作成
	scanBody := utils.MergeOptions(defaultScanBody, options)
	bodyBytes, err := json.Marshal(scanBody)
	if err != nil {
		return fmt.Errorf("failed to marshal scan body: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/scan?wait=false&keep=false", kubescapeBaseURL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create kubescape scan request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send kubescape scan request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("kubescape scan failed: %s", body)
	}

	var initResponse struct {
		ID       string `json:"id"`
		Response string `json:"response"`
		Type     string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&initResponse); err != nil {
		return fmt.Errorf("failed to parse kubescape scan response: %v", err)
	}

	// スキャンIDをDBに保存
	if err := db.Model(&models.ScanSummary{}).Where("id = ?", scanID).Update("tool_scan_id", initResponse.ID).Error; err != nil {
		return fmt.Errorf("failed to update scan_summary with tool_scan_id: %v", err)
	}

	// スキャン結果をポーリング
	return pollKubescapeResults(scanID, initResponse.ID)
}

func pollKubescapeResults(scanID int, toolScanID string) error {
	for {
		time.Sleep(10 * time.Second) // 10秒ごとにポーリング

		resp, err := http.Get(fmt.Sprintf("%s/v1/results?id=%s&keep=false", kubescapeBaseURL, toolScanID))
		if err != nil {
			return fmt.Errorf("failed to poll kubescape results: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("kubescape results poll failed: %s", body)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to parse kubescape result: %v", err)
		}

		// 結果がまだ取得できていない場合
		if result["type"] == "busy" {
			continue
		}

		// 結果をDBに保存してポーリング終了
		return saveKubescapeResults(scanID, result)
	}
}

func saveKubescapeResults(scanID int, result map[string]interface{}) error {
	// ScanTargetを保存
	target := models.ScanTarget{
		ScanSummaryID:  uint(scanID),
		TargetName:     "Kubernetes Cluster",
		TargetMetadata: utils.ToJSON(result["response"].(map[string]interface{})["metadata"]),
	}
	if err := db.Create(&target).Error; err != nil {
		return fmt.Errorf("failed to save kubescape scan target: %v", err)
	}

	// コントロール情報を保存
	controls := parseKubescapeControls(result)
	for _, control := range controls {
		control.ScanTargetID = target.ID
		if err := db.Create(&control).Error; err != nil {
			return fmt.Errorf("failed to save kubescape scan control: %v", err)
		}
	}

	return nil
}

// TODO: リファクタリング
func parseKubescapeControls(result map[string]interface{}) []models.ScanControl {
	var controls []models.ScanControl

	// summaryDetails.controls を取得
	summaryDetails, ok := result["response"].(map[string]interface{})["summaryDetails"].(map[string]interface{})
	if !ok {
		return controls
	}

	rawControls, ok := summaryDetails["controls"].(map[string]interface{})
	if !ok {
		return controls
	}

	// control_details に入れる情報は、response.summaryDetails.controls を元としている
	// kubescapeの出力では仕様?で、response.summaryDetails.controls.resourceIDs が空になってしまっているため
	// この情報を補完するためには、response.results からresourceID と controlID の対応付けをする必要がある

	// controlID(e.g. C-0185) をキーとして、そのコントロールに関連付けられた resourceID のリストを値として保持するマップを作成
	resourceControlMap := map[string][]map[string]string{}
	if rawResults, ok := result["response"].(map[string]interface{})["results"].([]interface{}); ok {
		// response.results から resourceID を取得
		for _, rawResult := range rawResults {
			resultMap, ok := rawResult.(map[string]interface{})
			if !ok {
				continue
			}

			resourceID := fmt.Sprintf("%v", resultMap["resourceID"])
			// response.resultsの resource に関連付けられた controls を取得
			if rawResultControls, ok := resultMap["controls"].([]interface{}); ok {
				for _, rawControl := range rawResultControls {
					controlMap, ok := rawControl.(map[string]interface{})
					if !ok {
						continue
					}

					controlID := fmt.Sprintf("%v", controlMap["controlID"])
					status := fmt.Sprintf("%v", controlMap["status"].(map[string]interface{})["status"])

					// resourceIDとstatusを対応付け
					resourceControlMap[controlID] = append(resourceControlMap[controlID], map[string]string{
						"resourceID": resourceID,
						"status":     status,
					})
				}
			}
		}
	}

	// controls をパース
	for controlID, controlData := range rawControls {
		controlMap, ok := controlData.(map[string]interface{})
		if !ok {
			continue
		}
		fmt.Println(controlID,resourceControlMap[controlID])

		// resourceIDs に対応するリソースデータを挿入
		resourceData := resourceControlMap[controlID]

		// ControlDetailsを構築
		controlDetails := map[string]interface{}{
			"resourceIDs":        resourceData,
			"ResourceCounters":   controlMap["ResourceCounters"],
			"subStatusCounters":  controlMap["subStatusCounters"],
			"score":              controlMap["score"],
			"complianceScore":    controlMap["complianceScore"],
			"category":           controlMap["category"],
		}

		// ScanControlデータを追加
		controls = append(controls, models.ScanControl{
			ControlID:      fmt.Sprintf("%v", controlMap["controlID"]),
			ControlName:    fmt.Sprintf("%v", controlMap["name"]),
			ControlStatus:  fmt.Sprintf("%v", controlMap["status"]),
			ControlDetails: utils.ToJSON(controlDetails),
		})
	}

	return controls
}
