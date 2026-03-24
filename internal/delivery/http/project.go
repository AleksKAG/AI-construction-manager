package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/AleksKAG/ai-construction-manager/internal/domain"
	"github.com/AleksKAG/ai-construction-manager/internal/repository/postgres"
	"github.com/AleksKAG/ai-construction-manager/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	repo *postgres.ProjectRepository
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{
		repo: postgres.NewProjectRepository(db),
	}
}

func (h *ProjectHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/projects", h.createProject)
	r.GET("/projects/:id", h.getProject)
	r.POST("/projects/:id/schedule", h.buildSchedule)
	r.POST("/projects/:id/estimate", h.calculateEstimate)
}

type createProjectRequest struct {
	Name    string  `json:"name" binding:"required"`
	Address string  `json:"address"`
	Budget  float64 `json:"budget" binding:"required"`
}

func (h *ProjectHandler) createProject(c *gin.Context) {
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := &domain.Project{
		Name:    req.Name,
		Address: req.Address,
		Budget:  req.Budget,
		Status:  "planning",
	}

	if err := h.repo.Create(c.Request.Context(), project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":     project.ID,
		"name":   project.Name,
		"budget": project.Budget,
	})
}

func (h *ProjectHandler) getProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	project, err := h.repo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	c.JSON(http.StatusOK, project)
}

type buildScheduleRequest struct {
	Tasks []struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		DurationDays int    `json:"duration_days"`
		Dependencies []uint `json:"dependencies"`
	} `json:"tasks"`
}

func (h *ProjectHandler) buildSchedule(c *gin.Context) {
	var req buildScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tasks []domain.Task
	for _, t := range req.Tasks {
		tasks = append(tasks, domain.Task{
			ID:           t.ID,
			Name:         t.Name,
			DurationDays: t.DurationDays,
			Dependencies: t.Dependencies,
			Status:       "pending",
		})
	}

	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := h.repo.FindByID(c.Request.Context(), uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	startDate := time.Now()
	if project.StartDate != nil {
		startDate = *project.StartDate
	}

	scheduledTasks, err := service.BuildSchedule(tasks, startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":   scheduledTasks,
		"message": "Gantt-график построен",
	})
}

func (h *ProjectHandler) calculateEstimate(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := h.repo.FindByID(c.Request.Context(), uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	result := service.CalculateEstimate(project.Tasks)

	c.JSON(http.StatusOK, gin.H{
		"total_cost":             result.TotalCost,
		"contingency":            result.Contingency,
		"total_with_contingency": result.TotalWithContingency,
		"by_category":            result.ByCategory,
		"updated_at":             result.UpdatedAt,
	})
}
