package handlers

import (
	"net/http"
	"strconv"

	"github.com/AleksKAG/ai-construction-manager/internal/models"
	"github.com/AleksKAG/ai-construction-manager/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ListObjects(repo repository.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		objs, err := repo.GetAllObjects(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, objs)
	}
}

func CreateObject(repo repository.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var obj models.ProjectObject
		if err := c.BindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		obj.ID = uuid.New().String()
		if err := repo.CreateObject(c.Request.Context(), &obj); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, obj)
	}
}

func GetObject(repo repository.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		obj, err := repo.GetObject(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
			return
		}
		c.JSON(http.StatusOK, obj)
	}
}

func UpdateObject(repo repository.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var obj models.ProjectObject
		if err := c.BindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		obj.ID = id
		if err := repo.UpdateObject(c.Request.Context(), &obj); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, obj)
	}
}

func DeleteObject(repo repository.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := repo.DeleteObject(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Object deleted"})
	}
}
