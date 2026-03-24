package service

import (
	"ai-construction-manager/internal/domain"
	"time"
)

type EstimateResult struct {
	TotalCost       float64
	ByCategory      map[string]float64
	Contingency     float64 // 10% непредвиденные расходы
	TotalWithContingency float64
	UpdatedAt       time.Time
}

// CalculateEstimate рассчитывает смету по бизнес-правилам
func CalculateEstimate(tasks []domain.Task) EstimateResult {
	byCategory := make(map[string]float64)
	total := 0.0

	for _, task := range tasks {
		for _, res := range task.Resources {
			cost := res.Quantity * res.UnitCost
			byCategory[res.Type] += cost
			total += cost
		}
	}

	contingency := total * 0.10 // 10% на непредвиденные расходы
	totalWithContingency := total + contingency

	return EstimateResult{
		TotalCost:       total,
		ByCategory:      byCategory,
		Contingency:     contingency,
		TotalWithContingency: totalWithContingency,
		UpdatedAt:       time.Now(),
	}
}

// AdjustForInflation корректирует смету с учётом инфляции (бизнес-правило)
func AdjustForInflation(baseCost float64, monthsFromNow int) float64 {
	// Пример: 8% годовая инфляция в стройке
	monthlyRate := 0.08 / 12
	return baseCost * (1 + monthlyRate*float64(monthsFromNow))
}
