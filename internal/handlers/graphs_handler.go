package handlers

import (
	"net/http"

	"github.com/AleksKAG/ai-construction-manager/internal/repository"
	"github.com/gin-gonic/gin"
)

func ListGraphs(repo repository.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		graphs, err := repo.GetAllGraphs(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, graphs)
	}
}

func GetGraphForObject(repo repository.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		objectID := c.Param("object_id")
		graph, err := repo.GetGraph(c.Request.Context(), objectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Graph not found"})
			return
		}
		c.JSON(http.StatusOK, graph)
	}
}
