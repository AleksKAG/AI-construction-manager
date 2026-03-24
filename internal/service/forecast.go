package service

import (
	"context"
	"time"

	"github.com/AleksKAG/ai-construction-manager/internal/domain"
	"github.com/AleksKAG/ai-construction-manager/pkg/gonum"
)

type ForecastResult struct {
	EstimatedCompletionDate time.Time
	DelayRisk               string  // low, medium, high
	DaysBehindSchedule      int
	Confidence              float64 // 0..1
}

// ForecastCompletion — основной метод прогнозирования
func ForecastCompletion(project domain.Project, actualProgress []struct {
	Day      int
	Progress float64
}) ForecastResult {

	var days, progress []float64
	for _, p := range actualProgress {
		days = append(days, float64(p.Day))
		progress = append(progress, p.Progress)
	}

	lr := &gonum.LinearRegression{}
	lr.Fit(days, progress)

	daysToComplete := lr.DaysToCompletion()

	risk := "low"
	if daysToComplete > float64(project.DurationDays)*1.3 {
		risk = "high"
	} else if daysToComplete > float64(project.DurationDays)*1.15 {
		risk = "medium"
	}

	startDate := time.Now()
	if project.StartDate != nil {
		startDate = *project.StartDate
	}
	completionDate := startDate.AddDate(0, 0, int(daysToComplete))

	return ForecastResult{
		EstimatedCompletionDate: completionDate,
		DelayRisk:               risk,
		DaysBehindSchedule:      int(daysToComplete) - project.DurationDays,
		Confidence:              calculateConfidence(len(actualProgress)),
	}
}

func calculateConfidence(samples int) float64 {
	if samples < 3 {
		return 0.3
	} else if samples < 7 {
		return 0.6
	}
	return 0.85
}