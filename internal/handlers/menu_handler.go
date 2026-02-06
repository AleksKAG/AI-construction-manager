package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MenuHandler(c *gin.Context) {
	menu := map[string]string{
		"objects":  "/api/v1/objects - List and manage project objects",
		"graphs":   "/api/v1/graphs - View Gantt graphs for projects",
		"upload":   "/api/v1/upload - Upload Excel/DOCX for parsing",
		"estimate": "/api/v1/estimate/:id - Get cost estimates for object",
	}
	c.JSON(http.StatusOK, menu)
}
