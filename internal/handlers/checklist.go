package handlers

import (
	"net/http"
	"os"

	"k41m_backend/internal/services"

	"github.com/gin-gonic/gin"
)

func GetChecklistItemsHandler(c *gin.Context) {
	// クエリパラメータからフェーズを取得
	phase := c.DefaultQuery("phase", "")

	// サービス層でデータを取得
	items, err := services.GetChecklistItems(phase)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve checklist items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"checklist_items": items})
}

func UploadChecklistHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	tempFilePath := "./temp_upload_" + file.Filename
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer os.Remove(tempFilePath)

	message, err := services.ProcessChecklistFile(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}