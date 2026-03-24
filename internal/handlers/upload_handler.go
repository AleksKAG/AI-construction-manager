package handlers

import (
	"fmt"
	
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AleksKAG/ai-construction-manager/internal/repository"
	"github.com/gin-gonic/gin"
	
	
)

func UploadFile(repo repository.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		tempDir := os.TempDir()
		filePath := filepath.Join(tempDir, file.Filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer os.Remove(filePath)

		ext := strings.ToLower(filepath.Ext(file.Filename))
		var parsedData interface{}
		switch ext {
		case ".xlsx":
			parsedData, err = parseExcel(filePath)
		case ".docx":
			parsedData, err = parseDocx(filePath)
		default:
			err = fmt.Errorf("unsupported file type: %s", ext)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Сохранить parsed в DB (пример: обновить объект)
		// repo.UpdateParsedData(...)

		c.JSON(http.StatusOK, gin.H{"parsed_data": parsedData})
	}
}

// parseExcel и parseDocx как в предыдущем примере
func parseExcel(filePath string) (map[string]interface{}, error) {
	// ... (тот же код, что в вашем исходном)
	return nil, nil // Заглушка
}

func parseDocx(filePath string) (map[string]string, error) {
	// ... (тот же код)
	return nil, nil // Заглушка
}

func isTitle(s string) bool {
	return strings.ToUpper(s) == s || len(s) < 20
}
