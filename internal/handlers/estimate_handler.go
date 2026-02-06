package handlers

import (
	"net/http"

	"github.com/AleksKAG/ai-construction-manager/internal/pkg/gonum"
	"github.com/AleksKAG/ai-construction-manager/internal/repository"
	"github.com/gin-gonic/gin"
)

func GetEstimate(repo repository.ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		obj, err := repo.GetObject(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
			return
		}

		// ML adjustment с Gonum
		lr := gonum.LinearRegression{}
		// Пример данных для fit (из DB или hardcoded)
		days := []float64{1, 2, 3} // Пример
		progress := []float64{0.1, 0.3, 0.5}
		lr.Fit(days, progress)

		adjustedCosts := make(map[string]float64)
		for k, v := range obj.CostEstimates {
			adjustedCosts[k] = v * (1 + lr.Predict(10)) // Пример adjustment
		}
		adjustedCosts["AdjustedTotal"] = adjustedCosts["TotalCost"]
		adjustedCosts["DaysToCompletion"] = lr.DaysToCompletion()

		c.JSON(http.StatusOK, adjustedCosts)
	}
}
